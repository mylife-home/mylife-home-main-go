package metadata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

		return MakeTypeRange(min, max), nil

	case "text":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return MakeTypeText(), nil

	case "float":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return MakeTypeFloat(), nil

	case "bool":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return MakeTypeBool(), nil

	case "enum":
		matchs := enumParser.FindStringSubmatch(args)
		if matchs == nil || len(matchs) != 2 {
			return nil, fmt.Errorf("invalid type '%s' (bad args)", value)
		}

		values := strings.Split(matchs[1], ",")
		if len(values) < 2 {
			return nil, fmt.Errorf("invalid type '%s' (bad args)", value)
		}

		return MakeTypeEnum(values...), nil

	case "complex":
		if args != "" {
			return nil, fmt.Errorf("invalid type '%s' (unexpected args)", value)
		}
		return MakeTypeComplex(), nil

	default:
		return nil, fmt.Errorf("invalid type '%s' (unknown type)", value)
	}
}
