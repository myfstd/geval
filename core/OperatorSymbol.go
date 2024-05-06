package core

/*
Represents the valid symbols for operators.
*/
type tOperatorSymbol int

const (
	tVALUE tOperatorSymbol = iota
	tLITERAL
	tNOOP
	tEQ
	tNEQ
	tGT
	tLT
	tGTE
	tLTE
	tREQ
	tNREQ
	tIN

	tAND
	tOR

	tPLUS
	tMINUS
	tBITWISE_AND
	tBITWISE_OR
	tBITWISE_XOR
	tBITWISE_LSHIFT
	tBITWISE_RSHIFT
	tMULTIPLY
	tDIVIDE
	tMODULUS
	tEXPONENT

	tNEGATE
	tINVERT
	tBITWISE_NOT

	tTERNARY_TRUE
	tTERNARY_FALSE
	tCOALESCE

	tFUNCTIONAL
	tACCESS
	tSEPARATE
)

type operatorPrecedence int

const (
	noopPrecedence operatorPrecedence = iota
	valuePrecedence
	functionalPrecedence
	prefixPrecedence
	exponentialPrecedence
	additivePrecedence
	bitwisePrecedence
	bitwiseShiftPrecedence
	multiplicativePrecedence
	comparatorPrecedence
	ternaryPrecedence
	logicalAndPrecedence
	logicalOrPrecedence
	separatePrecedence
)

func findOperatorPrecedenceForSymbol(symbol tOperatorSymbol) operatorPrecedence {

	switch symbol {
	case tNOOP:
		return noopPrecedence
	case tVALUE:
		return valuePrecedence
	case tEQ:
		fallthrough
	case tNEQ:
		fallthrough
	case tGT:
		fallthrough
	case tLT:
		fallthrough
	case tGTE:
		fallthrough
	case tLTE:
		fallthrough
	case tREQ:
		fallthrough
	case tNREQ:
		fallthrough
	case tIN:
		return comparatorPrecedence
	case tAND:
		return logicalAndPrecedence
	case tOR:
		return logicalOrPrecedence
	case tBITWISE_AND:
		fallthrough
	case tBITWISE_OR:
		fallthrough
	case tBITWISE_XOR:
		return bitwisePrecedence
	case tBITWISE_LSHIFT:
		fallthrough
	case tBITWISE_RSHIFT:
		return bitwiseShiftPrecedence
	case tPLUS:
		fallthrough
	case tMINUS:
		return additivePrecedence
	case tMULTIPLY:
		fallthrough
	case tDIVIDE:
		fallthrough
	case tMODULUS:
		return multiplicativePrecedence
	case tEXPONENT:
		return exponentialPrecedence
	case tBITWISE_NOT:
		fallthrough
	case tNEGATE:
		fallthrough
	case tINVERT:
		return prefixPrecedence
	case tCOALESCE:
		fallthrough
	case tTERNARY_TRUE:
		fallthrough
	case tTERNARY_FALSE:
		return ternaryPrecedence
	case tACCESS:
		fallthrough
	case tFUNCTIONAL:
		return functionalPrecedence
	case tSEPARATE:
		return separatePrecedence
	}

	return valuePrecedence
}

/*
Map of all valid comparators, and their string equivalents.
Used during parsing of expressions to determine if a symbol is, in fact, a comparator.
Also used during evaluation to determine exactly which comparator is being used.
*/
var comparatorSymbols = map[string]tOperatorSymbol{
	"==": tEQ,
	"!=": tNEQ,
	">":  tGT,
	">=": tGTE,
	"<":  tLT,
	"<=": tLTE,
	"=~": tREQ,
	"!~": tNREQ,
	"in": tIN,
}

var logicalSymbols = map[string]tOperatorSymbol{
	"&&": tAND,
	"||": tOR,
}

var bitwiseSymbols = map[string]tOperatorSymbol{
	"^": tBITWISE_XOR,
	"&": tBITWISE_AND,
	"|": tBITWISE_OR,
}

var bitwiseShiftSymbols = map[string]tOperatorSymbol{
	">>": tBITWISE_RSHIFT,
	"<<": tBITWISE_LSHIFT,
}

var additiveSymbols = map[string]tOperatorSymbol{
	"+": tPLUS,
	"-": tMINUS,
}

var multiplicativeSymbols = map[string]tOperatorSymbol{
	"*": tMULTIPLY,
	"/": tDIVIDE,
	"%": tMODULUS,
}

var exponentialSymbolsS = map[string]tOperatorSymbol{
	"**": tEXPONENT,
}

var prefixSymbols = map[string]tOperatorSymbol{
	"-": tNEGATE,
	"!": tINVERT,
	"~": tBITWISE_NOT,
}

var ternarySymbols = map[string]tOperatorSymbol{
	"?":  tTERNARY_TRUE,
	":":  tTERNARY_FALSE,
	"??": tCOALESCE,
}

// this is defined separately from additiveSymbols et al because it's needed for parsing, not stage planning.
var modifierSymbols = map[string]tOperatorSymbol{
	"+":  tPLUS,
	"-":  tMINUS,
	"*":  tMULTIPLY,
	"/":  tDIVIDE,
	"%":  tMODULUS,
	"**": tEXPONENT,
	"&":  tBITWISE_AND,
	"|":  tBITWISE_OR,
	"^":  tBITWISE_XOR,
	">>": tBITWISE_RSHIFT,
	"<<": tBITWISE_LSHIFT,
}

var separatorSymbols = map[string]tOperatorSymbol{
	",": tSEPARATE,
}

/*
Generally used when formatting type check errors.
We could store the stringified symbol somewhere else and not require a duplicated codeblock to translate
tOperatorSymbol to string, but that would require more memory, and another field somewhere.
Adding operators is rare enough that we just stringify it here instead.
*/
func (this tOperatorSymbol) String() string {

	switch this {
	case tNOOP:
		return "tNOOP"
	case tVALUE:
		return "tVALUE"
	case tEQ:
		return "="
	case tNEQ:
		return "!="
	case tGT:
		return ">"
	case tLT:
		return "<"
	case tGTE:
		return ">="
	case tLTE:
		return "<="
	case tREQ:
		return "=~"
	case tNREQ:
		return "!~"
	case tAND:
		return "&&"
	case tOR:
		return "||"
	case tIN:
		return "in"
	case tBITWISE_AND:
		return "&"
	case tBITWISE_OR:
		return "|"
	case tBITWISE_XOR:
		return "^"
	case tBITWISE_LSHIFT:
		return "<<"
	case tBITWISE_RSHIFT:
		return ">>"
	case tPLUS:
		return "+"
	case tMINUS:
		return "-"
	case tMULTIPLY:
		return "*"
	case tDIVIDE:
		return "/"
	case tMODULUS:
		return "%"
	case tEXPONENT:
		return "**"
	case tNEGATE:
		return "-"
	case tINVERT:
		return "!"
	case tBITWISE_NOT:
		return "~"
	case tTERNARY_TRUE:
		return "?"
	case tTERNARY_FALSE:
		return ":"
	case tCOALESCE:
		return "??"
	}
	return ""
}
