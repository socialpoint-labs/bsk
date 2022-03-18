# package logx

`package logx` is a minimal log package inspired by [logrus](https://github.com/Sirupsen/logrus) and [zap](https://github.com/uber-common/zap) that follows [these Dave Cheney guidelines](https://dave.cheney.net/2015/11/05/lets-talk-about-logging).

Features:

-  pluggable io.Writer with the `WithWriter` decorator, default is os.Stdout
-  2 marshallers available, a logstash one and a human-readable one
-  support ~~debug~~ info and error levels, and a `WithLevel` decorator to change the level
-  support structured fields
-  provides a Dummy logger for testing purposes
