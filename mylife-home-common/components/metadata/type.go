package metadata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Type interface {
	String() string
	Validate(value any) bool
}

var parser = regexp.MustCompile(`([a-z]+)(.*)`)
var rangeParser = regexp.MustCompile(`\[(-?\d+);(-?\d+)\]`)
var enumParser = regexp.MustCompile(`{(.[\w_\-,]+)}`)

func ParseType(value string) (Type, error) {
	matchs := parser.FindStringSubmatch(value)
	if matchs == nil {
		return nil, fmt.Errorf("invalid type '%s'", value)
	}

	var baseType, args string

	switch len(matchs) {
	case 2:
		baseType = matchs[1]

	case 3:
		baseType = matchs[1]
		args = matchs[2]

	default:
		return nil, fmt.Errorf("invalid type '%s' (bad match len)", value)
	}

	switch baseType {
	case "range":
		matchs := rangeParser.FindStringSubmatch(args)
		if matchs == nil || len(matchs) != 3 {
			return nil, fmt.Errorf("invalid type '%s' (bad args)", value)
		}

		min, err := strconv.ParseInt(matchs[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid type '%s' (%f)", value, err)
		}

		max, err := strconv.ParseInt(matchs[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid type '%s' (%f)", value, err)
		}

		if min >= max {
			return nil, fmt.Errorf("invalid type '%s' (min >= mX)", value)
		}

		return &RangeType{min: min, max: max}, nil

	case "text":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return &TextType{}, nil

	case "float":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return &FloatType{}, nil

	case "bool":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return &BoolType{}, nil

	case "enum":
		matchs := enumParser.FindStringSubmatch(args)
		if matchs == nil || len(matchs) != 2 {
			return nil, fmt.Errorf("invalid type '%s' (bad args)", value)
		}

		values := strings.Split(matchs[1], ",")
		if len(values) < 2 {
			return nil, fmt.Errorf("invalid type '%s' (bad args)", value)
		}

		return &EnumType{values: values}, nil

	case "complex":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return &ComplexType{}, nil

	default:
		return nil, fmt.Errorf("invalid type '%s' (unknown type)", value)
	}
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

type TextType struct {
}

func (typ *TextType) String() string {
	return "text"
}

func (typ *TextType) Validate(value any) bool {
	_, ok := value.(string)
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

type BoolType struct {
}

func (typ *BoolType) String() string {
	return "bool"
}

func (typ *BoolType) Validate(value any) bool {
	_, ok := value.(bool)
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
