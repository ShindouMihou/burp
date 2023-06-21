package burper

import "burp/utils"

// ExecStack assesses the arguments of the function to identify whether it is calling another function
// and executes the nested function if it is calling another function.
func (call *Call) ExecStack(tree *Tree) ([]string, error) {
	var stack []string
	for _, a := range call.Args {
		arg := []byte(a)
		if utils.HasPrefix(arg, COMPLETE_PREFIX_KEY) {
			call, err := extractComponents(utils.Ptr(Origin{Start: 0, End: 0, FullMatch: arg, Match: arg}))
			if err != nil {
				return nil, CreateError(call, err.Error())
			}
			res, err := call.Exec(arg, tree)
			if err != nil {
				return nil, CreateError(call, err.Error())
			}
			arg = res
		}
		stack = append(stack, a)
	}
	return stack, nil
}
