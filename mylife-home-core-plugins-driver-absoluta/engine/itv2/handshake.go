package itv2

import (
	"context"
	"errors"
	"fmt"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/serialization"
	"reflect"
	"strings"
)

type handshakeData struct {
	ctx  context.Context
	conn *connection
	pin  string
}

func handshake(ctx context.Context, conn *connection, pin string) error {
	handshake := &handshakeData{
		ctx:  ctx,
		conn: conn,
		pin:  pin,
	}

	return handshake.execute()
}

func (handshake *handshakeData) execute() error {
	if err := handshake.handshakeOpenSession(); err != nil {
		return err
	}

	if err := handshake.handshakeRequestAccess(); err != nil {
		return err
	}

	if err := handshake.handshakeSoftwareVersion(); err != nil {
		return err
	}

	if err := handshake.handshakeEnterLevelAccess(); err != nil {
		return err
	}

	return nil
}

func (handshake *handshakeData) handshakeOpenSession() error {

	req := &commands.OpenSession{
		DeviceTypeOrVendorID: 143,
		DeviceId:             0,
		//SoftwareVersion
		//ProtocolVersion
		TxSize:         50,
		RxSize:         1024,
		Unused:         1,
		EncryptionType: 0,
	}

	serialization.BCDEncode(req.SoftwareVersion[:], "0100")
	serialization.BCDEncode(req.ProtocolVersion[:], "0203")

	if err := handshake.sendCommandWithResponse(req); err != nil {
		return err
	}

	// Now server should send its opensession

	res, err := handshake.receive()
	if err != nil {
		return err
	}

	serverReq, ok := res.(*commands.OpenSession)
	if !ok {
		return fmt.Errorf("open session unexpected server request: %+v", res)
	}

	// TODO: check server data

	if serverReq.EncryptionType != 0 {
		return fmt.Errorf("server required encryption, not supported")
	}

	handshake.sendResponse(serverReq.AppSeq, commands.ResponseCodeSuccess)

	return nil
}

func (handshake *handshakeData) handshakeRequestAccess() error {

	req := &commands.RequestAccess{
		Identifier: &serialization.VarBytes{},
	}

	identifier := make([]byte, 4)
	serialization.BCDEncode(identifier, "00000000")
	req.Identifier.Set(identifier)

	if err := handshake.sendCommandWithResponse(req); err != nil {
		return err
	}

	// Now server should send its request access

	res, err := handshake.receive()
	if err != nil {
		return err
	}

	serverReq, ok := res.(*commands.RequestAccess)
	if !ok {
		return fmt.Errorf("request access unexpected server request: %+v", res)
	}

	logger.Debugf("Server identifier: %+v", serverReq.Identifier)

	handshake.sendResponse(serverReq.AppSeq, commands.ResponseCodeSuccess)

	return nil
}

func (handshake *handshakeData) handshakeSoftwareVersion() error {

	req := &commands.SoftwareVersion{
		//VersionFields
	}

	serialization.BCDEncode(req.VersionFields[:], strings.Replace("35 00 00 1E 02 03 00 00 01 03 01", " ", "", -1))

	handshake.send(req)

	// Now server should send its software version

	res, err := handshake.receive()
	if err != nil {
		return err
	}

	serverReq, ok := res.(*commands.SoftwareVersion)
	if !ok {
		return fmt.Errorf("software version unexpected server request: %+v", res)
	}

	logger.Debugf("Server version: %+v", serverReq.VersionFields)

	return nil
}

func (handshake *handshakeData) handshakeEnterLevelAccess() error {
	// Not part of the Java handshakeData, but needed to go further.

	req := &commands.EnterAccessLevel{
		PartitionNumber:       &serialization.VarBytes{},
		Type:                  commands.UserAccessLevel,
		ProgrammingAccessCode: &serialization.VarBytes{},
	}

	// StringUtils.leftPad(paramString, 6, 'A');
	pin := handshake.pin

	if len(pin) > 6 {
		pin = pin[:6]
	} else if len(pin) < 6 {
		pin = strings.Repeat("A", 6-len(pin)) + pin
	}

	accessCode := make([]byte, 3)
	serialization.BCDEncode(accessCode, pin)
	req.ProgrammingAccessCode.Set(accessCode)

	return handshake.sendCommandWithResponse(req)
}

// TODO: pass req instead of serverAppSeq
func (handshake *handshakeData) sendResponse(serverAppSeq byte, code commands.ResponseCode) {
	response := &commands.Response{
		CommandSeq: serverAppSeq,
		Code:       code,
	}

	handshake.send(response)
}

// TODO: merge appSeq
func (handshake *handshakeData) sendCommandWithResponse(req commands.CommandWithAppSeq) error {
	appSeq := handshake.conn.NextAppSeq()
	req.SetAppSeq(appSeq)
	handshake.send(req)

	res, err := handshake.receive()
	if err != nil {
		return err
	}

	response, ok := res.(*commands.Response)
	if !ok {
		return fmt.Errorf("unexpected response: %s %+v", reflect.TypeOf(res), res)
	}

	if response.CommandSeq != appSeq {
		return fmt.Errorf("appseq mismatch")
	}

	if response.Code != commands.ResponseCodeSuccess {
		return fmt.Errorf("response contains failure: %d", response.Code)
	}

	return nil
}

func (handshake *handshakeData) send(cmd commands.Command) {
	handshake.conn.Write(cmd)
}

func (handshake *handshakeData) receive() (commands.Command, error) {
	select {
	case <-handshake.ctx.Done():
		return nil, errors.New("client closing")

	case cmd := <-handshake.conn.Read():
		return cmd, nil

	case err := <-handshake.conn.Errors():
		return nil, err
	}
}
