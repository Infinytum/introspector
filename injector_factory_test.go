package introspector_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/infinytum/injector"
	"github.com/infinytum/introspector"
)

type TestInterface interface {
	GetValue() int
}

type TestDependency struct {
	Value int
}

func (d TestDependency) GetValue() int {
	return d.Value
}

type TestContext struct {
	Dep TestDependency `injector:"type"`
}

type TestFaultyContext struct {
	Dep http.ResponseWriter `injector:"type"`
}

type InjectorFactory func() reflect.Value

type InjectorResult struct {
	introspector.IntrospectorResult[InjectorFactory]
}

// Write a test that verifies non-pointer structs are injected correctly.
//
// The test should:
// - register a non-pointer struct dependency
// - call the NewIntrospector function
// - introspect a function that depends on the registered dependency
// - verify that the dependency is injected correctly
// - introspect a function that depends on a pointer to the registered dependency
// - verify that the dependency is injected correctly
func TestNonPointerStructDep(t *testing.T) {
	injector.Singleton(func() TestDependency {
		return TestDependency{
			Value: 69,
		}
	})
	i, err := introspector.NewIntrospector[InjectorFactory, InjectorResult]()
	if err != nil {
		t.Fatal(err)
	}

	i.SetDefaultFactory(func(t reflect.Type) (InjectorFactory, error) {
		val, err := introspector.InjectorFactoryFunc(t)
		return func() reflect.Value {
			return *val
		}, err
	})

	res, errs := i.Introspect(func(d TestDependency) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	factory := res.FactoryMap()[0]

	if factory().Interface().(TestDependency).Value != 69 {
		t.Fatal("expected dependency to be 69")
	}

	res, errs = i.Introspect(func(d *TestDependency) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	factory = res.FactoryMap()[0]

	if factory().Interface().(*TestDependency).Value != 69 {
		t.Fatal("expected dependency to be 69")
	}
}

// Write a test that verifies pointer structs are injected correctly.
//
// The test should:
// - register a pointer struct dependency
// - call the NewIntrospector function
// - introspect a function that depends on the registered dependency
// - verify that the dependency is injected correctly
func TestPointerStructDep(t *testing.T) {
	injector.Singleton(func() *TestDependency {
		return &TestDependency{
			Value: 69,
		}
	})
	i, err := introspector.NewIntrospector[InjectorFactory, InjectorResult]()
	if err != nil {
		t.Fatal(err)
	}

	i.SetDefaultFactory(func(t reflect.Type) (InjectorFactory, error) {
		val, err := introspector.InjectorFactoryFunc(t)
		return func() reflect.Value {
			return *val
		}, err
	})

	res, errs := i.Introspect(func(d *TestDependency) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	factory := res.FactoryMap()[0]

	if factory().Interface().(*TestDependency).Value != 69 {
		t.Fatal("expected dependency to be 69")
	}
}

// Write a test that verifies interfaces are injected correctly.
//
// The test should:
// - register an interface dependency
// - call the NewIntrospector function
// - introspect a function that depends on the registered dependency
// - verify that the dependency is injected correctly
func TestInterfaceDep(t *testing.T) {
	injector.Singleton(func() TestInterface {
		return TestDependency{
			Value: 69,
		}
	})
	i, err := introspector.NewIntrospector[InjectorFactory, InjectorResult]()
	if err != nil {
		t.Fatal(err)
	}

	i.SetDefaultFactory(func(t reflect.Type) (InjectorFactory, error) {
		val, err := introspector.InjectorFactoryFunc(t)
		return func() reflect.Value {
			return *val
		}, err
	})

	res, errs := i.Introspect(func(d TestInterface) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	factory := res.FactoryMap()[0]

	if factory().Interface().(TestInterface).GetValue() != 69 {
		t.Fatal("expected dependency to be 69")
	}
}

// Write a test that verifies ErrorDepFactoryNotFound is returned when a dependency is not registered.
//
// The test should:
// - call the NewIntrospector function
// - introspect a function that depends on a dependency that is not registered
// - verify that a ErrorDepFactoryNotFound is returned
func TestErrorDepFactoryNotFound(t *testing.T) {
	i, err := introspector.NewIntrospector[InjectorFactory, InjectorResult]()
	if err != nil {
		t.Fatal(err)
	}

	i.SetDefaultFactory(func(t reflect.Type) (InjectorFactory, error) {
		val, err := introspector.InjectorFactoryFunc(t)
		return func() reflect.Value {
			return *val
		}, err
	})

	_, errs := i.Introspect(func(d http.ResponseWriter) {})
	if len(errs) == 0 {
		t.Fatal("expected error")
	}

	if errs[0] != injector.ErrorDepFactoryNotFound {
		t.Fatalf("expected ErrorDepFactoryNotFound, got %v", err)
	}
}

// Write a test that verifies struct fill dependencies are injected correctly.
//
// The test should:
// - register a struct fill dependency
// - call the NewIntrospector function
// - introspect a function that depends on the registered dependency
// - verify that the dependency is injected correctly
func TestStructFillDep(t *testing.T) {
	i, err := introspector.NewIntrospector[InjectorFactory, InjectorResult]()
	if err != nil {
		t.Fatal(err)
	}

	i.SetDefaultFactory(func(t reflect.Type) (InjectorFactory, error) {
		val, err := introspector.InjectorFactoryFunc(t)
		return func() reflect.Value {
			return *val
		}, err
	})

	res, errs := i.Introspect(func(d TestContext) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	factory := res.FactoryMap()[0]

	if factory().Interface().(TestContext).Dep.Value != 69 {
		fmt.Println(factory().Interface().(TestContext))
		t.Fatal("expected dependency to be 69")
	}

	_, errs = i.Introspect(func(d TestFaultyContext) {})
	if len(errs) == 0 {
		t.Fatal("expected error")
	}

	if errs[0] != injector.ErrorDepFactoryNotFound {
		t.Fatalf("expected ErrorDepFactoryNotFound, got %v", errs[0])
	}
}

// Write a test that verifies built-in types are injected correctly.
//
// The test should:
// - register a built-in type dependency
// - call the NewIntrospector function
// - introspect a function that depends on the registered dependency
// - verify that the dependency is injected correctly
func TestBuiltinDep(t *testing.T) {
	injector.Singleton(func() int {
		return 69
	})
	i, err := introspector.NewIntrospector[InjectorFactory, InjectorResult]()
	if err != nil {
		t.Fatal(err)
	}

	i.SetDefaultFactory(func(t reflect.Type) (InjectorFactory, error) {
		val, err := introspector.InjectorFactoryFunc(t)
		return func() reflect.Value {
			return *val
		}, err
	})

	res, errs := i.Introspect(func(d int) {})
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	factory := res.FactoryMap()[0]

	if factory().Interface().(int) != 69 {
		t.Fatal("expected dependency to be 69")
	}
}
