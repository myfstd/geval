package core

import (
	"errors"
	"fmt"
)

type lexerState struct {
	isEOF          bool
	isNullable     bool
	kind           tTokenKind
	validNextKinds []tTokenKind
}

// lexer states.
// Constant for all purposes except compiler.
var validLexerStates = []lexerState{

	lexerState{
		kind:       tUNKNOWN,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tBOOLEAN,
			tVARIABLE,
			tPATTERN,
			tFUNCTION,
			tACCESSOR,
			tSTRING,
			tTIME,
			tCLAUSE,
		},
	},

	lexerState{

		kind:       tCLAUSE,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tBOOLEAN,
			tVARIABLE,
			tPATTERN,
			tFUNCTION,
			tACCESSOR,
			tSTRING,
			tTIME,
			tCLAUSE,
			tCLAUSE_CLOSE,
		},
	},

	lexerState{

		kind:       tCLAUSE_CLOSE,
		isEOF:      true,
		isNullable: true,
		validNextKinds: []tTokenKind{

			tCOMPARATOR,
			tMODIFIER,
			tNUMERIC,
			tBOOLEAN,
			tVARIABLE,
			tSTRING,
			tPATTERN,
			tTIME,
			tCLAUSE,
			tCLAUSE_CLOSE,
			tLOGICALOP,
			tTERNARY,
			tSEPARATOR,
		},
	},

	lexerState{

		kind:       tNUMERIC,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tTERNARY,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tBOOLEAN,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tTERNARY,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tSTRING,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tTERNARY,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tTIME,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tPATTERN,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tVARIABLE,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tTERNARY,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tMODIFIER,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tVARIABLE,
			tFUNCTION,
			tACCESSOR,
			tSTRING,
			tBOOLEAN,
			tCLAUSE,
			tCLAUSE_CLOSE,
		},
	},
	lexerState{

		kind:       tCOMPARATOR,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tBOOLEAN,
			tVARIABLE,
			tFUNCTION,
			tACCESSOR,
			tSTRING,
			tTIME,
			tCLAUSE,
			tCLAUSE_CLOSE,
			tPATTERN,
		},
	},
	lexerState{

		kind:       tLOGICALOP,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tBOOLEAN,
			tVARIABLE,
			tFUNCTION,
			tACCESSOR,
			tSTRING,
			tTIME,
			tCLAUSE,
			tCLAUSE_CLOSE,
		},
	},
	lexerState{

		kind:       tPREFIX,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tNUMERIC,
			tBOOLEAN,
			tVARIABLE,
			tFUNCTION,
			tACCESSOR,
			tCLAUSE,
			tCLAUSE_CLOSE,
		},
	},

	lexerState{

		kind:       tTERNARY,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tBOOLEAN,
			tSTRING,
			tTIME,
			tVARIABLE,
			tFUNCTION,
			tACCESSOR,
			tCLAUSE,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tFUNCTION,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []tTokenKind{
			tCLAUSE,
		},
	},
	lexerState{

		kind:       tACCESSOR,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []tTokenKind{
			tCLAUSE,
			tMODIFIER,
			tCOMPARATOR,
			tLOGICALOP,
			tCLAUSE_CLOSE,
			tTERNARY,
			tSEPARATOR,
		},
	},
	lexerState{

		kind:       tSEPARATOR,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []tTokenKind{

			tPREFIX,
			tNUMERIC,
			tBOOLEAN,
			tSTRING,
			tTIME,
			tVARIABLE,
			tFUNCTION,
			tACCESSOR,
			tCLAUSE,
		},
	},
}

func (this lexerState) canTransitionTo(kind tTokenKind) bool {

	for _, validKind := range this.validNextKinds {

		if validKind == kind {
			return true
		}
	}

	return false
}

func checkExpressionSyntax(tokens []tExpressionToken) error {

	var state lexerState
	var lastToken tExpressionToken
	var err error

	state = validLexerStates[0]

	for _, token := range tokens {

		if !state.canTransitionTo(token.Kind) {

			// call out a specific error for tokens looking like they want to be functions.
			if lastToken.Kind == tVARIABLE && token.Kind == tCLAUSE {
				return errors.New("Undefined function " + lastToken.Value.(string))
			}

			firstStateName := fmt.Sprintf("%s [%v]", state.kind.tString(), lastToken.Value)
			nextStateName := fmt.Sprintf("%s [%v]", token.Kind.tString(), token.Value)

			return errors.New("Cannot transition token types from " + firstStateName + " to " + nextStateName)
		}

		state, err = getLexerStateForToken(token.Kind)
		if err != nil {
			return err
		}

		if !state.isNullable && token.Value == nil {

			errorMsg := fmt.Sprintf("Token kind '%v' cannot have a nil value", token.Kind.tString())
			return errors.New(errorMsg)
		}

		lastToken = token
	}

	if !state.isEOF {
		return errors.New("Unexpected end of expression")
	}
	return nil
}

func getLexerStateForToken(kind tTokenKind) (lexerState, error) {

	for _, possibleState := range validLexerStates {

		if possibleState.kind == kind {
			return possibleState, nil
		}
	}

	errorMsg := fmt.Sprintf("No lexer state found for token kind '%v'\n", kind.tString())
	return validLexerStates[0], errors.New(errorMsg)
}
