package serialization

import "github.com/gookit/goutil/errorx/panics"

type BitMaskBytes interface {
	Get() []byte // Note: buffer is updated in place when 'set'
}

type BitMask struct {
	target       BitMaskBytes
	offset       int
	littleEndian bool
}

func NewBitMask(target BitMaskBytes, offset int, littleEndian bool) *BitMask {
	return &BitMask{
		target:       target,
		offset:       offset,
		littleEndian: littleEndian,
	}
}

func NewBitMaskFromBytes(target []byte, offset int, littleEndian bool) *BitMask {
	return &BitMask{
		target:       &bitmaskBytes{data: target},
		offset:       offset,
		littleEndian: littleEndian,
	}
}

var _ BitMaskBytes = (*bitmaskBytes)(nil)

type bitmaskBytes struct {
	data []byte
}

func (bb *bitmaskBytes) Get() []byte {
	return bb.data
}

func (bitMask *BitMask) Size() int {
	data := bitMask.target.Get()
	return len(data)*8 + bitMask.offset
}

func (bitMask *BitMask) Offset() int {
	return bitMask.offset
}

func (bitMask *BitMask) Set(index int, value bool) {
	panics.IsTrue(index >= bitMask.offset && index < bitMask.Size())

	data := bitMask.target.Get()
	realIndex := index - bitMask.offset
	byteIndex := bitMask.byteIndex(realIndex)
	byteMask := bitMask.byteMask(realIndex)

	if value {
		data[byteIndex] = data[byteIndex] | byteMask
	} else {
		data[byteIndex] = data[byteIndex] & ^byteMask
	}
}

func (bitMask *BitMask) Get(index int) bool {
	panics.IsTrue(index >= bitMask.offset && index < bitMask.Size())

	data := bitMask.target.Get()
	realIndex := index - bitMask.offset
	byteIndex := bitMask.byteIndex(realIndex)
	byteMask := bitMask.byteMask(realIndex)

	return (data[byteIndex] & byteMask) != 0
}

func (bitMask *BitMask) byteIndex(realIndex int) int {
	realIndex = realIndex >> 3
	if bitMask.littleEndian {
		return realIndex
	} else {
		data := bitMask.target.Get()
		return len(data) - realIndex - 1
	}
}

func (bitMask *BitMask) byteMask(realIndex int) byte {
	return byte(0x1) << (realIndex & 0x7)
}

func (bitMask *BitMask) GetTrueIndexes() []int {
	array := make([]int, 0)

	for index := bitMask.Offset(); index < bitMask.Size(); index += 1 {
		if bitMask.Get(index) {
			array = append(array, index)
		}
	}

	return array
}

func (bitMask *BitMask) GetBitset() []bool {
	array := make([]bool, bitMask.Offset()+bitMask.Size())

	for index := bitMask.Offset(); index < bitMask.Size(); index += 1 {
		array[index] = bitMask.Get(index)
	}

	return array
}
