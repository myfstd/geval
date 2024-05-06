package core

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func parseTokens(expression string, functions map[string]tExpressionFunction) ([]tExpressionToken, error) {

	var ret []tExpressionToken
	var token tExpressionToken
	var stream *lexerStream
	var state lexerState
	var err error
	var found bool

	stream = newLexerStream(expression)
	state = validLexerStates[0]

	for stream.canRead() {

		token, err, found = readToken(stream, state, functions)

		if err != nil {
			return ret, err
		}

		if !found {
			break
		}

		state, err = getLexerStateForToken(token.Kind)
		if err != nil {
			return ret, err
		}

		// append this valid token
		ret = append(ret, token)
	}

	err = checkBalance(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func readToken(stream *lexerStream, state lexerState, functions map[string]tExpressionFunction) (tExpressionToken, error, bool) {

	var function tExpressionFunction
	var ret tExpressionToken
	var tokenValue interface{}
	var tokenTime time.Time
	var tokenString string
	var kind tTokenKind
	var character rune
	var found bool
	var completed bool
	var err error

	// numeric is 0-9, or . or 0x followed by digits
	// string starts with '
	// variable is alphanumeric, always starts with a letter
	// bracket always means variable
	// symbols are anything non-alphanumeric
	// all others read into a buffer until they reach the end of the stream
	for stream.canRead() {

		character = stream.readCharacter()

		if unicode.IsSpace(character) {
			continue
		}

		kind = tUNKNOWN

		// numeric constant
		if isNumeric(character) {

			if stream.canRead() && character == '0' {
				character = stream.readCharacter()

				if stream.canRead() && character == 'x' {
					tokenString, _ = readUntilFalse(stream, false, true, true, isHexDigit)
					tokenValueInt, err := strconv.ParseUint(tokenString, 16, 64)

					if err != nil {
						errorMsg := fmt.Sprintf("Unable to parse hex value '%v' to uint64\n", tokenString)
						return tExpressionToken{}, errors.New(errorMsg), false
					}

					kind = tNUMERIC
					tokenValue = float64(tokenValueInt)
					break
				} else {
					stream.rewind(1)
				}
			}

			tokenString = readTokenUntilFalse(stream, isNumeric)
			tokenValue, err = strconv.ParseFloat(tokenString, 64)

			if err != nil {
				errorMsg := fmt.Sprintf("Unable to parse numeric value '%v' to float64\n", tokenString)
				return tExpressionToken{}, errors.New(errorMsg), false
			}
			kind = tNUMERIC
			break
		}

		// comma, separator
		if character == ',' {

			tokenValue = ","
			kind = tSEPARATOR
			break
		}

		// escaped variable
		if character == '[' {

			tokenValue, completed = readUntilFalse(stream, true, false, true, isNotClosingBracket)
			kind = tVARIABLE

			if !completed {
				return tExpressionToken{}, errors.New("Unclosed parameter bracket"), false
			}

			// above method normally rewinds us to the closing bracket, which we want to skip.
			stream.rewind(-1)
			break
		}

		// regular variable - or function?
		if unicode.IsLetter(character) {

			tokenString = readTokenUntilFalse(stream, isVariableName)

			tokenValue = tokenString
			kind = tVARIABLE

			// boolean?
			if tokenValue == "true" {

				kind = tBOOLEAN
				tokenValue = true
			} else {

				if tokenValue == "false" {

					kind = tBOOLEAN
					tokenValue = false
				}
			}

			// textual operator?
			if tokenValue == "in" || tokenValue == "tIN" {

				// force lower case for consistency
				tokenValue = "in"
				kind = tCOMPARATOR
			}

			// function?
			function, found = functions[tokenString]
			if found {
				kind = tFUNCTION
				tokenValue = function
			}

			// accessor?
			accessorIndex := strings.Index(tokenString, ".")
			if accessorIndex > 0 {

				// check that it doesn't end with a hanging period
				if tokenString[len(tokenString)-1] == '.' {
					errorMsg := fmt.Sprintf("Hanging accessor on token '%s'", tokenString)
					return tExpressionToken{}, errors.New(errorMsg), false
				}

				kind = tACCESSOR
				splits := strings.Split(tokenString, ".")
				tokenValue = splits

				// check that none of them are unexported
				for i := 1; i < len(splits); i++ {

					firstCharacter := getFirstRune(splits[i])

					if unicode.ToUpper(firstCharacter) != firstCharacter {
						errorMsg := fmt.Sprintf("Unable to access unexported field '%s' in token '%s'", splits[i], tokenString)
						return tExpressionToken{}, errors.New(errorMsg), false
					}
				}
			}
			break
		}

		if !isNotQuote(character) {
			tokenValue, completed = readUntilFalse(stream, true, false, true, isNotQuote)

			if !completed {
				return tExpressionToken{}, errors.New("Unclosed string literal"), false
			}

			// advance the stream one position, since reading until false assumes the terminator is a real token
			stream.rewind(-1)

			// check to see if this can be parsed as a time.
			tokenTime, found = tryParseTime(tokenValue.(string))
			if found {
				kind = tTIME
				tokenValue = tokenTime
			} else {
				kind = tSTRING
			}
			break
		}

		if character == '(' {
			tokenValue = character
			kind = tCLAUSE
			break
		}

		if character == ')' {
			tokenValue = character
			kind = tCLAUSE_CLOSE
			break
		}

		// must be a known symbol
		tokenString = readTokenUntilFalse(stream, isNotAlphanumeric)
		tokenValue = tokenString

		// quick hack for the case where "-" can mean "prefixed negation" or "minus", which are used
		// very differently.
		if state.canTransitionTo(tPREFIX) {
			_, found = prefixSymbols[tokenString]
			if found {

				kind = tPREFIX
				break
			}
		}
		_, found = modifierSymbols[tokenString]
		if found {

			kind = tMODIFIER
			break
		}

		_, found = logicalSymbols[tokenString]
		if found {

			kind = tLOGICALOP
			break
		}

		_, found = comparatorSymbols[tokenString]
		if found {

			kind = tCOMPARATOR
			break
		}

		_, found = ternarySymbols[tokenString]
		if found {

			kind = tTERNARY
			break
		}

		errorMessage := fmt.Sprintf("Invalid token: '%s'", tokenString)
		return ret, errors.New(errorMessage), false
	}

	ret.Kind = kind
	ret.Value = tokenValue

	return ret, nil, (kind != tUNKNOWN)
}

func readTokenUntilFalse(stream *lexerStream, condition func(rune) bool) string {

	var ret string

	stream.rewind(1)
	ret, _ = readUntilFalse(stream, false, true, true, condition)
	return ret
}

/*
Returns the string that was read until the given [condition] was false, or whitespace was broken.
Returns false if the stream ended before whitespace was broken or condition was met.
*/
func readUntilFalse(stream *lexerStream, includeWhitespace bool, breakWhitespace bool, allowEscaping bool, condition func(rune) bool) (string, bool) {

	var tokenBuffer bytes.Buffer
	var character rune
	var conditioned bool

	conditioned = false

	for stream.canRead() {

		character = stream.readCharacter()

		// Use backslashes to escape anything
		if allowEscaping && character == '\\' {

			character = stream.readCharacter()
			tokenBuffer.WriteString(string(character))
			continue
		}

		if unicode.IsSpace(character) {

			if breakWhitespace && tokenBuffer.Len() > 0 {
				conditioned = true
				break
			}
			if !includeWhitespace {
				continue
			}
		}

		if condition(character) {
			tokenBuffer.WriteString(string(character))
		} else {
			conditioned = true
			stream.rewind(1)
			break
		}
	}

	return tokenBuffer.String(), conditioned
}

/*
Checks to see if any optimizations can be performed on the given [tokens], which form a complete, valid expression.
The returns slice will represent the optimized (or unmodified) list of tokens to use.
*/
func optimizeTokens(tokens []tExpressionToken) ([]tExpressionToken, error) {

	var token tExpressionToken
	var symbol tOperatorSymbol
	var err error
	var index int

	for index, token = range tokens {

		// if we find a regex operator, and the right-hand value is a constant, precompile and replace with a pattern.
		if token.Kind != tCOMPARATOR {
			continue
		}

		symbol = comparatorSymbols[token.Value.(string)]
		if symbol != tREQ && symbol != tNREQ {
			continue
		}

		index++
		token = tokens[index]
		if token.Kind == tSTRING {

			token.Kind = tPATTERN
			token.Value, err = regexp.Compile(token.Value.(string))

			if err != nil {
				return tokens, err
			}

			tokens[index] = token
		}
	}
	return tokens, nil
}

/*
Checks the balance of tokens which have multiple parts, such as parenthesis.
*/
func checkBalance(tokens []tExpressionToken) error {

	var stream *tokenStream
	var token tExpressionToken
	var parens int

	stream = newTokenStream(tokens)

	for stream.hasNext() {

		token = stream.next()
		if token.Kind == tCLAUSE {
			parens++
			continue
		}
		if token.Kind == tCLAUSE_CLOSE {
			parens--
			continue
		}
	}

	if parens != 0 {
		return errors.New("Unbalanced parenthesis")
	}
	return nil
}

func isDigit(character rune) bool {
	return unicode.IsDigit(character)
}

func isHexDigit(character rune) bool {

	character = unicode.ToLower(character)

	return unicode.IsDigit(character) ||
		character == 'a' ||
		character == 'b' ||
		character == 'c' ||
		character == 'd' ||
		character == 'e' ||
		character == 'f'
}

func isNumeric(character rune) bool {

	return unicode.IsDigit(character) || character == '.'
}

func isNotQuote(character rune) bool {

	return character != '\'' && character != '"'
}

func isNotAlphanumeric(character rune) bool {

	return !(unicode.IsDigit(character) ||
		unicode.IsLetter(character) ||
		character == '(' ||
		character == ')' ||
		character == '[' ||
		character == ']' || // starting to feel like there needs to be an `isOperation` func (#59)
		!isNotQuote(character))
}

func isVariableName(character rune) bool {

	return unicode.IsLetter(character) ||
		unicode.IsDigit(character) ||
		character == '_' ||
		character == '.'
}

func isNotClosingBracket(character rune) bool {

	return character != ']'
}

/*
Attempts to parse the [candidate] as a Time.
Tries a series of standardized date formats, returns the Time if one applies,
otherwise returns false through the second return.
*/
func tryParseTime(candidate string) (time.Time, bool) {

	var ret time.Time
	var found bool

	timeFormats := [...]string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.Kitchen,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",                         // RFC 3339
		"2006-01-02 15:04",                   // RFC 3339 with minutes
		"2006-01-02 15:04:05",                // RFC 3339 with seconds
		"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
		"2006-01-02T15Z0700",                 // ISO8601 with hour
		"2006-01-02T15:04Z0700",              // ISO8601 with minutes
		"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
		"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
	}

	for _, format := range timeFormats {

		ret, found = tryParseExactTime(candidate, format)
		if found {
			return ret, true
		}
	}

	return time.Now(), false
}

func tryParseExactTime(candidate string, format string) (time.Time, bool) {

	var ret time.Time
	var err error

	ret, err = time.ParseInLocation(format, candidate, time.Local)
	if err != nil {
		return time.Now(), false
	}

	return ret, true
}

func getFirstRune(candidate string) rune {

	for _, character := range candidate {
		return character
	}

	return 0
}
