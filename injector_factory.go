package introspector

import (
	"reflect"

	"github.com/infinytum/injector"
)

func InjectorFactoryFunc(forType reflect.Type) (*reflect.Value, error) {
	isPointer := false
	if _, resolveErr := injector.InjectT(forType); resolveErr != nil && forType.Kind() == reflect.Pointer {
		isPointer = true
		forType = forType.Elem()
	}

	ctx := reflect.New(forType).Interface()
	hasInjected := false

	if forType.Kind() == reflect.Struct || forType.Kind() == reflect.Pointer {
		// First test if its a struct dependency
		if err := injector.InjectInto(ctx); err != nil {
			// If not, structs are usually Context objects that hold one or multiple fields
			// that must be filled with dependencies
			if err2 := injector.Fill(ctx); err2 != nil {
				if err != injector.ErrorDepFactoryNotFound {
					return nil, err
				}
				return nil, err2
			}
		}
		hasInjected = true
	}

	// Interfaces / built-in types are always singular dependencies.
	if !hasInjected {
		if err := injector.InjectInto(ctx); err != nil {
			return nil, err
		}
	}

	value := reflect.ValueOf(ctx)
	if !isPointer {
		value = value.Elem()
	}
	return &value, nil
}
