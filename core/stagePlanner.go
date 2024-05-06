package core

import (
	"errors"
	"fmt"
	"time"
)

var stageSymbolMap = map[tOperatorSymbol]evaluationOperator{
	tEQ:             equalStage,
	tNEQ:            notEqualStage,
	tGT:             gtStage,
	tLT:             ltStage,
	tGTE:            gteStage,
	tLTE:            lteStage,
	tREQ:            regexStage,
	tNREQ:           notRegexStage,
	tAND:            andStage,
	tOR:             orStage,
	tIN:             inStage,
	tBITWISE_OR:     bitwiseOrStage,
	tBITWISE_AND:    bitwiseAndStage,
	tBITWISE_XOR:    bitwiseXORStage,
	tBITWISE_LSHIFT: leftShiftStage,
	tBITWISE_RSHIFT: rightShiftStage,
	tPLUS:           addStage,
	tMINUS:          subtractStage,
	tMULTIPLY:       multiplyStage,
	tDIVIDE:         divideStage,
	tMODULUS:        modulusStage,
	tEXPONENT:       exponentStage,
	tNEGATE:         negateStage,
	tINVERT:         invertStage,
	tBITWISE_NOT:    bitwiseNotStage,
	tTERNARY_TRUE:   ternaryIfStage,
	tTERNARY_FALSE:  ternaryElseStage,
	tCOALESCE:       ternaryElseStage,
	tSEPARATE:       separatorStage,
}

/*
A "precedent" is a function which will recursively parse new evaluateionStages from a given stream of tokens.
It's called a `precedent` because it is expected to handle exactly what precedence of operator,
and defer to other `precedent`s for other operators.
*/
type precedent func(stream *tokenStream) (*evaluationStage, error)

/*
A convenience function for specifying the behavior of a `precedent`.
Most `precedent` functions can be described by the same function, just with different type checks, symbols, and error formats.
This struct is passed to `makePrecedentFromPlanner` to create a `precedent` function.
*/
type precedencePlanner struct {
	validSymbols map[string]tOperatorSymbol
	validKinds   []tTokenKind

	typeErrorFormat string

	next      precedent
	nextRight precedent
}

var planPrefix precedent
var planExponential precedent
var planMultiplicative precedent
var planAdditive precedent
var planBitwise precedent
var planShift precedent
var planComparator precedent
var planLogicalAnd precedent
var planLogicalOr precedent
var planTernary precedent
var planSeparator precedent

func init() {

	// all these stages can use the same code (in `planPrecedenceLevel`) to execute,
	// they simply need different type checks, symbols, and recursive precedents.
	// While not all precedent phases are listed here, most can be represented this way.
	planPrefix = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    prefixSymbols,
		validKinds:      []tTokenKind{tPREFIX},
		typeErrorFormat: prefixErrorFormat,
		nextRight:       planFunction,
	})
	planExponential = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    exponentialSymbolsS,
		validKinds:      []tTokenKind{tMODIFIER},
		typeErrorFormat: modifierErrorFormat,
		next:            planFunction,
	})
	planMultiplicative = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    multiplicativeSymbols,
		validKinds:      []tTokenKind{tMODIFIER},
		typeErrorFormat: modifierErrorFormat,
		next:            planExponential,
	})
	planAdditive = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    additiveSymbols,
		validKinds:      []tTokenKind{tMODIFIER},
		typeErrorFormat: modifierErrorFormat,
		next:            planMultiplicative,
	})
	planShift = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    bitwiseShiftSymbols,
		validKinds:      []tTokenKind{tMODIFIER},
		typeErrorFormat: modifierErrorFormat,
		next:            planAdditive,
	})
	planBitwise = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    bitwiseSymbols,
		validKinds:      []tTokenKind{tMODIFIER},
		typeErrorFormat: modifierErrorFormat,
		next:            planShift,
	})
	planComparator = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    comparatorSymbols,
		validKinds:      []tTokenKind{tCOMPARATOR},
		typeErrorFormat: comparatorErrorFormat,
		next:            planBitwise,
	})
	planLogicalAnd = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    map[string]tOperatorSymbol{"&&": tAND},
		validKinds:      []tTokenKind{tLOGICALOP},
		typeErrorFormat: logicalErrorFormat,
		next:            planComparator,
	})
	planLogicalOr = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    map[string]tOperatorSymbol{"||": tOR},
		validKinds:      []tTokenKind{tLOGICALOP},
		typeErrorFormat: logicalErrorFormat,
		next:            planLogicalAnd,
	})
	planTernary = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    ternarySymbols,
		validKinds:      []tTokenKind{tTERNARY},
		typeErrorFormat: ternaryErrorFormat,
		next:            planLogicalOr,
	})
	planSeparator = makePrecedentFromPlanner(&precedencePlanner{
		validSymbols: separatorSymbols,
		validKinds:   []tTokenKind{tSEPARATOR},
		next:         planTernary,
	})
}

/*
Given a planner, creates a function which will evaluate a specific precedence level of operators,
and link it to other `precedent`s which recurse to parse other precedence levels.
*/
func makePrecedentFromPlanner(planner *precedencePlanner) precedent {

	var generated precedent
	var nextRight precedent

	generated = func(stream *tokenStream) (*evaluationStage, error) {
		return planPrecedenceLevel(
			stream,
			planner.typeErrorFormat,
			planner.validSymbols,
			planner.validKinds,
			nextRight,
			planner.next,
		)
	}

	if planner.nextRight != nil {
		nextRight = planner.nextRight
	} else {
		nextRight = generated
	}

	return generated
}

/*
Creates a `evaluationStageList` object which represents an execution plan (or tree)
which is used to completely evaluate a set of tokens at evaluation-time.
The three stages of evaluation can be thought of as parsing strings to tokens, then tokens to a stage list, then evaluation with parameters.
*/
func planStages(tokens []tExpressionToken) (*evaluationStage, error) {

	stream := newTokenStream(tokens)

	stage, err := planTokens(stream)
	if err != nil {
		return nil, err
	}

	// while we're now fully-planned, we now need to re-order same-precedence operators.
	// this could probably be avoided with a different planning method
	reorderStages(stage)

	stage = elideLiterals(stage)
	return stage, nil
}

func planTokens(stream *tokenStream) (*evaluationStage, error) {

	if !stream.hasNext() {
		return nil, nil
	}

	return planSeparator(stream)
}

/*
The most usual method of parsing an evaluation stage for a given precedence.
Most stages use the same logic
*/
func planPrecedenceLevel(
	stream *tokenStream,
	typeErrorFormat string,
	validSymbols map[string]tOperatorSymbol,
	validKinds []tTokenKind,
	rightPrecedent precedent,
	leftPrecedent precedent) (*evaluationStage, error) {

	var token tExpressionToken
	var symbol tOperatorSymbol
	var leftStage, rightStage *evaluationStage
	var checks typeChecks
	var err error
	var keyFound bool

	if leftPrecedent != nil {

		leftStage, err = leftPrecedent(stream)
		if err != nil {
			return nil, err
		}
	}

	for stream.hasNext() {

		token = stream.next()

		if len(validKinds) > 0 {

			keyFound = false
			for _, kind := range validKinds {
				if kind == token.Kind {
					keyFound = true
					break
				}
			}

			if !keyFound {
				break
			}
		}

		if validSymbols != nil {

			if !isString(token.Value) {
				break
			}

			symbol, keyFound = validSymbols[token.Value.(string)]
			if !keyFound {
				break
			}
		}

		if rightPrecedent != nil {
			rightStage, err = rightPrecedent(stream)
			if err != nil {
				return nil, err
			}
		}

		checks = findTypeChecks(symbol)

		return &evaluationStage{

			symbol:     symbol,
			leftStage:  leftStage,
			rightStage: rightStage,
			operator:   stageSymbolMap[symbol],

			leftTypeCheck:   checks.left,
			rightTypeCheck:  checks.right,
			typeCheck:       checks.combined,
			typeErrorFormat: typeErrorFormat,
		}, nil
	}

	stream.rewind()
	return leftStage, nil
}

/*
A special case where functions need to be of higher precedence than values, and need a special wrapped execution stage operator.
*/
func planFunction(stream *tokenStream) (*evaluationStage, error) {

	var token tExpressionToken
	var rightStage *evaluationStage
	var err error

	token = stream.next()

	if token.Kind != tFUNCTION {
		stream.rewind()
		return planAccessor(stream)
	}

	rightStage, err = planAccessor(stream)
	if err != nil {
		return nil, err
	}

	return &evaluationStage{

		symbol:          tFUNCTIONAL,
		rightStage:      rightStage,
		operator:        makeFunctionStage(token.Value.(tExpressionFunction)),
		typeErrorFormat: "Unable to run function '%v': %v",
	}, nil
}

func planAccessor(stream *tokenStream) (*evaluationStage, error) {

	var token, otherToken tExpressionToken
	var rightStage *evaluationStage
	var err error

	if !stream.hasNext() {
		return nil, nil
	}

	token = stream.next()

	if token.Kind != tACCESSOR {
		stream.rewind()
		return planValue(stream)
	}

	// check if this is meant to be a function or a field.
	// fields have a clause next to them, functions do not.
	// if it's a function, parse the arguments. Otherwise leave the right stage null.
	if stream.hasNext() {

		otherToken = stream.next()
		if otherToken.Kind == tCLAUSE {

			stream.rewind()

			rightStage, err = planTokens(stream)
			if err != nil {
				return nil, err
			}
		} else {
			stream.rewind()
		}
	}

	return &evaluationStage{

		symbol:          tACCESS,
		rightStage:      rightStage,
		operator:        makeAccessorStage(token.Value.([]string)),
		typeErrorFormat: "Unable to access parameter field or method '%v': %v",
	}, nil
}

/*
A truly special precedence function, this handles all the "lowest-case" errata of the process, including literals, parmeters,
clauses, and prefixes.
*/
func planValue(stream *tokenStream) (*evaluationStage, error) {

	var token tExpressionToken
	var symbol tOperatorSymbol
	var ret *evaluationStage
	var operator evaluationOperator
	var err error

	if !stream.hasNext() {
		return nil, nil
	}

	token = stream.next()

	switch token.Kind {

	case tCLAUSE:

		ret, err = planTokens(stream)
		if err != nil {
			return nil, err
		}
		stream.next()
		ret = &evaluationStage{
			rightStage: ret,
			operator:   noopStageRight,
			symbol:     tNOOP,
		}

		return ret, nil

	case tCLAUSE_CLOSE:
		stream.rewind()
		return nil, nil

	case tVARIABLE:
		operator = makeParameterStage(token.Value.(string))
	case tNUMERIC:
		fallthrough
	case tSTRING:
		fallthrough
	case tPATTERN:
		fallthrough
	case tBOOLEAN:
		symbol = tLITERAL
		operator = makeLiteralStage(token.Value)
	case tTIME:
		symbol = tLITERAL
		operator = makeLiteralStage(float64(token.Value.(time.Time).Unix()))

	case tPREFIX:
		stream.rewind()
		return planPrefix(stream)
	}

	if operator == nil {
		errorMsg := fmt.Sprintf("Unable to plan token kind: '%s', value: '%v'", token.Kind.tString(), token.Value)
		return nil, errors.New(errorMsg)
	}

	return &evaluationStage{
		symbol:   symbol,
		operator: operator,
	}, nil
}

/*
Convenience function to pass a triplet of typechecks between `findTypeChecks` and `planPrecedenceLevel`.
Each of these members may be nil, which indicates that type does not matter for that value.
*/
type typeChecks struct {
	left     stageTypeCheck
	right    stageTypeCheck
	combined stageCombinedTypeCheck
}

/*
Maps a given [symbol] to a set of typechecks to be used during runtime.
*/
func findTypeChecks(symbol tOperatorSymbol) typeChecks {

	switch symbol {
	case tGT:
		fallthrough
	case tLT:
		fallthrough
	case tGTE:
		fallthrough
	case tLTE:
		return typeChecks{
			combined: comparatorTypeCheck,
		}
	case tREQ:
		fallthrough
	case tNREQ:
		return typeChecks{
			left:  isString,
			right: isRegexOrString,
		}
	case tAND:
		fallthrough
	case tOR:
		return typeChecks{
			left:  isBool,
			right: isBool,
		}
	case tIN:
		return typeChecks{
			right: isArray,
		}
	case tBITWISE_LSHIFT:
		fallthrough
	case tBITWISE_RSHIFT:
		fallthrough
	case tBITWISE_OR:
		fallthrough
	case tBITWISE_AND:
		fallthrough
	case tBITWISE_XOR:
		return typeChecks{
			left:  isFloat64,
			right: isFloat64,
		}
	case tPLUS:
		return typeChecks{
			combined: additionTypeCheck,
		}
	case tMINUS:
		fallthrough
	case tMULTIPLY:
		fallthrough
	case tDIVIDE:
		fallthrough
	case tMODULUS:
		fallthrough
	case tEXPONENT:
		return typeChecks{
			left:  isFloat64,
			right: isFloat64,
		}
	case tNEGATE:
		return typeChecks{
			right: isFloat64,
		}
	case tINVERT:
		return typeChecks{
			right: isBool,
		}
	case tBITWISE_NOT:
		return typeChecks{
			right: isFloat64,
		}
	case tTERNARY_TRUE:
		return typeChecks{
			left: isBool,
		}

	// unchecked cases
	case tEQ:
		fallthrough
	case tNEQ:
		return typeChecks{}
	case tTERNARY_FALSE:
		fallthrough
	case tCOALESCE:
		fallthrough
	default:
		return typeChecks{}
	}
}

/*
During stage planning, stages of equal precedence are parsed such that they'll be evaluated in reverse order.
For commutative operators like "+" or "-", it's no big deal. But for order-specific operators, it ruins the expected result.
*/
func reorderStages(rootStage *evaluationStage) {

	// traverse every rightStage until we find multiples in a row of the same precedence.
	var identicalPrecedences []*evaluationStage
	var currentStage, nextStage *evaluationStage
	var precedence, currentPrecedence operatorPrecedence

	nextStage = rootStage
	precedence = findOperatorPrecedenceForSymbol(rootStage.symbol)

	for nextStage != nil {

		currentStage = nextStage
		nextStage = currentStage.rightStage

		// left depth first, since this entire method only looks for precedences down the right side of the tree
		if currentStage.leftStage != nil {
			reorderStages(currentStage.leftStage)
		}

		currentPrecedence = findOperatorPrecedenceForSymbol(currentStage.symbol)

		if currentPrecedence == precedence {
			identicalPrecedences = append(identicalPrecedences, currentStage)
			continue
		}

		// precedence break.
		// See how many in a row we had, and reorder if there's more than one.
		if len(identicalPrecedences) > 1 {
			mirrorStageSubtree(identicalPrecedences)
		}

		identicalPrecedences = []*evaluationStage{currentStage}
		precedence = currentPrecedence
	}

	if len(identicalPrecedences) > 1 {
		mirrorStageSubtree(identicalPrecedences)
	}
}

/*
Performs a "mirror" on a subtree of stages.
This mirror functionally inverts the order of execution for all members of the [stages] list.
That list is assumed to be a root-to-leaf (ordered) list of evaluation stages, where each is a right-hand stage of the last.
*/
func mirrorStageSubtree(stages []*evaluationStage) {

	var rootStage, inverseStage, carryStage, frontStage *evaluationStage

	stagesLength := len(stages)

	// reverse all right/left
	for _, frontStage = range stages {

		carryStage = frontStage.rightStage
		frontStage.rightStage = frontStage.leftStage
		frontStage.leftStage = carryStage
	}

	// end left swaps with root right
	rootStage = stages[0]
	frontStage = stages[stagesLength-1]

	carryStage = frontStage.leftStage
	frontStage.leftStage = rootStage.rightStage
	rootStage.rightStage = carryStage

	// for all non-root non-end stages, right is swapped with inverse stage right in list
	for i := 0; i < (stagesLength-2)/2+1; i++ {

		frontStage = stages[i+1]
		inverseStage = stages[stagesLength-i-1]

		carryStage = frontStage.rightStage
		frontStage.rightStage = inverseStage.rightStage
		inverseStage.rightStage = carryStage
	}

	// swap all other information with inverse stages
	for i := 0; i < stagesLength/2; i++ {

		frontStage = stages[i]
		inverseStage = stages[stagesLength-i-1]
		frontStage.swapWith(inverseStage)
	}
}

/*
Recurses through all operators in the entire tree, eliding operators where both sides are literals.
*/
func elideLiterals(root *evaluationStage) *evaluationStage {

	if root.leftStage != nil {
		root.leftStage = elideLiterals(root.leftStage)
	}

	if root.rightStage != nil {
		root.rightStage = elideLiterals(root.rightStage)
	}

	return elideStage(root)
}

/*
Elides a specific stage, if possible.
Returns the unmodified [root] stage if it cannot or should not be elided.
Otherwise, returns a new stage representing the condensed value from the elided stages.
*/
func elideStage(root *evaluationStage) *evaluationStage {

	var leftValue, rightValue, result interface{}
	var err error

	// right side must be a non-nil value. Left side must be nil or a value.
	if root.rightStage == nil ||
		root.rightStage.symbol != tLITERAL ||
		root.leftStage == nil ||
		root.leftStage.symbol != tLITERAL {
		return root
	}

	// don't elide some operators
	switch root.symbol {
	case tSEPARATE:
		fallthrough
	case tIN:
		return root
	}

	// both sides are values, get their actual values.
	// errors should be near-impossible here. If we encounter them, just abort this optimization.
	leftValue, err = root.leftStage.operator(nil, nil, nil)
	if err != nil {
		return root
	}

	rightValue, err = root.rightStage.operator(nil, nil, nil)
	if err != nil {
		return root
	}

	// typcheck, since the grammar checker is a bit loose with which operator symbols go together.
	err = typeCheck(root.leftTypeCheck, leftValue, root.symbol, root.typeErrorFormat)
	if err != nil {
		return root
	}

	err = typeCheck(root.rightTypeCheck, rightValue, root.symbol, root.typeErrorFormat)
	if err != nil {
		return root
	}

	if root.typeCheck != nil && !root.typeCheck(leftValue, rightValue) {
		return root
	}

	// pre-calculate, and return a new stage representing the result.
	result, err = root.operator(leftValue, rightValue, nil)
	if err != nil {
		return root
	}

	return &evaluationStage{
		symbol:   tLITERAL,
		operator: makeLiteralStage(result),
	}
}
