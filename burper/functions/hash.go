package functions

import (
	"burp/burper"
	"burp/utils"
)

var _ = burper.Add(burper.Function{
	Name: "hash",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, burper.CreateMissingArgumentError(call, utils.Array("key", "string"))
		}
		// TODO: Add argon2id hashing here.
		hash := []byte(call.Args[0])
		return hash, nil
	},
})
