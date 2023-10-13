package serialization

import (
	"encoding/hex"

	"github.com/gookit/goutil/errorx/panics"
)

func BCDEncode(target []byte, value string) {
	raw, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}

	panics.IsTrue(len(raw) == len(target))

	copy(target, raw)
}
