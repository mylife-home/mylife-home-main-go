package metadata

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

type Type interface {
	String() string
	Validate(value any) bool
	Equals(other Type) bool
}

type RangeType struct {
	min int64
	max int64
}

func (typ *RangeType) String() string {
	return fmt.Sprintf("range[%d;%d]", typ.min, typ.max)
}

func (typ *RangeType) Validate(value any) bool {
	intValue, ok := value.(int64)
	if !ok {
		return false
	}

	return intValue >= typ.min && intValue <= typ.max
}

func (typ *RangeType) Min() int64 {
	return typ.min
}

func (typ *RangeType) Max() int64 {
	return typ.max
}

func (typ *RangeType) Equals(other Type) bool {
	otherRange, ok := other.(*RangeType)
	if !ok {
		return false
	}

	return typ.min == otherRange.min && typ.max == otherRange.max
}

type TextType struct {
}

func (typ *TextType) String() string {
	return "text"
}

func (typ *TextType) Validate(value any) bool {
	_, ok := value.(string)
	return ok
}

func (typ *TextType) Equals(other Type) bool {
	_, ok := other.(*TextType)
	return ok
}

type FloatType struct {
}

func (typ *FloatType) String() string {
	return "float"
}

func (typ *FloatType) Validate(value any) bool {
	_, ok := value.(float64)
	return ok
}

func (typ *FloatType) Equals(other Type) bool {
	_, ok := other.(*FloatType)
	return ok
}

type BoolType struct {
}

func (typ *BoolType) String() string {
	return "bool"
}

func (typ *BoolType) Validate(value any) bool {
	_, ok := value.(bool)
	return ok
}

func (typ *BoolType) Equals(other Type) bool {
	_, ok := other.(*BoolType)
	return ok
}

type EnumType struct {
	values []string
}

func (typ *EnumType) String() string {
	return fmt.Sprintf("enum{%s}", strings.Join(typ.values, ","))
}

func (typ *EnumType) Validate(value any) bool {
	strValue, ok := value.(string)
	if !ok {
		return false
	}

	return slices.Contains(typ.values, strValue)
}

func (typ *EnumType) Equals(other Type) bool {
	otherEnum, ok := other.(*EnumType)
	if !ok {
		return false
	}

	if len(typ.values) != len(otherEnum.values) {
		return false
	}

	values := make(map[string]struct{})
	for _, value := range typ.values {
		values[value] = struct{}{}
	}

	for _, value := range otherEnum.values {
		if _, exists := values[value]; !exists {
			return false
		}
	}

	return true
}

func (typ *EnumType) NumValues() int {
	return len(typ.values)
}

func (typ *EnumType) Value(index int) string {
	return typ.values[index]
}

type ComplexType struct {
}

func (typ *ComplexType) String() string {
	return "complex"
}

func (typ *ComplexType) Validate(value any) bool {
	return true
}

func (typ *ComplexType) Equals(other Type) bool {
	_, ok := other.(*ComplexType)
	return ok
}
