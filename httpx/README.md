# package httpx

`package httpx` provides a set of extensions to the `net/http` package.

## Design Principles

- Embrace the standard library
    - use the standard library for as long as possible
    - you may not need an external package at all
    - use http.Handler and http.HandlerFunc
- Use HTTP query strings instead of adding the complexity of a complex and slow multiplexer
- Complex path structures ties to a fixed structure are difficult to remember and document
- This is intended to be consumed by machines in most cases, not for final users
- Method agnostic

## Why not using an HTTP micro-framework?

Because it's not needed at all:
- We can use the standard library for as long as possible
- We can use http.Handler and http.HandlerFunc as the core foundation of our HTTP services
- Micro-frameworks often adds features we don't need:
   - pretty (user friendly) URLs
   - defaults intended for web sites and web development

However, most of our services will be consumed by machines, where we can follow this principles:
- HTTP query strings usage, instead of adding a complex and slow multiplexer to decode pretty paths
- Method agnostic, this is not REST, we can express intentions in other ways (for example, RPC)


## Decorators

- They are shared functionality that you want to run for many (or even all) HTTP requests.
- They wrap/decorate `http.Handler` with additional functionality.
- They must be composable.

For example, you might want to log every request, gzip every response, instrument or check security. These shared
responsibilities are better implemented by decorating existing `http.Handler` implementations.

To learn more see code and usage examples.
