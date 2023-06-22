package functions

import (
	burper "burp/internal/burper"
	"burp/pkg/utils"
	"strings"
)

var _ = burper.Add(burper.Function{
	Name: "concat",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		args, err := call.ExecStack(tree)
		if err != nil {
			return nil, burper.CreateError(call, err.Error())
		}
		var res string
		if len(args) > 2 {
			var b strings.Builder
			for _, arg := range args {
				b.WriteString(arg)
			}
			res = b.String()
		} else {
			if len(args) < 2 {
				return nil, burper.CreateMissingArgumentError(call, utils.Array("addends", "string"))
			}
			res = args[0] + args[1]
		}
		return []byte(res), nil
	},
})
