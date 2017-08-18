# BSK [![Build Status](https://travis-ci.org/socialpoint-labs/bsk.svg?branch=master)](https://travis-ci.org/socialpoint-labs/bsk) [![Coverage Status](https://coveralls.io/repos/github/socialpoint-labs/bsk/badge.svg?branch=master)](https://coveralls.io/github/socialpoint-labs/bsk?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/socialpoint-labs/bsk)](https://goreportcard.com/report/github.com/socialpoint-labs/bsk)

[![Join the chat at https://gitter.im/socialpoint-labs/bsk](https://badges.gitter.im/socialpoint-labs/bsk.svg)](https://gitter.im/socialpoint-labs/bsk?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Packages

There are two types of packages: 
- `ext`: extensions to the stdlib, they augment stdlib package functionalities.
  Usually named after a stdlib package name plus `x`.
- `new`: they provide logic not present in the stdlib.

| Package                | Type  | Description                                                                       |
| ---                    | ----  | -----------                                                                       |
| [`httpx`](httpx)       | `ext` | Extends `net/http` with routing utilities, decorators, etc                        |
| [`logx`](logx)         | `ext` | Extends `log` with structured logging and logstash support                        |
| [`contextx`](contextx) | `ext` | Extends `context` with types and decorators to run tasks with a `context.Context` |
| [`timex`](timex)       | `ext` | Extends `time` with other time functions and utilities                            |

## Contributors

This code was initially part of a private repository at Social Point.

It's a collaborative effort of all the backend team!

In order to simplify the process of moving the code to a public repository, original commits authors' could be lost.

Special thanks to all the people who has been collaborating (in alphabetical order):

- [Adán Lobato](https://github.com/adanlobato)
- [Alex Carol](https://github.com/alexcarol)
- [Emili Calonge](https://github.com/1000i1)
- [Gonzalo Serrano](https://github.com/gonzaloserrano)
- [Guillem Nieto](https://github.com/gnieto)
- [Hernán Kleiman](https://github.com/mrjusti)
- [Javier Expósito](https://github.com/javierExposito)
- [Jordi Forns](https://github.com/jforns)
- [Julio de la Calle](https://github.com/dixso)
- [Ludovic Livain](https://github.com/sp-ludovic-ivain)
- [Manuel Jurado](https://github.com/manuelljb)
- [Manuel Peralta](https://github.com/---)
- [Roger Clotet](https://github.com/rogerclotet)
- [Ronny López](https://github.com/ronnylt)
- [Rubén Simón](https://github.com/neomede)
- [Sergio Toledo](https://github.com/toledoom)
