package itv2

import (
	"bytes"
	"context"
	"errors"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/transport"
	"sync"
)

type connection struct {
	sock       *socket
	pipeline   *transport.Pipeline
	write      chan commands.Command
	read       chan commands.Command
	errors     chan error
	exit       chan struct{}
	workerSync sync.WaitGroup
	appSeq     int
	appSeqLock sync.Mutex
}

var errConnectionRemotelyClosed = errors.New("connection closed by remote side")

func makeConnection(ctx context.Context, address string) (*connection, error) {
	sock, err := makeSocket(ctx, address)
	if err != nil {
		return nil, err
	}

	conn := &connection{
		sock:   sock,
		write:  make(chan commands.Command, 10),
		read:   make(chan commands.Command, 10),
		errors: make(chan error, 10),
		exit:   make(chan struct{}, 1),
	}

	receiveCommand := func(buffer *bytes.Buffer) error {
		cmd, err := commands.DecodeCommand(buffer)
		if err != nil {
			return err
		}

		// logger.Debugf("Recv command %s %+v", reflect.TypeOf(cmd), cmd)
		conn.read <- cmd
		return nil
	}

	sendData := func(buffer *bytes.Buffer) error {
		conn.sock.Write(buffer.Bytes())
		return nil
	}

	conn.pipeline = transport.MakePipeline(receiveCommand, sendData)

	conn.workerSync.Add(1)
	go conn.worker()

	return conn, nil
}

func (conn *connection) worker() {
	defer conn.workerSync.Done()

	for {
		select {
		case <-conn.exit:
			return

		case err := <-conn.sock.Errors():
			conn.errors <- err

		case data := <-conn.sock.Read():
			conn.processRead(data)

		case cmd := <-conn.write:
			conn.processWrite(cmd)
		}
	}
}

func (conn *connection) processRead(data []byte) {
	if len(data) == 0 {
		conn.errors <- errConnectionRemotelyClosed
		return
	}

	if err := conn.pipeline.ReceiveData(bytes.NewBuffer(data)); err != nil {
		conn.errors <- err
		return
	}
}

func (conn *connection) processWrite(cmd commands.Command) {
	// logger.Debugf("Send command %s %+v", reflect.TypeOf(cmd), cmd)
	data, err := commands.EncodeCommand(cmd)
	if err != nil {
		conn.errors <- err
		return
	}

	if err := conn.pipeline.SendCommand(data); err != nil {
		conn.errors <- err
		return
	}
}

func (conn *connection) Write(cmd commands.Command) {
	conn.write <- cmd
}

func (conn *connection) Read() <-chan commands.Command {
	return conn.read
}

func (conn *connection) Errors() <-chan error {
	return conn.errors
}

func (conn *connection) Close() {
	close(conn.exit)
	conn.workerSync.Wait()

	conn.sock.Close()
}

func (conn *connection) NextAppSeq() uint8 {
	conn.appSeqLock.Lock()
	defer conn.appSeqLock.Unlock()

	conn.appSeq = (conn.appSeq + 1) % 256
	return uint8(conn.appSeq)
}
