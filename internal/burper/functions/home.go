package functions

import (
	"burp/internal/burper"
	"burp/pkg/fileutils"
)

var _ = burper.Add(burper.Function{
	Name: "home",
	Transformer: func(call *burper.FunctionCall, flow *burper.Flow) ([]byte, error) {
		args, err := call.ExecStack(flow)
		if err != nil {
			return nil, call.FormatErr(err)
		}
		return []byte(fileutils.JoinHomePath(args...)), nil
	},
})
