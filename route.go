package mux

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

// Server const and types
type Handler func(w http.ResponseWriter, r *http.Request) error
type ErrorHandler func(e *HttpError, w http.ResponseWriter, r *http.Request)
type Matcher func(r *http.Request) (bool, error)

var METHODS = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}

type Route struct {
	handler Handler
	match   Matcher
}

// Create routes
func CreateRoute(match Matcher, handler Handler) *Route {
	var route = &Route{}
	route.match = match
	route.handler = handler
	return route
}

func CreateFileRoute(exact string, filePath string, writeHeaders http.Header) *Route {
	return CreateRoute(
		MatchPathExact([]string{"GET"}, filePath),
		CreateFileHandler(filePath, writeHeaders),
	)
}

// Create matchers
func MatchPathExact(methods []string, exact string) Matcher {
	return func(r *http.Request) (bool, error) {
		path := r.RequestURI
		method := r.Method
		if exact == path {
			if isStringInArray(method, methods) == true {
				return true, nil
			} else {
				return false, NewHttpError(405, fmt.Sprintf("The method %q is not allowed for path %q", method, path))
			}
		}
		return false, nil
	}
}

func MatchPathStartWith(methods []string, startWith string) Matcher {
	return func(r *http.Request) (bool, error) {
		path := r.RequestURI
		method := r.Method
		if startWith == path[:len(startWith)] {
			if isStringInArray(method, methods) == true {
				return true, nil
			} else {
				return false, NewHttpError(405, fmt.Sprintf("The method %q is not allowed for path %q", method, path))
			}
		}
		return false, nil
	}
}

func MatchPathRegexp(methods []string, rx string) Matcher {
	return func(r *http.Request) (bool, error) {
		path := r.RequestURI
		match, err := regexp.Match(rx, []byte(path))
		if err != nil {
			return false, err
		}
		method := r.Method
		if match {
			if isStringInArray(method, methods) == true {
				return true, nil
			} else {
				return false, NewHttpError(405, fmt.Sprintf("The method %q is not allowed for path %q", method, path))
			}
		}
		return false, nil
	}
}

// Create handlers
func CreateFileHandler(filePath string, writeHeaders http.Header) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		fileHandler, err := os.Open(filePath)
		defer fileHandler.Close()

		if err != nil {
			return NewHttpError(404, fmt.Sprintf("The resource at path %q is not found", filePath))
		}

		if writeHeaders != nil {
			for field, values := range writeHeaders {
				if len(values) > 0 {
					for index, value := range values {
						if index == 0 {
							w.Header().Set(field, value)
						} else {
							w.Header().Add(field, value)
						}
					}
				}
			}
		}

		fmt.Println(fileHandler)

		io.Copy(w, fileHandler)

		return nil
	}
}

// Helpers
func isStringInArray(method string, into []string) bool {
	for _, item := range into {
		if item == method {
			return true
		}
	}
	return false
}
