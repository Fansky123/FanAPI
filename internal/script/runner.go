package script

import (
	"fmt"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// RunMapRequest executes the MapRequest function in a yaegi script.
func RunMapRequest(scriptSrc string, input map[string]interface{}) (map[string]interface{}, error) {
	return runMapFn(scriptSrc, "main.MapRequest", input)
}

// RunMapResponse executes the MapResponse function in a yaegi script.
func RunMapResponse(scriptSrc string, input map[string]interface{}) (map[string]interface{}, error) {
	return runMapFn(scriptSrc, "main.MapResponse", input)
}

func runMapFn(scriptSrc, fnPath string, input map[string]interface{}) (map[string]interface{}, error) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	if _, err := i.Eval(scriptSrc); err != nil {
		return nil, fmt.Errorf("script eval error: %w", err)
	}

	v, err := i.Eval(fnPath)
	if err != nil {
		return nil, fmt.Errorf("function %q not found: %w", fnPath, err)
	}

	fn, ok := v.Interface().(func(map[string]interface{}) map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("function %q must have signature func(map[string]interface{}) map[string]interface{}", fnPath)
	}

	_ = reflect.TypeOf(fn) // satisfy import
	result := fn(input)
	if result == nil {
		return input, nil
	}
	return result, nil
}
