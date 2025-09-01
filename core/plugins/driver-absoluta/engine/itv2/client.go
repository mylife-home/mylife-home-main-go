package itv2

import (
	"context"
	"fmt"
	"mylife-home-common/log"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/http"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
	"reflect"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
)

var logger = log.CreateLogger("mylife:home:core:plugins:absoluta:engine:itv2")

type ConnectionStatus uint8

const (
	ConnectionClosed      ConnectionStatus = 0
	ConnectionHandshaking ConnectionStatus = 1
	ConnectionOpen        ConnectionStatus = 2
)

func (status ConnectionStatus) String() string {
	switch status {
	case ConnectionClosed:
		return "Closed"
	case ConnectionHandshaking:
		return "Handshaking"
	case ConnectionOpen:
		return "Open"
	default:
		return fmt.Sprintf("<%d>", status)
	}
}

const heartbeatInterval = time.Second * 5

type Client struct {
	servAddr string
	uid      string
	pin      string

	status                    ConnectionStatus
	connectionStatusCallbacks []func(ConnectionStatus)
	notificationsCallbacks    []func(commands.Command)
	mux                       sync.Mutex // Synchronize external callback/status manipulations

	ctx        context.Context
	close      context.CancelFunc
	workerSync sync.WaitGroup

	conn         *connection
	transactions *transactionManager
}

func MakeClient(servAddr string, uid string, pin string) *Client {
	ctx, close := context.WithCancel(context.Background())

	client := &Client{
		servAddr:                  servAddr,
		uid:                       uid,
		pin:                       pin,
		ctx:                       ctx,
		close:                     close,
		status:                    ConnectionClosed,
		connectionStatusCallbacks: make([]func(ConnectionStatus), 0),
		notificationsCallbacks:    make([]func(commands.Command), 0),
	}

	client.workerSync.Add(1)

	go client.worker()

	return client
}

func (client *Client) Close() {
	client.close()
	client.workerSync.Wait()
}

func (client *Client) changeStatus(newStatus ConnectionStatus) {
	client.mux.Lock()
	defer client.mux.Unlock()

	if client.status == newStatus {
		return
	}

	client.status = newStatus

	for _, callback := range client.connectionStatusCallbacks {
		go callback(newStatus)
	}
}

func (client *Client) Status() ConnectionStatus {
	client.mux.Lock()
	defer client.mux.Unlock()

	return client.status
}

func (client *Client) RegisterStatusChange(callback func(ConnectionStatus)) {
	client.mux.Lock()
	defer client.mux.Unlock()

	client.connectionStatusCallbacks = append(client.connectionStatusCallbacks, callback)
}

func (client *Client) RegisterNotifications(callback func(commands.Command)) {
	client.mux.Lock()
	defer client.mux.Unlock()

	client.notificationsCallbacks = append(client.notificationsCallbacks, callback)
}

func (client *Client) worker() {
	defer client.workerSync.Done()

	for {
		if http.CheckAvailability(client.ctx, client.uid) {
			client.connection()
		}

		select {
		case <-client.ctx.Done():
			return
		case <-time.After(time.Second * 5):
			// reconnect
		}
	}
}

func (client *Client) connection() {
	logger.Debugf("Dial to '%s'", client.servAddr)

	conn, err := makeConnection(client.ctx, client.servAddr)
	if err != nil {
		logger.WithError(err).Errorf("Could not connect to '%s'", client.servAddr)
		return
	}

	client.conn = conn
	defer func() {
		client.conn.Close()
		client.conn = nil
		client.changeStatus(ConnectionClosed)
		logger.Debug("Connection closed")
	}()

	client.changeStatus(ConnectionHandshaking)
	logger.Debug("Start handshake")

	if err := handshake(client.ctx, client.conn, client.pin); err != nil {
		logger.WithError(err).Error("Handshake failed")
		return
	}

	client.changeStatus(ConnectionOpen)
	logger.Debug("Handshake done")

	client.transactions = newTransactionManager()
	defer func() {
		client.transactions.CancelAll()
		client.transactions = nil
	}()

	for {
		select {
		case <-client.ctx.Done():
			return

		case <-time.After(heartbeatInterval):
			go client.heartbeat()

		case err := <-client.conn.Errors():
			logger.WithError(err).Error("Error on connection")
			return

		case cmd := <-client.conn.Read():
			client.processCommand(cmd)
		}
	}
}

func (client *Client) processCommand(cmd commands.Command) {
	if client.transactions.ProcessCommand(cmd) {
		return
	}

	client.mux.Lock()
	defer client.mux.Unlock()

	// Not matched by transaction manager, consider it notification
	for _, callback := range client.notificationsCallbacks {
		go callback(cmd)
	}
}

// send command and expect basic response
func (client *Client) executeCommand(cmd commands.CommandWithAppSeq) error {
	_, err := client.execCmdInternal(cmd)
	return err
}

// send request and expect response
func (client *Client) executeRequest(req commands.RequestData) (commands.ResponseData, error) {
	cmd := &commands.Request{
		ReqCode: req.RequestCode(),
		ReqData: req,
	}

	res, err := client.execCmdInternal(cmd)

	var resData commands.ResponseData
	if err == nil {
		resData = res.(commands.ResponseData)
	}

	return resData, err
}

func (client *Client) execCmdInternal(cmd commands.CommandWithAppSeq) (commands.Command, error) {
	// Note: can be set to nil in the middle
	conn := client.conn
	transactions := client.transactions

	if conn == nil || transactions == nil {
		return nil, fmt.Errorf("not connected")
	}

	cmd.SetAppSeq(conn.NextAppSeq())

	conn.Write(cmd)

	transaction := makeTransaction(cmd)
	transactions.addTransaction(transaction)
	res, err := transaction.Wait()
	transactions.removeTransaction(transaction)

	return res, err
}

func (client *Client) heartbeat() {
	cmd := &commands.UserActivity{
		PartitionNumber: &serialization.VarBytes{},
		Type:            4,
	}

	if err := client.executeCommand(cmd); err != nil {
		logger.WithError(err).Errorf("Heartbeat error")
	}

	logger.Debugf("Heartbeat OK")
}

func (client *Client) getLabel(optionId uint8, offset int, index int) (string, error) {
	req := &commands.ConfigurationRequest{
		OptionId:           &serialization.VarBytes{},
		OptionIdOffsetFrom: &serialization.VarBytes{},
		OptionIdOffsetTo:   &serialization.VarBytes{},
	}

	req.OptionId.SetUint(uint64(optionId))

	req.OptionIdOffsetFrom.SetUint(uint64(index + offset))
	req.OptionIdOffsetTo.SetUint(uint64(index + offset))

	cmd, err := client.executeRequest(req)

	if err != nil {
		return "", err
	}

	config, ok := cmd.(*commands.Configuration)
	if !ok {
		return "", fmt.Errorf("invalid response type %s", reflect.TypeOf(cmd))
	}

	strs := config.GetStrings(charmap.Windows1252)
	if len(strs) == 0 {
		return "", fmt.Errorf("got response without value")
	}

	return strings.TrimSpace(strs[0]), nil

}

func (client *Client) GetZoneLabel(zoneIndex int) (string, error) {
	return client.getLabel(commands.ConfigurationOptionAbsolutaZoneLabel, 0, zoneIndex)
}

func (client *Client) GetPartitionLabel(partitionIndex int) (string, error) {
	return client.getLabel(commands.ConfigurationOptionAbsolutaPartitionLabel, 1, partitionIndex)
}

func (client *Client) GetZonesAssignment(partitionIndex int) ([]int, error) {
	req := &commands.ZoneAssignmentConfigurationRequest{
		PartitionNumber: &serialization.VarBytes{},
	}

	req.PartitionNumber.SetUint(uint64(partitionIndex))

	cmd, err := client.executeRequest(req)
	if err != nil {
		return nil, err
	}

	res, ok := cmd.(*commands.ZoneAssignmentConfiguration)
	if !ok {
		return nil, fmt.Errorf("bad response type: %s", reflect.TypeOf(cmd))
	}

	return res.GetAssignedZones(), nil
}

func (client *Client) GetPartitionStatus() (*commands.PartitionStatus, error) {
	req := &commands.PartitionStatusRequest{
		Partitions: &serialization.VarBytes{},
	}

	// Note : ABS 42 only output partition bitmask with one bit set over 16: [0x01 0x00]
	req.Partitions.SetUint(0x100)

	cmd, err := client.executeRequest(req)
	if err != nil {
		return nil, err
	}

	res, ok := cmd.(*commands.PartitionStatus)
	if !ok {
		return nil, fmt.Errorf("bad response type: %s", reflect.TypeOf(cmd))
	}

	return res, nil
}

func (client *Client) GetZoneStatuses() (*commands.ZoneStatus, error) {
	req := &commands.ZoneStatusRequest{
		ZoneNumber:    &serialization.VarBytes{},
		NumberOfZones: &serialization.VarBytes{},
	}

	// Note : ABS 42 only output this (regardless of request params)
	req.ZoneNumber.SetUint(1)
	req.NumberOfZones.SetUint(42)

	cmd, err := client.executeRequest(req)
	if err != nil {
		return nil, err
	}

	res, ok := cmd.(*commands.ZoneStatus)
	if !ok {
		return nil, fmt.Errorf("bad response type: %s", reflect.TypeOf(cmd))
	}

	return res, nil
}
