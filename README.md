# BSK [![Build Status](https://travis-ci.org/socialpoint-labs/bsk.svg?branch=master)](https://travis-ci.org/socialpoint-labs/bsk) [![codecov](https://codecov.io/gh/socialpoint-labs/bsk/branch/master/graph/badge.svg)](https://codecov.io/gh/socialpoint-labs/bsk) [![Go Report Card](https://goreportcard.com/badge/github.com/socialpoint-labs/bsk)](https://goreportcard.com/report/github.com/socialpoint-labs/bsk)

[![Join the chat at https://gitter.im/socialpoint-labs/bsk](https://badges.gitter.im/socialpoint-labs/bsk.svg)](https://gitter.im/socialpoint-labs/bsk?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Packages

### Extensions

Extensions are packages tha extends other packages from the standard library or from an official package.

| Package | Description|
| --- | --- |
| [`awsx`](awsx)         | Extends package `github.com/aws/aws-sdk-go` with testing utilities and easier to use helpers. |
| [`contextx`](contextx) | Extends `context` with types and decorators to run tasks with a `context.Context` |
| [`grpcx`](grpcx)       | Extends package `google.golang.org/grpc` with interceptors |
| [`httpx`](httpx)       | Extends package `net/http` with routing utilities, decorators, etc... |
| [`logx`](logx)         | Extends `log` with structured logging and logstash support |
| [`timex`](timex)       | Extends `time` with other time functions and utilities |

### Utilities

Utilities are general purpose packages that provides specific functionalities.

| Package | Description |
| --- | --- |
| [`dispatcher`](dispatcher) | Adds a reflect-based framework for dispatching events |
| [`metrics`](metrics)       | Adds metrics to instrument and publish to e.g DataDog |
| [`multierror`](multierror) | Allows piling multiple errors in a single error object |
| [`netutil`](netutil)       | Net utilities to get free network ports |
| [`recovery`](recovery)     | Offers panic recovery utils |
| [`run`](run)               | Functions to manage runtime execution flow |
| [`uuid`](uuid)             | An utility package to generate time-ordered UUIDs |

## Contributors

This code was initially part of a private repository at Social Point.

It's a collaborative effort of all the backend team!

In order to simplify the process of moving the code to a public repository, original commits authors' could be lost.

Special thanks to all the people who have collaborated (in alphabetical order):

- [Adán Lobato](https://github.com/adanlobato)
- [Alex Carol](https://github.com/alexcarol)
- [Emili Calonge](https://github.com/1000i1)
- [Gonzalo Serrano](https://github.com/gonzaloserrano)
- [Guillem Nieto](https://github.com/gnieto)
- [Hernán Kleiman](https://github.com/mrjusti)
- [Javier Expósito](https://github.com/javierExposito)
- [Jordi Forns](https://github.com/jforns)
- [Julio de la Calle](https://github.com/dixso)
- [Ludovic Ivain](https://github.com/sp-ludovic-ivain)
- [Manuel Jurado](https://github.com/manuelljb)
- [Manuel Peralta](https://github.com/---)
- [Roger Clotet](https://github.com/rogerclotet)
- [Ronny López](https://github.com/ronnylt)
- [Rubén Simón](https://github.com/neomede)
- [Sergio Toledo](https://github.com/toledoom)
