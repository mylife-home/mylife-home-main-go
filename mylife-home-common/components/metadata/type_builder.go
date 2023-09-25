package metadata

import (
	"github.com/gookit/goutil/errorx/panics"
)

func MakeTypeRange(min int64, max int64) Type {
	panics.IsTrue(min < max)

	return &RangeType{min, max}
}

func MakeTypeText() Type {
	return &TextType{}
}

func MakeTypeFloat() Type {
	return &FloatType{}
}

func MakeTypeBool() Type {
	return &BoolType{}
}

// Note: values should be sorted
func MakeTypeEnum(values ...string) Type {
	panics.IsTrue(values != nil && len(values) > 0)

	return &EnumType{values}
}

func MakeTypeComplex() Type {
	return &ComplexType{}
}
