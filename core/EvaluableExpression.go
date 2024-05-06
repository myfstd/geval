package core

import (
	"errors"
	"fmt"
)

const isoDateFormat string = "2006-01-02T15:04:05.999999999Z0700"
const shortCircuitHolder int = -1

var tDUMMY_PARAMETERS = tMapParameters(map[string]interface{}{})

type tEvaluableExpression struct {
	QueryDateFormat  string
	ChecksTypes      bool
	tokens           []tExpressionToken
	evaluationStages *evaluationStage
	inputExpression  string
}

func TNewEvaluableExpression(expression string) (*tEvaluableExpression, error) {
	functions := make(map[string]tExpressionFunction)
	return tNewEvaluableExpressionWithFunctions(expression, functions)
}

func tNewEvaluableExpressionWithFunctions(expression string, functions map[string]tExpressionFunction) (*tEvaluableExpression, error) {
	var ret *tEvaluableExpression
	var err error
	ret = new(tEvaluableExpression)
	ret.QueryDateFormat = isoDateFormat
	ret.inputExpression = expression
	ret.tokens, err = parseTokens(expression, functions)
	if err != nil {
		return nil, err
	}

	err = checkBalance(ret.tokens)
	if err != nil {
		return nil, err
	}
	err = checkExpressionSyntax(ret.tokens)
	if err != nil {
		return nil, err
	}
	ret.tokens, err = optimizeTokens(ret.tokens)
	if err != nil {
		return nil, err
	}

	ret.evaluationStages, err = planStages(ret.tokens)
	if err != nil {
		return nil, err
	}

	ret.ChecksTypes = true
	return ret, nil
}

func (t tEvaluableExpression) TEvaluate(parameters map[string]interface{}) (interface{}, error) {

	if parameters == nil {
		return t.tEval(nil)
	}

	return t.tEval(tMapParameters(parameters))
}

func (t tEvaluableExpression) tEval(parameters tParameters) (interface{}, error) {

	if t.evaluationStages == nil {
		return nil, nil
	}

	if parameters != nil {
		parameters = &sanitizedParameters{parameters}
	} else {
		parameters = tDUMMY_PARAMETERS
	}

	return t.evaluateStage(t.evaluationStages, parameters)
}

func (t tEvaluableExpression) evaluateStage(stage *evaluationStage, parameters tParameters) (interface{}, error) {

	var left, right interface{}
	var err error

	if stage.leftStage != nil {
		left, err = t.evaluateStage(stage.leftStage, parameters)
		if err != nil {
			return nil, err
		}
	}

	if stage.isShortCircuitable() {
		switch stage.symbol {
		case tAND:
			if left == false {
				return false, nil
			}
		case tOR:
			if left == true {
				return true, nil
			}
		case tCOALESCE:
			if left != nil {
				return left, nil
			}

		case tTERNARY_TRUE:
			if left == false {
				right = shortCircuitHolder
			}
		case tTERNARY_FALSE:
			if left != nil {
				right = shortCircuitHolder
			}
		}
	}

	if right != shortCircuitHolder && stage.rightStage != nil {
		right, err = t.evaluateStage(stage.rightStage, parameters)
		if err != nil {
			return nil, err
		}
	}

	if t.ChecksTypes {
		if stage.typeCheck == nil {

			err = typeCheck(stage.leftTypeCheck, left, stage.symbol, stage.typeErrorFormat)
			if err != nil {
				return nil, err
			}

			err = typeCheck(stage.rightTypeCheck, right, stage.symbol, stage.typeErrorFormat)
			if err != nil {
				return nil, err
			}
		} else {
			// special case where the type check needs to know both sides to determine if the operator can handle it
			if !stage.typeCheck(left, right) {
				errorMsg := fmt.Sprintf(stage.typeErrorFormat, left, stage.symbol.String())
				return nil, errors.New(errorMsg)
			}
		}
	}

	return stage.operator(left, right, parameters)
}

func typeCheck(check stageTypeCheck, value interface{}, symbol tOperatorSymbol, format string) error {

	if check == nil {
		return nil
	}

	if check(value) {
		return nil
	}

	errorMsg := fmt.Sprintf(format, value, symbol.String())
	return errors.New(errorMsg)
}
