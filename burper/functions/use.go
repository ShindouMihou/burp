package functions

import (
	"burp/burper"
	"burp/utils"
)

var _ = burper.Add(burper.Function{
	Name: "use",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		if len(call.Args) < 1 {
			return nil, burper.CreateMissingArgumentError(call, utils.Array("key", "string"))
		}
		v, e := tree.Store[call.Args[0]]
		if !e {
			return nil, burper.CreateError(call, "no such element called \""+call.Args[0]+"\"")
		}
		return v, nil
	},
})
