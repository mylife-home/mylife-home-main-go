package itv2

import (
	"fmt"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"reflect"
	"sync"
	"time"
)

const transactionTimeout = time.Second * 10

type transactionResponse struct {
	cmd commands.Command
	err error
}

type transaction struct {
	appSeq       uint8
	cmdCode      uint16
	reqData      any
	response     chan transactionResponse
	responseDone bool
	responseMux  sync.Mutex
}

func makeTransaction(cmd commands.CommandWithAppSeq) *transaction {
	code, err := commands.GetCommandCode(cmd)
	if err != nil {
		panic(err)
	}

	trans := &transaction{
		appSeq:   cmd.GetAppSeq(),
		cmdCode:  code,
		reqData:  nil,
		response: make(chan transactionResponse, 1),
	}

	if cmd, ok := cmd.(*commands.Request); ok {
		trans.reqData = cmd.ReqData
	}

	return trans
}

// client side
func (trans *transaction) Wait() (commands.Command, error) {
	select {
	case respData := <-trans.response:
		return respData.cmd, respData.err

	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("transaction timeout")
	}
}

// connection side
func (trans *transaction) TryResponse(cmd commands.Command) bool {
	switch cmd := cmd.(type) {
	case *commands.Error:
		if cmd.ReceivedCommand == trans.cmdCode {
			resp := transactionResponse{
				err: fmt.Errorf("error %d received", cmd.ErrorCode),
			}

			return trans.setResponse(resp)
		}

	case *commands.Response:
		if cmd.CommandSeq == trans.appSeq {
			var resp transactionResponse
			if cmd.Code != commands.ResponseCodeSuccess {
				resp.err = fmt.Errorf("response with error code %s received", cmd.CodeString())
			}
			// else no error but response with no additional data
			return trans.setResponse(resp)
		}

	case commands.ResponseData:
		if trans.reqData != nil && reflect.DeepEqual(trans.reqData, cmd.GetRequest()) {
			resp := transactionResponse{
				cmd: cmd,
			}

			return trans.setResponse(resp)
		}
	}

	return false
}

// connection side
func (trans *transaction) Cancel() bool {
	resp := transactionResponse{
		err: fmt.Errorf("transaction canceled"),
	}

	return trans.setResponse(resp)
}

func (trans *transaction) setResponse(resp transactionResponse) bool {
	trans.responseMux.Lock()
	defer trans.responseMux.Unlock()

	if trans.responseDone {
		return false
	}

	trans.response <- resp
	close(trans.response)
	trans.responseDone = true

	return true
}

type transactionManager struct {
	pendings map[*transaction]struct{}
	lock     sync.Mutex
}

func newTransactionManager() *transactionManager {
	return &transactionManager{
		pendings: make(map[*transaction]struct{}),
	}
}

// connection side
func (manager *transactionManager) CancelAll() {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	for trans := range manager.pendings {
		trans.Cancel()
	}
}

// connection side
func (manager *transactionManager) ProcessCommand(cmd commands.Command) bool {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	for trans := range manager.pendings {
		if trans.TryResponse(cmd) {
			return true
		}
	}

	return false
}

// client side
func (manager *transactionManager) addTransaction(trans *transaction) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	manager.pendings[trans] = struct{}{}
}

// client side
func (manager *transactionManager) removeTransaction(trans *transaction) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	delete(manager.pendings, trans)
}
