package functions

import (
	burper "burp/internal/burper"
	"strings"
)

var _ = burper.Add(burper.Function{
	Name: "concat",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		args, err := call.ExecStack(flow)
		if err != nil {
			return nil, call.FormatErr(err)
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
				return nil, call.MissingArgumentErr("addends", "string")
			}
			res = args[0] + args[1]
		}
		return []byte(res), nil
	},
})
