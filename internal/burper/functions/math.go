package functions

import (
	burper "burp/internal/burper"
	"strconv"
)

var _ = burper.Add(burper.Function{
	Name: "add",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		return numericOperationTransformer(call, flow, func(origin *int64, change int64) {
			*origin += change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "sub",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		return numericOperationTransformer(call, flow, func(origin *int64, change int64) {
			*origin -= change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "div",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		return numericOperationTransformer(call, flow, func(origin *int64, change int64) {
			*origin /= change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "mpy",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		return numericOperationTransformer(call, flow, func(origin *int64, change int64) {
			*origin *= change
		})
	},
})

var _ = burper.Add(burper.Function{
	Name: "mod",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		return numericOperationTransformer(call, flow, func(origin *int64, change int64) {
			*origin %= change
		})
	},
})

func numericOperationTransformer(call *burper.FunctionCall, flow *burper.Flow, operation func(origin *int64, change int64)) ([]byte, error) {
	args, err := call.ExecStack(flow)
	if err != nil {
		return nil, call.FormatErr(err)
	}
	if len(args) < 2 {
		return nil, call.MissingArgumentErr("numbers", "vararg numbers")
	}
	origin, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, call.Err("the value " + args[0] + " is not a number.")
	}
	for _, arg := range args[1:] {
		number, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, call.Err("the value " + arg + " is not a number.")
		}
		operation(&origin, number)
	}
	return []byte(strconv.FormatInt(origin, 10)), nil
}
