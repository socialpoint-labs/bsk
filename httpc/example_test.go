package httpc_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/socialpoint-labs/bsk/httpc"
)

func ExampleHeader() {
	// Decorate the default client with a decorator that adds a header to all
	// request issued
	client := httpc.Decorate(http.DefaultClient, httpc.Header("test", "123"))

	req, err := http.NewRequest("GET", "http://www.google.com/robots.txt", nil)

	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StatusCode)

	// Output: 200
}

func ExampleFaultTolerance() {
	attempts := 5
	backoff := time.Millisecond * 500

	client := httpc.Decorate(http.DefaultClient, httpc.FaultTolerance(attempts, backoff))

	req, err := http.NewRequest("GET", "http://www.google.com/robots.txt", nil)

	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StatusCode)

	// Output: 200
}
