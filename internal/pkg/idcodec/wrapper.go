package idcodec

import (
	"errors"

	"github.com/mlogclub/simple/common/strs"
	"github.com/spf13/cast"
)

var ErrNotInitialized = errors.New("idcodec is not initialized")

var Instance *Codec

func Init(key uint64) {
	Instance = NewCodec(key)
}

// Encode uses the global codec instance.
func Encode(id int64) string {
	if Instance == nil {
		panic(ErrNotInitialized)
	}
	if id <= 0 {
		return ""
	}
	return Instance.Encode(id)
}

// Decode uses the global codec instance.
func Decode(s string) int64 {
	if strs.IsBlank(s) {
		return 0
	}
	if id := cast.ToInt64(s); id > 0 {
		return id
	}
	if Instance == nil {
		panic(ErrNotInitialized)
	}
	ret, _ := Instance.Decode(s)
	return ret
}
