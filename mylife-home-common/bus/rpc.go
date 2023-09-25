package bus

import (
	"fmt"
	"math/rand"
	"mylife-home-common/tools"
	"time"
)

const rpcDomain = "rpc"
const rpcServices = "services"
const rpcReplies = "replies"

const RpcTimeout = time.Second * 2

type Rpc struct {
	client   *client
	services map[string]RpcService
}

func newRpc(client *client) *Rpc {
	return &Rpc{
		client:   client,
		services: make(map[string]RpcService),
	}
}

type RpcService interface {
	init(client *client, address string) error
	terminate() error
}

func (rpc *Rpc) Serve(address string, svc RpcService) error {
	_, exists := rpc.services[address]
	if exists {
		panic(fmt.Errorf("service with address '%s' does already exist", address))
	}

	if err := svc.init(rpc.client, address); err != nil {
		return err
	}

	rpc.services[address] = svc
	return nil
}

func (rpc *Rpc) Unserve(address string) error {
	svc, exists := rpc.services[address]
	if !exists {
		panic(fmt.Errorf("service with address '%s' does not exist", address))
	}

	err := svc.terminate()
	delete(rpc.services, address)
	return err
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

	replyChan := make(chan []byte, 1)
	msgToken := rpc.client.OnMessage().Register(func(m *message) {
		if m.InstanceName() != rpc.client.InstanceName() || m.Domain() != rpcDomain || m.Path() != rpcReplies+"/"+replyId {
			return
		}

		replyChan <- m.Payload()
	})

	defer rpc.client.OnMessage().Unregister(msgToken)

	if err := rpc.client.Subscribe(replyTopic); err != nil {
		return nilOutput, err
	}

	if err := rpc.client.Publish(remoteTopic, Encoding.WriteJson(&request), false); err != nil {
		return nilOutput, err
	}

	var reply []byte

	select {
	case reply = <-replyChan:
		// Go ahead
	case <-time.After(timeout):
		return nilOutput, fmt.Errorf("timeout occured while waiting for message on topic '%s' (call address: '%s', timeout: %s)", replyTopic, address, timeout)
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

type rpcServiceImpl[TInput any, TOutput any] struct {
	client         *client
	address        string
	implementation func(TInput) (TOutput, error)
	msgToken       tools.RegistrationToken
}

func NewRpcService[TInput any, TOutput any](implementation func(TInput) (TOutput, error)) RpcService {
	return &rpcServiceImpl[TInput, TOutput]{
		implementation: implementation,
	}
}

func (svc *rpcServiceImpl[TInput, TOutput]) init(client *client, address string) error {
	svc.client = client
	svc.address = address

	svc.msgToken = svc.client.OnMessage().Register(svc.onMessage)
	return svc.client.Subscribe(svc.buildTopic())
}

func (svc *rpcServiceImpl[TInput, TOutput]) terminate() error {
	err := svc.client.Unsubscribe(svc.buildTopic())
	svc.client.OnMessage().Unregister(svc.msgToken)
	return err
}

func (svc *rpcServiceImpl[TInput, TOutput]) onMessage(m *message) {
	if m.InstanceName() != svc.client.InstanceName() || m.Domain() != rpcDomain || m.Path() != rpcServices+"/"+svc.address {
		return
	}

	fireAndForget(func() error {
		var req request[TInput]
		Encoding.ReadTypedJson(m.Payload(), &req)
		resp := svc.handle(&req)
		output := Encoding.WriteJson(resp)
		return svc.client.Publish(req.ReplyTopic, output, false)
	})
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
