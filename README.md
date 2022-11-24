# Introspector

The introspector package allows to create different instrospectors that hold known factory functions.
These introspectors can then introspect any given function's parameters and determine a matching factory
to produce the argument for one or many of the parameters.

## Purpose

The introspector works best as an extension to the injector. The injector is a dependency registry / injector
which allows you to get dependencies that are constructed the same for all callers. The introspector
allowy you to have function parameters that are generated on the fly based on the original input.

This sounds more complicated than it really is. For example, mojito uses the introspector to know what
kind of Context a handler is expecting. The original input is always a **mojito.Context** which can for
example be converted to a **mojito.RenderContext**. The function should not have to deal with this and
the original caller cannot know what a handler function looks like.

## Explanation

We will work with the mojito example in this explanation as it shows a great use-case for the introspector.
The introspector will be responsible for generating handler arguments.

### Factory Function

Every request that goes into mojito is converted into a **mojito.Context**. For mojito, this is the _original input_.
With this information we can define the first type that is necessary for an introspector: The factory function

```go
type HandlerArgFactory func(ctx mojito.Context) reflect.Value
```

A common naming scheme for these is <Product>Factory. Mojito only produces the context, so its the sole parameter
to this type / func. It's common to return a reflect.Value from a factory, since it's a common return type every
function can fulfil. Also later when calling the actual handler, you will need an array of reflect.Value anyways.

### Result Type

Every introspector generates a result after introspecting a function. This result contains at least the type
of the introspected function and a map of factories that were found for the parameters. The introspector requires
this type to implement and embed the **introspector.IntrospectorResult** type.

```go
type HandlerResult struct {
    introspector.IntrospectorResult[HandlerArgFactory]
}
```

This is the minimum type you have to declare. You can add more fields and functions if need be.

### Creating the introspector

You have defined both required types and can now create an introspector. Creating an introspector is
a simple function call.

```go
i, err := introspector.NewIntrospector[HandlerArgFactory, HandlerResult]()
```

Your introspector is now ready, but it won't be able to map any factories yet, as there are none registered.

### Registering a factory function

We are now ready to create factories. Let's create a hypothetical factory that will create **mojito.RenderContext**.
The type definition and proper initialization will not be part of the example.

```go
introspector.RegisterFactory[mojito.RenderContext](func (ctx mojito.Context) reflect.Value {
    newCtx := myRenderContext{
        Context: ctx,
        someField: defaultValue,
        ...
    }
    return reflect.ValueOf(newCtx)
}, i)
```

While you can directly register a type using i.RegisterFactory, the helper function allows you to use generics
to specify the type you are registering a factory for. This is much easier for some types instead of trying to
get the reflect.TypeOf(...) yourself.

### Enable introspector to resolve injector dependencies

Chances are that your handler also wants to make use of regular injector dependencies to access a database
or similar. This can be enabled by defining a default factory on the introspector.

```
i.SetDefaultFactory(func(r reflect.Type) (HandlerArgFactory, error) {
	val, err := introspector.InjectorFactoryFunc(r)
	return func(ctx mojito.Context) reflect.Value {
		return *val
	}, err
})
```

As you can see, the introspector package provides an adapter that you just have to convert to your factory
func signature. As you don't require any of the inputs, it is quite easy.

**Important to note**: The position of the introspector.InjectorFactoryFunc call determines if the injector
is called once during introspection or every time the factory is called! If you want the dependency to only
get resolved once, put it as shown. If you want the dependency to be resolved every time the factory is called
put it inside the function that is being returned!

### Introspect a function

The introspector is now ready to introspect functions and map factories to its parameters.

```go
result, err := i.Introspect(func(ctx mojito.RenderContext) {})
```

If there is any errors, it usually means one or more arguments couldn't be resolved to a factory.
Assuming that is not the case, the factory map should contain a factory for each parameter of the function,
indexed by the position of the parameter in the function signature.

```go
factory := result.FactoryMap()[0] // Should contain the factory for a mojito.RenderContext
```

These factories are ready to generate parameters. That means once you introspected a function you can just
cache the factory map and you never have to do it again. Now you are ready to call your handlers.

### Using the factory map

To use the factory map to create arguments, you best iterate over the map to create a new array of reflect.Value.
That array will be used to make the call to your handler using reflection.

```go
var ctx mojito.Context // We get this from the incoming request

args := make([]reflect.Value, result.Type().NumIn())
for paramIndex, factory := range result.FactoryMap() {
    args[paramIndex] = factory(ctx)
}
reflect.ValueOf(func(ctx mojito.RenderContext) {}).Call(args)
```

The function will now be called with a mojito.RenderContext. Neither the calling code nor the function
had to worry about any conversion and the factory func ensured type safety even though we are using reflect.
