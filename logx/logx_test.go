package logx_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/logx"
)

func TestDefaultAndLogstashLogging(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	rec := make(recorder, 1)
	defaultLogger := logx.New(logx.WriterOpt(rec), logx.WithoutTimeOpt())
	defaultLoggerWithoutFileInfo := logx.New(logx.WriterOpt(rec), logx.WithoutTimeOpt(), logx.WithoutFileInfo())
	logstashLogger := logx.NewLogstash("mychan", "myprod", "myapp", logx.WriterOpt(rec), logx.WithoutTimeOpt())
	logstashLoggerWithOriginalValues := logx.New(logx.MarshalerOpt(logx.NewLogstashMarshaler("mychan", "myprod", "myapp", logx.WithOriginalValueTypes())), logx.WriterOpt(rec), logx.WithoutTimeOpt())
	logstashLoggerWithoutFileInfo := logx.NewLogstash("mychan", "myprod", "myapp", logx.WriterOpt(rec), logx.WithoutTimeOpt(), logx.WithoutFileInfo())
	logstashLoggerWithEnvironment := logx.New(logx.MarshalerOpt(logx.NewLogstashMarshaler("mychan", "myprod", "myapp", logx.WithEnvironment("prod"))), logx.WriterOpt(rec), logx.WithoutTimeOpt())

	hostname, _ := os.Hostname()

	for _, tc := range []struct {
		logger  logx.Logger
		message string
		fields  []logx.Field
		output  string
	}{
		{logger: defaultLogger, output: "INFO  File: logx_test.go:57\n"},
		{logger: defaultLogger, message: "Test", output: "INFO Test File: logx_test.go:57\n"},
		{logger: defaultLoggerWithoutFileInfo, message: "Test 2", fields: []logx.Field{logx.F("foo", "some stuff")}, output: "INFO Test 2 FIELDS foo=some stuff\n"},
		// "type" is a logstash reserved keyword but just changes in logstash log
		{logger: defaultLogger, message: "Test 3", fields: []logx.Field{logx.F("type", "val")}, output: "INFO Test 3 FIELDS type=val File: logx_test.go:55\n"},
		{logger: defaultLogger, message: "Test 4", fields: []logx.Field{logx.F("number", 111)}, output: "INFO Test 4 FIELDS number=111 File: logx_test.go:55\n"},
		{logger: defaultLoggerWithoutFileInfo, message: "Test 5", fields: []logx.Field{logx.F("type", "val"), logx.F("myint", 111), logx.F("myfloat", 3.1416)}, output: "INFO Test 5 FIELDS type=val myint=111 myfloat=3.1416\n"},

		{logger: logstashLogger, output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"file\":\"logx_test.go:57\",\"message\":\"\",\"product\":\"myprod\",\"severity\":\"INFO\"}\n", hostname)},
		{logger: logstashLogger, message: "Test", output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"file\":\"logx_test.go:57\",\"message\":\"Test\",\"product\":\"myprod\",\"severity\":\"INFO\"}\n", hostname)},
		{logger: logstashLoggerWithoutFileInfo, message: "Test 2", fields: []logx.Field{logx.F("foo", "some stuff")}, output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"foo\":\"some stuff\",\"message\":\"Test 2\",\"product\":\"myprod\",\"severity\":\"INFO\"}\n", hostname)},
		// "type" is a logstash reserved keyword but just changes in logstash log
		{logger: logstashLogger, message: "Test 3", fields: []logx.Field{logx.F("type", "val")}, output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"file\":\"logx_test.go:55\",\"message\":\"Test 3\",\"product\":\"myprod\",\"severity\":\"INFO\",\"typex\":\"val\"}\n", hostname)},
		{logger: logstashLogger, message: "Test 4", fields: []logx.Field{logx.F("number", 111)}, output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"file\":\"logx_test.go:55\",\"message\":\"Test 4\",\"number\":\"111\",\"product\":\"myprod\",\"severity\":\"INFO\"}\n", hostname)},
		{logger: logstashLoggerWithoutFileInfo, message: "Test 5", fields: []logx.Field{logx.F("type", "val"), logx.F("number", 111)}, output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"message\":\"Test 5\",\"number\":\"111\",\"product\":\"myprod\",\"severity\":\"INFO\",\"typex\":\"val\"}\n", hostname)},

		{logger: logstashLoggerWithOriginalValues, message: "Test With Original Values But No Fields", output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"file\":\"logx_test.go:57\",\"message\":\"Test With Original Values But No Fields\",\"product\":\"myprod\",\"severity\":\"INFO\"}\n", hostname)},
		{logger: logstashLoggerWithOriginalValues, message: "Test With Original Values", fields: []logx.Field{logx.F("string", "hi there"), logx.F("number", 123), logx.F("array", []int{1, 2, 3}), logx.F("map", map[string]int{"foo": 123, "bar": 456})}, output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"array\":[1,2,3],\"channel\":\"mychan\",\"file\":\"logx_test.go:55\",\"map\":{\"bar\":456,\"foo\":123},\"message\":\"Test With Original Values\",\"number\":123,\"product\":\"myprod\",\"severity\":\"INFO\",\"string\":\"hi there\"}\n", hostname)},
		{logger: logstashLoggerWithEnvironment, message: "Test", output: fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"environment\":\"prod\",\"file\":\"logx_test.go:57\",\"message\":\"Test\",\"product\":\"myprod\",\"severity\":\"INFO\"}\n", hostname)},
	} {
		if tc.fields != nil {
			tc.logger.Info(tc.message, tc.fields...)
		} else {
			tc.logger.Info(tc.message)
		}
		a.Equal(tc.output, <-rec)
	}
}

func TestLoggingWithCustomSkipLevel(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	rec := make(recorder, 1)
	defaultLogger := logx.New(logx.WriterOpt(rec), logx.WithoutTimeOpt(), logx.AdditionalFileSkipLevel(1))

	log(defaultLogger, "Test")
	a.Equal("INFO Test File: logx_test.go:69\n", <-rec)
}

func log(logger logx.Logger, message string) {
	logger.Info(message)
}

func TestLogLevel(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	var buf bytes.Buffer
	logger := logx.New(logx.WriterOpt(&buf))

	logger.Info("test")
	content, err := ioutil.ReadAll(&buf)
	a.NoError(err)
	a.True(len(content) > 0)

	logger.Error("test2")
	content, err = ioutil.ReadAll(&buf)
	a.NoError(err)
	a.True(len(content) > 0)

	logger = logx.New(logx.WriterOpt(&buf), logx.LevelOpt(logx.ErrorLevel))

	// since now the min level is error then a debug message won't be logged
	logger.Info("test")
	content, err = ioutil.ReadAll(&buf)
	a.NoError(err)
	a.Len(content, 0)

	logger.Error("test2")
	content, err = ioutil.ReadAll(&buf)
	a.NoError(err)
	a.True(len(content) > 0)
}

func TestDummy(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	var buf bytes.Buffer
	logger := logx.NewDummy(logx.WriterOpt(&buf))

	logger.Info("test")
	content, err := ioutil.ReadAll(&buf)
	a.NoError(err)
	a.Len(content, 0)

	logger.Error("test2")
	content, err = ioutil.ReadAll(&buf)
	a.NoError(err)
	a.Len(content, 0)
}

type recorder chan string

func (r recorder) Write(b []byte) (n int, err error) {
	r <- string(b)
	return len(b), nil
}
