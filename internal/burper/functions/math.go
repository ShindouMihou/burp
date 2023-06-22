package functions

import (
	burper "burp/internal/burper"
	"burp/pkg/utils"
	"strconv"
)

var _ = burper.Add(burper.Function{
	Name: "add",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		return numericOperationTransformer(call, tree, func(origin *int64, change int64) {
			*origin += change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "sub",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		return numericOperationTransformer(call, tree, func(origin *int64, change int64) {
			*origin -= change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "div",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		return numericOperationTransformer(call, tree, func(origin *int64, change int64) {
			*origin /= change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "mpy",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		return numericOperationTransformer(call, tree, func(origin *int64, change int64) {
			*origin *= change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "mod",
	Transformer: func(call *burper.Call, tree *burper.Tree) ([]byte, error) {
		return numericOperationTransformer(call, tree, func(origin *int64, change int64) {
			*origin %= change
		})
	},
})

func numericOperationTransformer(call *burper.Call, tree *burper.Tree, operation func(origin *int64, change int64)) ([]byte, error) {
	args, err := call.ExecStack(tree)
	if err != nil {
		return nil, burper.CreateError(call, err.Error())
	}
	if len(args) < 2 {
		return nil, burper.CreateMissingArgumentError(call, utils.Array("numbers", "number"))
	}
	origin, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, burper.CreateError(call, "the value "+args[0]+" is not a number.")
	}
	for _, arg := range args[1:] {
		number, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, burper.CreateError(call, "the value "+arg+" is not a number.")
		}
		operation(&origin, number)
	}
	return []byte(strconv.FormatInt(origin, 10)), nil
}
