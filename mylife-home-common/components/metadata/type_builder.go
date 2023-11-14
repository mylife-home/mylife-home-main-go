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

func MakeTypeEnum(values ...string) Type {
	panics.IsTrue(len(values) > 0)

	uniques := make(map[string]struct{})
	for _, value := range values {
		uniques[value] = struct{}{}
	}
	panics.IsTrue(len(uniques) == len(values))

	return &EnumType{values}
}

func MakeTypeComplex() Type {
	return &ComplexType{}
}
