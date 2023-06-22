package functions

import (
	burper "burp/internal/burper"
	"burp/pkg/utils"
	"github.com/alexedwards/argon2id"
)

var _ = burper.Add(burper.Function{
	Name: "hash",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, burper.CreateMissingArgumentError(call, utils.Array("key", "string"))
		}
		args, err := call.ExecStack(tree)
		if err != nil {
			return nil, burper.CreateError(call, err.Error())
		}
		hash, err := argon2id.CreateHash(args[0], argon2id.DefaultParams)
		if err != nil {
			return nil, burper.CreateError(call, err.Error())
		}
		return []byte(hash), nil
	},
})
