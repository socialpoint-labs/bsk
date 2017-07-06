package logx_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/socialpoint-labs/bsk/logx"
	"github.com/stretchr/testify/assert"
)

func TestDefaultAndLogstashLogging(t *testing.T) {
	assert := assert.New(t)

	rec := make(recorder, 1)
	defaultLogger := logx.New(logx.WriterOpt(rec), logx.WithoutTimeOpt())
	logstashLogger := logx.NewLogstash("mychan", "myprod", "myapp", logx.WriterOpt(rec), logx.WithoutTimeOpt())

	hostname, _ := os.Hostname()

	for _, tc := range []struct {
		logger  logx.Logger
		message string
		fields  []logx.Field
		output  string
	}{
		{defaultLogger, "", nil, "DEBU \n"},
		{defaultLogger, "Test", nil, "DEBU Test\n"},
		{defaultLogger, "Test 2", []logx.Field{logx.F("foo", "some stuff")}, "DEBU Test 2 FIELDS foo=some stuff\n"},
		// "type" is a logstash reserved keyword but just changes in logstash log
		{defaultLogger, "Test 3", []logx.Field{logx.F("type", "val")}, "DEBU Test 3 FIELDS type=val\n"},
		{defaultLogger, "Test 4", []logx.Field{logx.F("number", 111)}, "DEBU Test 4 FIELDS number=111\n"},
		{defaultLogger, "Test 5", []logx.Field{logx.F("type", "val"), logx.F("myint", 111), logx.F("myfloat", 3.1416)}, "DEBU Test 5 FIELDS type=val myint=111 myfloat=3.1416\n"},

		{logstashLogger, "", nil, fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"message\":\"\",\"product\":\"myprod\",\"severity\":\"DEBU\"}\n", hostname)},
		{logstashLogger, "Test", nil, fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"message\":\"Test\",\"product\":\"myprod\",\"severity\":\"DEBU\"}\n", hostname)},
		{logstashLogger, "Test 2", []logx.Field{logx.F("foo", "some stuff")}, fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"foo\":\"some stuff\",\"message\":\"Test 2\",\"product\":\"myprod\",\"severity\":\"DEBU\"}\n", hostname)},
		// "type" is a logstash reserved keyword but just changes in logstash log
		{logstashLogger, "Test 3", []logx.Field{logx.F("type", "val")}, fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"message\":\"Test 3\",\"product\":\"myprod\",\"severity\":\"DEBU\",\"typex\":\"val\"}\n", hostname)},
		{logstashLogger, "Test 4", []logx.Field{logx.F("number", 111)}, fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"message\":\"Test 4\",\"number\":\"111\",\"product\":\"myprod\",\"severity\":\"DEBU\"}\n", hostname)},
		{logstashLogger, "Test 5", []logx.Field{logx.F("type", "val"), logx.F("number", 111)}, fmt.Sprintf("{\"@version\":1,\"app_server_name\":\"%s\",\"application\":\"myapp\",\"channel\":\"mychan\",\"message\":\"Test 5\",\"number\":\"111\",\"product\":\"myprod\",\"severity\":\"DEBU\",\"typex\":\"val\"}\n", hostname)},
	} {
		if tc.fields != nil {
			tc.logger.Debug(tc.message, tc.fields...)
		} else {
			tc.logger.Debug(tc.message)
		}
		assert.Equal(tc.output, <-rec)
	}
}

func TestLogLevel(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	logger := logx.New(logx.WriterOpt(&buf))

	logger.Debug("test")
	content, err := ioutil.ReadAll(&buf)
	assert.NoError(err)
	assert.True(len(content) > 0)

	logger.Info("test2")
	content, err = ioutil.ReadAll(&buf)
	assert.NoError(err)
	assert.True(len(content) > 0)

	logger = logx.New(logx.WriterOpt(&buf), logx.LevelOpt(logx.InfoLevel))

	// since now the min level is info then a debug message won't be logged
	logger.Debug("test")
	content, err = ioutil.ReadAll(&buf)
	assert.NoError(err)
	assert.Len(content, 0)

	logger.Info("test2")
	content, err = ioutil.ReadAll(&buf)
	assert.NoError(err)
	assert.True(len(content) > 0)
}

func TestDummy(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	logger := logx.NewDummy(logx.WriterOpt(&buf))

	logger.Debug("test")
	content, err := ioutil.ReadAll(&buf)
	assert.NoError(err)
	assert.Len(content, 0)

	logger.Info("test2")
	content, err = ioutil.ReadAll(&buf)
	assert.NoError(err)
	assert.Len(content, 0)
}

type recorder chan string

func (r recorder) Write(b []byte) (n int, err error) {
	r <- string(b)
	return len(b), nil
}
