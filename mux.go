package mux

import (
	"net/http"

	"github.com/lucas-clemente/quic-go/http3"
)

// Mux type and methods
type Mux struct {
	routes         []*Route
	defaultHandler Handler
	errorHandler   ErrorHandler
}

func New() *Mux {
	var srv = &Mux{}
	return srv
}

func (s *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// try to get handler
	handler, err := s.getHandlerByRequest(r)
	if err != nil {
		s.handleError(err, w, r)
		return
	}

	// try to run handler
	if handler != nil {
		err = handler(w, r)
		if err != nil {
			s.handleError(err, w, r)
		}
		return
	}

	s.handleDefault(w, r)
}

func (s *Mux) Listen(address, pubPath, prvPath string) error {
	return http3.ListenAndServe(address, pubPath, prvPath, s)
}

// setters
func (s *Mux) AddRoute(route *Route) {
	s.routes = append(s.routes, route)
}

func (s *Mux) AddRoutes(routes []*Route) {
	for _, route := range routes {
		s.AddRoute(route)
	}
}

func (s *Mux) SetDefaultHandler(handler Handler) {
	s.defaultHandler = handler
}

func (s *Mux) SetErrorHandler(handler ErrorHandler) {
	s.errorHandler = handler
}

// private methods
func (s *Mux) getHandlerByRequest(r *http.Request) (Handler, error) {
	var lastErr error = nil
	for i := 0; i < len(s.routes); i++ {
		route := s.routes[i]
		isMatching, err := route.match(r)
		if isMatching == true {
			return route.handler, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	return nil, lastErr
}

func (s *Mux) handleDefault(w http.ResponseWriter, r *http.Request) {
	var err error
	if s.defaultHandler != nil {
		err = s.defaultHandler(w, r)
	} else {
		err = defaultHandler(w)
	}

	if err != nil {
		s.handleError(err, w, r)
	}
}

func (s *Mux) handleError(err error, w http.ResponseWriter, r *http.Request) {
	httpErr, isHttpError := err.(*HttpError)
	if !isHttpError {
		httpErr = FromError(err)
	}

	if s.errorHandler != nil {
		s.errorHandler(httpErr, w, r)
	} else {
		defaultErrorHandler(httpErr, w)
	}
}

func defaultHandler(w http.ResponseWriter) error {
	return NewHttpError(404, "The resource expected is not found")
}

func defaultErrorHandler(err *HttpError, w http.ResponseWriter) {
	err.Send(w)
}
