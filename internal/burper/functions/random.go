package functions

import (
	burper "burp/internal/burper"
	"github.com/dchest/uniuri"
	"math/rand"
	"strconv"
)

var _ = burper.Add(burper.Function{
	Name: "random",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, call.MissingArgumentErr("length", "uint16")
		}
		length, err := strconv.ParseUint(call.Args[0], 10, 16)
		if err != nil {
			return nil, err
		}
		return []byte(uniuri.NewLen(int(length))), nil
	},
})

var _ = burper.Add(burper.Function{
	Name: "randn",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, call.MissingArgumentErr("bound", "uint64")
		}
		length, err := strconv.ParseInt(call.Args[0], 10, 16)
		if err != nil {
			return nil, call.FormatErr(err)
		}
		return []byte(strconv.FormatInt(rand.Int63n(length), 10)), nil
	},
})
