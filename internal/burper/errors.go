package burper

import (
	"burp/pkg/utils"
	"errors"
	"strings"
)

var MalformedBurpCall = errors.New("malformed call to burp")

func mergeArgumentType(args [][]string) []string {
	return utils.Map(args, func(v []string) string {
		return strings.Join(v, ": ")
	})
}

func CreateMissingArgumentError(call *Call, args [][]string) error {
	return CreateError(call, "missing argument ["+strings.Join(mergeArgumentType(args), ",")+"]")
}

func CreateError(call *Call, error string) error {
	return errors.New(error + " in " + call.Function + " call on \"" + string(call.Source.Match) + "\"")
}
