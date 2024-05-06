package core

import (
	"errors"
)

/*
tParameters is a collection of named parameters that can be used by an tEvaluableExpression to retrieve parameters
when an expression tries to use them.
*/
type tParameters interface {

	/*
		Get gets the parameter of the given name, or an error if the parameter is unavailable.
		Failure to find the given parameter should be indicated by returning an error.
	*/
	tGet(name string) (interface{}, error)
}

type tMapParameters map[string]interface{}

func (p tMapParameters) tGet(name string) (interface{}, error) {

	value, found := p[name]

	if !found {
		errorMessage := "No parameter '" + name + "' found."
		return nil, errors.New(errorMessage)
	}

	return value, nil
}
