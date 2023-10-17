package serialization

type RemainArray[T any] struct {
	Items []T
}

var _ CustomMarshal = (*RemainArray[any])(nil)

func (value *RemainArray[T]) Write(writer BinaryWriter) error {
	for _, item := range value.Items {
		if err := MarshalItem(writer, item); err != nil {
			return err
		}
	}

	return nil
}

func (value *RemainArray[T]) Read(reader BinaryReader) error {
	value.Items = make([]T, 0)

	for reader.Remain() > 0 {
		var item T
		if err := UnmarshalItem(reader, &item); err != nil {
			return err
		}

		value.Items = append(value.Items, item)
	}

	return nil
}
