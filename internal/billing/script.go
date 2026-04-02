package billing

import (
	"fmt"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// RunBillingScript executes a custom billing script.
// If resp is nil, calls CalcCost(req); otherwise calls CalcActualCost(req, resp).
func RunBillingScript(script string, req, resp map[string]interface{}) (int64, error) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	if _, err := i.Eval(script); err != nil {
		return 0, fmt.Errorf("billing script eval error: %w", err)
	}

	var v reflect.Value
	var err error
	if resp == nil {
		v, err = i.Eval(`main.CalcCost`)
	} else {
		v, err = i.Eval(`main.CalcActualCost`)
	}
	if err != nil {
		return 0, fmt.Errorf("billing script function not found: %w", err)
	}

	fn, ok := v.Interface().(func(map[string]interface{}) int)
	if !ok {
		fn2, ok2 := v.Interface().(func(map[string]interface{}, map[string]interface{}) int)
		if !ok2 {
			return 0, fmt.Errorf("billing script function has wrong signature")
		}
		return int64(fn2(req, resp)), nil
	}
	return int64(fn(req)), nil
}
