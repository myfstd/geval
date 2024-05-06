package geval

import "github.com/myfstd/geval/core"

func Eval(expression string) interface{} {
	evalExpression, err := core.TNewEvaluableExpression(expression)
	if err != nil {
		return false
	}
	evaluate, err := evalExpression.TEvaluate(nil)
	if evaluate == nil {
		return false
	}
	return evaluate
}
