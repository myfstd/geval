package core

/*
Represents all valid types of tokens that a token can be.
*/
type tTokenKind int

const (
	tUNKNOWN tTokenKind = iota

	tPREFIX
	tNUMERIC
	tBOOLEAN
	tSTRING
	tPATTERN
	tTIME
	tVARIABLE
	tFUNCTION
	tSEPARATOR
	tACCESSOR

	tCOMPARATOR
	tLOGICALOP
	tMODIFIER

	tCLAUSE
	tCLAUSE_CLOSE

	tTERNARY
)

/*
GetTokenKindString returns a string that describes the given tTokenKind.
e.g., when passed the tNUMERIC tTokenKind, this returns the string "tNUMERIC".
*/
func (kind tTokenKind) tString() string {

	switch kind {

	case tPREFIX:
		return "tPREFIX"
	case tNUMERIC:
		return "tNUMERIC"
	case tBOOLEAN:
		return "tBOOLEAN"
	case tSTRING:
		return "tSTRING"
	case tPATTERN:
		return "tPATTERN"
	case tTIME:
		return "tTIME"
	case tVARIABLE:
		return "tVARIABLE"
	case tFUNCTION:
		return "tFUNCTION"
	case tSEPARATOR:
		return "tSEPARATOR"
	case tCOMPARATOR:
		return "tCOMPARATOR"
	case tLOGICALOP:
		return "tLOGICALOP"
	case tMODIFIER:
		return "tMODIFIER"
	case tCLAUSE:
		return "tCLAUSE"
	case tCLAUSE_CLOSE:
		return "tCLAUSE_CLOSE"
	case tTERNARY:
		return "tTERNARY"
	case tACCESSOR:
		return "tACCESSOR"
	}

	return "tUNKNOWN"
}
