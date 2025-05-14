package main

import (
	"context"
	"net/http"
)

func handleMiddleware(c context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  // TODO Do something with our context
	  next.ServeHTTP(w, r)
	})
  }