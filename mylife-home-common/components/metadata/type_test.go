package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testParseType(t *testing.T, str string) {
	typ, err := ParseType(str)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, typ.String(), str)
}

func TestParseRange(t *testing.T) {
	testParseType(t, "range[-12;42]")
}

func TestParseText(t *testing.T) {
	testParseType(t, "text")
}

func TestParseFloat(t *testing.T) {
	testParseType(t, "float")
}

func TestParseBool(t *testing.T) {
	testParseType(t, "bool")
}

func TestParseEnum(t *testing.T) {
	testParseType(t, "enum{one,two,three}")
}

func TestParseComplex(t *testing.T) {
	testParseType(t, "complex")
}
