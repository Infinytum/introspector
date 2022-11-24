package introspector_test

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/infinytum/injector"
	"github.com/infinytum/introspector"
)

type TestFactoryFunc func(int) reflect.Value

type TestResult struct {
	introspector.IntrospectorResult[TestFactoryFunc]
}

type InvalidTestResult struct {
}

func (i InvalidTestResult) FactoryMap() map[int]TestFactoryFunc {
	return nil
}

func (i InvalidTestResult) Type() reflect.Type {
	return nil
}

// Write a test that verifies that NewIntrospector returns an instance of the Introspector interface.
//
// The test should:
// - call the NewIntrospector function
// - verify that the returned value is an instance of the Introspector interface
func TestNewIntrospector(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	if i == nil {
		t.Fatal("expected introspector to be not nil")
	}
}

// Write a test that verifies that NewIntrospector returns an error if the result type does not implement the IntrospectorResult interface correctly.
//
// The test should:
// - call the NewIntrospector function with a result type that does not implement the IntrospectorResult interface
// - verify that the returned error is not nil
func TestNewIntrospectorInvalidResultType(t *testing.T) {
	_, err := introspector.NewIntrospector[TestFactoryFunc, InvalidTestResult]()
	if err == nil {
		t.Fatal("expected error to be not nil")
	}
}

// Write a test that verifies that FactoryMap is not nil after calling NewIntrospector.
//
// The test should:
// - call the NewIntrospector function
// - verify that the FactoryMap method of the returned value is not nil
func TestFactoryMapNotNil(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	if i.FactoryMap() == nil {
		t.Fatal("expected factory map to be not nil")
	}
}

// Write a test that verifies that the RegisterFactory function works as expected.
//
// The test should:
// - create an instance of the Introspector interface
// - call the RegisterFactory function with a factory function and the Introspector instance
// - verify that the factory function is registered on the Introspector instance
func TestRegisterFactory(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	introspector.RegisterFactory[int](func(i int) reflect.Value {
		return reflect.ValueOf(i * 2)
	}, i)
	if len(i.FactoryMap()) != 1 {
		t.Fatal("expected factory map to have one entry")
	}
	factory, ok := i.FactoryMap()[reflect.TypeOf(int(1))]
	if !ok {
		t.Fatal("expected factory map to have entry for type int")
	}
	if factory(1).Interface().(int) != 2 {
		t.Fatal("expected factory to return 2")
	}
}

// Write a test that verifies that the SetDefaultFactory function works as expected.
//
// The test should:
// - create an instance of the Introspector interface
// - call the SetDefaultFactory function with a default factory function and the Introspector instance
// - verify that the default factory function is registered on the Introspector instance
func TestSetDefaultFactory(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	u := url.URL{}
	injector.Singleton(func() url.URL {
		return u
	})
	i.SetDefaultFactory(func(r reflect.Type) (TestFactoryFunc, error) {
		val, err := introspector.InjectorFactoryFunc(r)
		return func(i int) reflect.Value {
			return *val
		}, err
	})
	if len(i.FactoryMap()) != 0 {
		t.Fatal("expected factory map to have zero entry")
	}

	res, errs := i.Introspect(func(i url.URL) {})
	if len(errs) != 0 {
		t.Fatal(err)
	}

	if len(res.FactoryMap()) != 1 {
		t.Fatal("expected result factory map to have one entry")
	}

	factory, ok := res.FactoryMap()[0]
	if !ok {
		t.Fatal("expected factory map to have entry for arg 0")
	}

	if factory(1).Interface().(url.URL) != u {
		t.Fatal("expected factory to return same url")
	}
}

// Write a test that verifies that the Introspect function works as expected.
//
// The test should:
// - create an instance of the Introspector interface
// - register a factory function on the Introspector instance
// - call the Introspect function with a function and the Introspector instance
// - verify that the Introspect function returns an instance of the IntrospectorResult interface
func TestIntrospect(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	introspector.RegisterFactory[int](func(i int) reflect.Value {
		return reflect.ValueOf(i * 2)
	}, i)
	res, errs := i.Introspect(func(i int) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	if res == nil {
		t.Fatal("expected result to be not nil")
	}

	if len(res.FactoryMap()) != 1 {
		t.Fatal("expected result factory map to have one entry")
	}

	if res.Type() != reflect.TypeOf(func(i int) {}) {
		t.Fatal("expected result type to be func(i int)")
	}

	factory, ok := res.FactoryMap()[0]
	if !ok {
		t.Fatal("expected factory map to have entry for arg 0")
	}

	if factory(1).Interface().(int) != 2 {
		t.Fatal("expected factory to return 2")
	}
}

// Write a test that verifies that the Introspect function returns an error if no factory function is registered for a parameter type.
//
// The test should:
// - create an instance of the Introspector interface
// - call the Introspect function with a function and the Introspector instance
// - verify that the Introspect function returns an error
func TestIntrospectNoFactory(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	_, errs := i.Introspect(func(i int) {})
	if len(errs) == 0 {
		t.Fatal("expected error to be not nil")
	}
}

// Write a test that verifies that the Introspect function returns an error if the function has a variadic parameter.
//
// The test should:
// - create an instance of the Introspector interface
// - call the Introspect function with a function that has a variadic parameter
// - verify that the Introspect function returns an error
func TestIntrospectVariadic(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	_, errs := i.Introspect(func(i ...int) {})
	if len(errs) == 0 {
		t.Fatal("expected error to be not nil")
	}
}

// Write a test that verifies that the Introspect function returns an error if the default factory function returns an error.
//
// The test should:
// - create an instance of the Introspector interface
// - call the SetDefaultFactory function with a default factory function that returns an error
// - call the Introspect function with a function
// - verify that the Introspect function returns an error
func TestIntrospectDefaultFactoryError(t *testing.T) {
	i, err := introspector.NewIntrospector[TestFactoryFunc, TestResult]()
	if err != nil {
		t.Fatal(err)
	}
	i.SetDefaultFactory(func(r reflect.Type) (TestFactoryFunc, error) {
		return nil, errors.New("test error")
	})
	_, errs := i.Introspect(func(i int) {})
	if len(errs) == 0 {
		t.Fatal("expected error to be not nil")
	}
}
