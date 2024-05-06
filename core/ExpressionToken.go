package core

/*
Represents a single parsed token.
*/
type tExpressionToken struct {
	Kind  tTokenKind
	Value interface{}
}
