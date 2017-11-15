# Changelog

## v2 - 2017-11-15

The experiment engine has been completely rewritten. It's a lot simpler in usage
and architecture.

## v1.1.0 - 2016-08-22

- Removed testify dependency
- Removed internal interface dependencies, rely on structs from now on
- Removed `x/net/context` dependency

### Context

The context interface has been copied into the package for backward
compatibility reasons. In Go 1.7 this has been moved to the standard library,
but this would mean this package isn't available for other Go versions. In the
tests we also use the `x/net/context` package to test some context behaviour.

Previously, one could inject a `nil` context and it would be converted to a
proper `context.Background()`. This is not the case anymore. The user should
always inject a context. If the purpose of the context is still unknown, use
`context.TODO()`.

## v1.0.0 - 2016-08-08
First official release.
