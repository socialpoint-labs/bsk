package httpx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/httpx"
)

func TestIsClientError(t *testing.T) {
	assert.True(t, httpx.IsClientError(403))
	assert.False(t, httpx.IsClientError(200))
}

func TestIsSuccessful(t *testing.T) {
	assert.True(t, httpx.IsSuccessful(204))
	assert.False(t, httpx.IsSuccessful(500))
}

func TestIsRedirection(t *testing.T) {
	assert.True(t, httpx.IsRedirection(301))
	assert.False(t, httpx.IsRedirection(200))
}

func TestIsServerError(t *testing.T) {
	assert.True(t, httpx.IsServerError(502))
	assert.False(t, httpx.IsServerError(200))
}
