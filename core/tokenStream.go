package core

type tokenStream struct {
	tokens      []tExpressionToken
	index       int
	tokenLength int
}

func newTokenStream(tokens []tExpressionToken) *tokenStream {

	var ret *tokenStream

	ret = new(tokenStream)
	ret.tokens = tokens
	ret.tokenLength = len(tokens)
	return ret
}

func (t *tokenStream) rewind() {
	t.index -= 1
}

func (t *tokenStream) next() tExpressionToken {

	var token tExpressionToken

	token = t.tokens[t.index]

	t.index += 1
	return token
}

func (t tokenStream) hasNext() bool {

	return t.index < t.tokenLength
}
