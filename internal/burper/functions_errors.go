package burper

import (
	"burp/pkg/utils"
	"errors"
	"strings"
)

var ErrMalformedBurpCall = errors.New("malformed call to burp")

func mergeArgumentType(args [][]string) []string {
	return utils.Map(args, func(v []string) string {
		return strings.Join(v, ": ")
	})
}

func (call *FunctionCall) MissingArgumentErr(argsType ...string) error {
	return call.Err("missing argument [" + strings.Join(mergeArgumentType(utils.Array(argsType...)), ",") + "]")
}

func (call *FunctionCall) Err(message string) error {
	return errors.New(message + " in " + call.Function + " call on \"" + string(call.Source.Match) + "\"")
}

func (call *FunctionCall) FormatErr(err error) error {
	return call.Err(err.Error())
}
