package functions

import (
	"burp/internal/burper"
	"burp/pkg/fileutils"
)

var _ = burper.Add(burper.Function{
	Name: "home",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		args, err := call.ExecStack(tree)
		if err != nil {
			return nil, burper.CreateError(call, err.Error())
		}
		return []byte(fileutils.JoinHomePath(args...)), nil
	},
})
