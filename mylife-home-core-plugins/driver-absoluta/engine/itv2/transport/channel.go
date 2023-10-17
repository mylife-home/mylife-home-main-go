package transport

import (
	"bytes"
)

type Processor interface {
	Process(data *bytes.Buffer) error
}

type ProcessorNode interface {
	SetNext(next Processor)
	Process(data *bytes.Buffer) error
}

type ChannelNode struct {
	encoder ProcessorNode
	decoder ProcessorNode
}

func MakeChannelNode(encoder ProcessorNode, decoder ProcessorNode) *ChannelNode {
	node := &ChannelNode{
		encoder,
		decoder,
	}

	if node.encoder == nil {
		node.encoder = &NoopProcessorNode{}
	}

	if node.decoder == nil {
		node.decoder = &NoopProcessorNode{}
	}

	return node
}

type NoopProcessorNode struct {
	next Processor
}

func (noop *NoopProcessorNode) SetNext(next Processor) {
	noop.next = next
}

func (noop *NoopProcessorNode) Process(data *bytes.Buffer) error {
	return noop.next.Process(data)
}
