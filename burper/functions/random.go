package functions

import (
	"burp/burper"
	"burp/utils"
	"github.com/dchest/uniuri"
	"strconv"
)

var _ = burper.Add(burper.Function{
	Name: "random",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, burper.CreateMissingArgumentError(call, utils.Array("length", "uint16"))
		}
		length, err := strconv.ParseUint(call.Args[0], 10, 16)
		if err != nil {
			return nil, err
		}
		return []byte(uniuri.NewLen(int(length))), nil
	},
})
