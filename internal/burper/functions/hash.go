package functions

import (
	burper "burp/internal/burper"
	"github.com/alexedwards/argon2id"
)

var _ = burper.Add(burper.Function{
	Name: "hash",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, call.MissingArgumentErr("key", "string")
		}
		args, err := call.ExecStack(flow)
		if err != nil {
			return nil, call.FormatErr(err)
		}
		hash, err := argon2id.CreateHash(args[0], argon2id.DefaultParams)
		if err != nil {
			return nil, call.FormatErr(err)
		}
		return []byte(hash), nil
	},
})
