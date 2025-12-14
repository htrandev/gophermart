package middleware

import (
	"errors"
	"net/http"
)

var errTest = errors.New("test error")

func dummyHandler() http.HandlerFunc {
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(dummyHandler)
}

func dummyHandlerWithResponse() http.HandlerFunc {
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("200 OK"))
		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(dummyHandler)
}
