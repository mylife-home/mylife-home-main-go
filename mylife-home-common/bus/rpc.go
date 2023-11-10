package bus

import (
	"fmt"
	"math/rand"
	"mylife-home-common/tools"
	"sync"
	"time"
)

const rpcDomain = "rpc"
const rpcServices = "services"
const rpcReplies = "replies"

const RpcTimeout = time.Second * 2

type Rpc struct {
	client     *client
	services   map[string]RpcService
	mux        sync.RWMutex
	onlineChan chan bool
}

func newRpc(client *client) *Rpc {
	rpc := &Rpc{
		client:     client,
		services:   make(map[string]RpcService),
		onlineChan: make(chan bool),
	}

	go rpc.worker()
	rpc.client.Online().Subscribe(rpc.onlineChan, false)

	return rpc
}

func (rpc *Rpc) terminate() {
	rpc.client.Online().Unsubscribe(rpc.onlineChan)
	close(rpc.onlineChan)

	rpc.mux.Lock()
	defer rpc.mux.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(len(rpc.services))

	for address, svc := range rpc.services {
		address, svc := address, svc
		go func() {
			defer wg.Done()
			if err := svc.unbind(); err != nil {
				logger.WithError(err).Errorf("Could not unbind service '%s'", address)
			}
		}()
	}

	wg.Wait()
	clear(rpc.services)
}

func (rpc *Rpc) worker() {
	for online := range rpc.onlineChan {
		if online {
			go rpc.rebind()
		}
	}
}

func (rpc *Rpc) rebind() {
	rpc.mux.Lock()
	defer rpc.mux.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(len(rpc.services))

	for address, svc := range rpc.services {
		address, svc := address, svc
		go func() {
			defer wg.Done()
			if err := svc.bind(); err != nil {
				logger.WithError(err).Errorf("Could not rebind service '%s'", address)
			}
		}()
	}
}

type RpcService interface {
	setup(client *client, address string)
	bind() error
	unbind() error
}

func (rpc *Rpc) Serve(address string, svc RpcService) {
	rpc.mux.Lock()
	defer rpc.mux.Unlock()

	_, exists := rpc.services[address]
	if exists {
		panic(fmt.Errorf("service with address '%s' does already exist", address))
	}

	svc.setup(rpc.client, address)
	rpc.services[address] = svc

	go func() {
		if rpc.client.Online().Get() {
			if err := svc.bind(); err != nil {
				logger.WithError(err).Errorf("Could not bind service '%s'", address)
			}
		}
	}()
}

func (rpc *Rpc) Unserve(address string) {
	rpc.mux.Lock()
	defer rpc.mux.Unlock()

	svc, exists := rpc.services[address]
	if !exists {
		panic(fmt.Errorf("service with address '%s' does not exist", address))
	}

	delete(rpc.services, address)

	go func() {
		if err := svc.unbind(); err != nil {
			logger.WithError(err).Errorf("Could not unbind service '%s'", address)
		}
	}()
}

// Cannot use member function because of generic
func RpcCall[TInput any, TOutput any](rpc *Rpc, targetInstance string, address string, data TInput, timeout time.Duration) (TOutput, error) {
	replyId := randomTopicPart()
	replyTopic := rpc.client.BuildTopic(rpcDomain, rpcReplies, replyId)
	remoteTopic := rpc.client.BuildRemoteTopic(targetInstance, rpcDomain, rpcServices, address)
	var nilOutput TOutput

	request := request[TInput]{
		Input:      data,
		ReplyTopic: replyTopic,
	}

	replyChan := make(chan []byte, 10)
	onMessage := func(m *message) {
		replyChan <- m.Payload()
		close(replyChan)
	}

	onlineChan := make(chan bool, 10)
	rpc.client.Online().Subscribe(onlineChan, true)
	defer func() {
		rpc.client.Online().Unsubscribe(onlineChan)
	}()

	if err := rpc.client.Subscribe(replyTopic, onMessage); err != nil {
		return nilOutput, err
	}

	defer func() {
		if err := rpc.client.Unsubscribe(replyTopic); err != nil {
			logger.WithError(err).Errorf("could not unregister from reply topic '%s'", replyTopic)
		}
	}()

	if err := rpc.client.Publish(remoteTopic, Encoding.WriteJson(&request)); err != nil {
		return nilOutput, err
	}

	var reply []byte

	select {
	case online := <-onlineChan:
		if !online {
			return nilOutput, fmt.Errorf("connection lost while waiting for message on topic '%s' (call address: '%s')", replyTopic, address)
		}
	case <-time.After(timeout):
		return nilOutput, fmt.Errorf("timeout occured while waiting for message on topic '%s' (call address: '%s', timeout: %s)", replyTopic, address, timeout)
	case reply = <-replyChan:
		// Go ahead
	}

	var resp response[TOutput]
	Encoding.ReadTypedJson(reply, &resp)

	if respErr := resp.Error; respErr != nil {
		// Log the stacktrace here but do not forward it
		logger.Errorf("Remote error: %s, stacktrace: %s", respErr.Message, respErr.Stacktrace)

		return nilOutput, fmt.Errorf("remote error: %s", respErr.Message)
	}

	return *resp.Output, nil
}

var _ RpcService = (*rpcServiceImpl[int, int])(nil)

type rpcServiceImpl[TInput any, TOutput any] struct {
	client         *client
	address        string
	implementation func(TInput) (TOutput, error)
}

func NewRpcService[TInput any, TOutput any](implementation func(TInput) (TOutput, error)) RpcService {
	return &rpcServiceImpl[TInput, TOutput]{
		implementation: implementation,
	}
}

func (svc *rpcServiceImpl[TInput, TOutput]) setup(client *client, address string) {
	svc.client = client
	svc.address = address
}

func (svc *rpcServiceImpl[TInput, TOutput]) bind() error {
	return svc.client.Subscribe(svc.buildTopic(), svc.handleMessage)
}

func (svc *rpcServiceImpl[TInput, TOutput]) unbind() error {
	return svc.client.Unsubscribe(svc.buildTopic())
}

func (svc *rpcServiceImpl[TInput, TOutput]) handleMessage(m *message) {
	go func() {
		var req request[TInput]
		Encoding.ReadTypedJson(m.Payload(), &req)

		resp := svc.handle(&req)

		output := Encoding.WriteJson(resp)
		if err := svc.client.Publish(req.ReplyTopic, output); err != nil {
			logger.WithError(err).Errorf("Could not send RPC reply to topic '%s'", req.ReplyTopic)
		}
	}()
}

func (svc *rpcServiceImpl[TInput, TOutput]) handle(req *request[TInput]) *response[TOutput] {
	output, err := svc.implementation(req.Input)

	if err != nil {
		return &response[TOutput]{
			Error: &reponseError{
				Message:    err.Error(),
				Stacktrace: tools.GetStackTraceStr(err),
			},
		}
	} else {
		return &response[TOutput]{
			Output: &output,
		}
	}
}

func (svc *rpcServiceImpl[TInput, TOutput]) buildTopic() string {
	return svc.client.BuildTopic(rpcDomain, rpcServices, svc.address)
}

func randomTopicPart() string {
	const charset = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz0123456789"
	const charsetLen = len(charset)
	const len = 16

	array := make([]byte, len)
	for index := range array {
		array[index] = charset[rand.Intn(charsetLen)]
	}

	return string(array)
}

type request[TInput any] struct {
	Input      TInput `json:"input"`
	ReplyTopic string `json:"replyTopic"`
}

type response[TOutput any] struct {
	Output *TOutput      `json:"output"`
	Error  *reponseError `json:"error"`
}

type reponseError struct {
	Message    string `json:"message"`
	Stacktrace string `json:"stacktrace"`
}
