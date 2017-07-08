package httpx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/socialpoint-labs/bsk/httpx"
)

func Example_Router_Multi_Route_Endpoint_With_Prefix() {
	writer := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.URL.String())
		}
	}

	nestedRouter := httpx.NewRouter()
	nestedRouter.Route("/p1", writer())
	nestedRouter.Route("/p2", writer())

	router := httpx.NewRouter()
	router.Route("/prefix", nestedRouter)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/prefix/p1", nil)
	if err != nil {
		panic(err)
	}

	router.ServeHTTP(w, r)

	w = httptest.NewRecorder()
	r, err = http.NewRequest("GET", "/prefix/p2", nil)
	if err != nil {
		panic(err)
	}

	router.ServeHTTP(w, r)

	// Output:
	// /p1
	// /p2
}
