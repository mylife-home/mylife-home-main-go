package itv2

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
)

type socket struct {
	conn        net.Conn
	write       chan []byte
	read        chan []byte
	errors      chan error
	closing     atomic.Bool
	workersSync sync.WaitGroup
}

const writeChannelSize = 10
const readChannelSize = 1000
const errorsChannelSize = 10
const readBufferSize = 1024

func makeSocket(ctx context.Context, address string) (*socket, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	sock := &socket{
		conn:   conn,
		write:  make(chan []byte, writeChannelSize),
		read:   make(chan []byte, readChannelSize),
		errors: make(chan error, errorsChannelSize),
	}

	sock.closing.Store(false)
	sock.workersSync.Add(2)

	go sock.writer()
	go sock.reader()

	return sock, nil
}

func (sock *socket) Close() {
	sock.closing.Store(true)

	close(sock.write)
	sock.conn.Close()

	sock.workersSync.Wait()

	close(sock.read)
	close(sock.errors)
}

func (sock *socket) Write(data []byte) {
	sock.write <- data
}

// Read 0 = close
func (sock *socket) Read() <-chan []byte {
	return sock.read
}

func (sock *socket) Errors() <-chan error {
	return sock.errors
}

func (sock *socket) writer() {
	defer sock.workersSync.Done()

	for data := range sock.write {
		for len(data) > 0 {
			n, err := sock.conn.Write(data)
			if sock.closing.Load() {
				return
			}

			if err != nil {
				sock.errors <- err
				break
			}

			data = data[n:]
		}
	}
}

func (sock *socket) reader() {
	defer sock.workersSync.Done()

	for {
		data := make([]byte, readBufferSize)
		n, err := sock.conn.Read(data)
		if sock.closing.Load() {
			return
		}

		if err != nil {
			sock.errors <- err
			break
		}

		sock.read <- data[:n]
	}
}
