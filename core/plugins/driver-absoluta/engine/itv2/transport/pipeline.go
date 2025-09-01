package transport

import (
	"bytes"

	"github.com/gookit/goutil/errorx/panics"
	"golang.org/x/exp/slices"
)

type Pipeline struct {
	encoder Processor
	decoder Processor
}

func (pipeline *Pipeline) SendCommand(data *bytes.Buffer) error {
	return pipeline.encoder.Process(data)
}

func (pipeline *Pipeline) ReceiveData(data *bytes.Buffer) error {
	return pipeline.decoder.Process(data)
}

func MakePipeline(receiveCommand func(*bytes.Buffer) error, sendData func(*bytes.Buffer) error) *Pipeline {
	nodes := []*ChannelNode{
		MakeFrameChannelNode(),
		//aes
		MakeDataChannelNode(),
		MakeTransportChannelNode(),
		MakeMultiplePacketChannelNode(),
	}

	encoders := make([]ProcessorNode, len(nodes))
	decoders := make([]ProcessorNode, len(nodes))

	for index, node := range nodes {
		encoders[index] = node.encoder
		decoders[index] = node.decoder
	}

	slices.Reverse(encoders)

	encoder := MakeChannel(encoders, &CallbackChannelNode{callback: sendData})
	decoder := MakeChannel(decoders, &CallbackChannelNode{callback: receiveCommand})

	return &Pipeline{
		encoder: encoder,
		decoder: decoder,
	}
}

type CallbackChannelNode struct {
	callback func(*bytes.Buffer) error
}

func (cbnode *CallbackChannelNode) Process(data *bytes.Buffer) error {
	return cbnode.callback(data)
}

func MakeChannel(nodes []ProcessorNode, tail Processor) Processor {
	panics.IsTrue(len(nodes) > 0)
	var prev ProcessorNode
	var head Processor

	for _, node := range nodes {
		if head == nil {
			head = node
		} else {
			prev.SetNext(node)
		}

		prev = node
	}

	prev.SetNext(tail)

	return head
}
