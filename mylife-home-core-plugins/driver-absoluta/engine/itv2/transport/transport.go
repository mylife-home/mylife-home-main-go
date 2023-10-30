package transport

import (
	"bytes"
	"fmt"
)

func MakeTransportChannelNode() *ChannelNode {
	data := &TransportData{
		firstMessage:         true,
		prevSequenceNumber:   0,
		sequenceNumber:       0,
		remoteSequenceNumber: 0,
		lastSentACK:          -1,
		lastReceivedACK:      -1,
		sendQueue:            make([]*bytes.Buffer, 0),
	}

	encoder := &TransportEncoder{data: data}
	decoder := &TransportDecoder{data: data, encoder: encoder}

	return MakeChannelNode(encoder, decoder)
}

type TransportData struct {
	firstMessage         bool
	prevSequenceNumber   int
	sequenceNumber       int
	remoteSequenceNumber int
	lastSentACK          int
	lastReceivedACK      int
	sendQueue            []*bytes.Buffer
}

func (data *TransportData) setNextSequenceNumber(isack bool) {
	if !isack && !data.firstMessage {
		data.prevSequenceNumber = data.sequenceNumber
		data.sequenceNumber = data.computeNext(data.sequenceNumber)
	}

	data.firstMessage = false
}

func (data *TransportData) messageReceived(isack bool, remote int, local int) {
	if !isack {
		data.remoteSequenceNumber = remote
	}
	data.lastReceivedACK = local
	data.firstMessage = false
}

func (data *TransportData) setSentRemoteSequenceNumber(remote int) {
	data.lastSentACK = remote
}

func (data *TransportData) isReadyForANewCommand() bool {
	return (data.sequenceNumber == data.lastReceivedACK || data.firstMessage)
}

func (data *TransportData) isOutgoingACKRequired() bool {
	return (data.remoteSequenceNumber != data.lastSentACK && !data.firstMessage)
}

func (data *TransportData) computeNext(value int) int {
	if value < 255 {
		return value + 1
	} else {
		return 1
	}
}

func (data *TransportData) pushSendQueue(input *bytes.Buffer) {
	data.sendQueue = append(data.sendQueue, input)
}

func (data *TransportData) popSendQueue() *bytes.Buffer {
	if data.isSendQueueEmpty() {
		return nil
	}

	input := data.sendQueue[0]
	data.sendQueue = data.sendQueue[1:]
	return input
}

func (data *TransportData) isSendQueueEmpty() bool {
	return len(data.sendQueue) == 0
}

type TransportEncoder struct {
	data *TransportData
	next Processor
}

func (encoder *TransportEncoder) SetNext(next Processor) {
	encoder.next = next
}

func (encoder *TransportEncoder) Process(input *bytes.Buffer) error {

	if !encoder.data.isReadyForANewCommand() || !encoder.data.isSendQueueEmpty() {
		encoder.data.pushSendQueue(input)
		return nil
	}

	return encoder.sendData(input)
}

func (encoder *TransportEncoder) sendData(input *bytes.Buffer) error {
	output := &bytes.Buffer{}

	isack := input.Len() == 0
	encoder.data.setNextSequenceNumber(isack)
	local := encoder.data.sequenceNumber
	remote := encoder.data.remoteSequenceNumber

	if err := output.WriteByte(byte(local)); err != nil {
		return err
	}

	if err := output.WriteByte(byte(remote)); err != nil {
		return err
	}

	if _, err := output.Write(input.Bytes()); err != nil {
		return err
	}

	encoder.data.setSentRemoteSequenceNumber(remote)

	return encoder.next.Process(output)
}

func (encoder *TransportEncoder) sendAck() error {
	return encoder.sendData(&bytes.Buffer{})
}

type TransportDecoder struct {
	data    *TransportData
	encoder *TransportEncoder // To send ack/process sendQueue
	next    Processor
}

func (decoder *TransportDecoder) SetNext(next Processor) {
	decoder.next = next
}

func (decoder *TransportDecoder) Process(input *bytes.Buffer) error {
	b, err := input.ReadByte()
	if err != nil {
		return err
	}

	remote := int(b)

	b, err = input.ReadByte()
	if err != nil {
		return err
	}

	local := int(b)

	isack := input.Len() == 0

	err = decoder.validate(isack, remote, local)

	decoder.data.messageReceived(isack, remote, local)

	if err != nil {
		return err
	}

	if isack {
		// logger.Debug("Got ACK")
		return decoder.dequeue()
	}

	if decoder.data.isOutgoingACKRequired() {
		if err := decoder.encoder.sendAck(); err != nil {
			return err
		}
	}

	if err := decoder.dequeue(); err != nil {
		return err
	}

	return decoder.next.Process(input)
}

func (decoder *TransportDecoder) dequeue() error {
	if !decoder.data.isReadyForANewCommand() {
		return nil
	}

	if input := decoder.data.popSendQueue(); input != nil {
		if err := decoder.encoder.sendData(input); err != nil {
			return err
		}
	}

	return nil
}

func (decoder *TransportDecoder) validate(isack bool, remote int, local int) error {
	if !isack {
		if remote == 0 {
			if !decoder.data.firstMessage {
				return fmt.Errorf("reset request")
			}
		} else {
			if remote == decoder.data.remoteSequenceNumber {
				logger.Warnf("Repeated sequence number %d: ignoring message", remote)
				return nil
			}
			if remote != decoder.data.computeNext(decoder.data.remoteSequenceNumber) {
				return fmt.Errorf("unexpected sequence number: %d instead of %d", remote, decoder.data.computeNext(decoder.data.remoteSequenceNumber))
			}
		}
	}

	if local != decoder.data.sequenceNumber && local != decoder.data.prevSequenceNumber {
		return fmt.Errorf("unexpected remote sequence number: %d instead of %d or %d", local, decoder.data.sequenceNumber, decoder.data.prevSequenceNumber)
	}

	return nil
}
