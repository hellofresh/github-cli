package test

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
)

// Handler is the interface used by httpmock instead of http.Handler so that it can be mocked very easily.
type Handler interface {
	Handle(method, path string, header *http.Header, body []byte) Response
}

// Response holds the response a handler wants to return to the client.
type Response struct {
	// The HTTP status code to write (default: 200)
	Status int
	// Headers to add to the response
	Header http.Header
	// The response body to write (default: no body)
	Body []byte
}

// Server listens for requests and interprets them into calls to your Handler.
type Server struct {
	httpServer *httptest.Server
	handler    Handler
}

// NewServer constructs a new server and starts it (compare to httptest.NewServer). It needs to be Closed()ed.
func NewServer(handler Handler) *Server {
	s := NewUnstartedServer(handler)
	s.Start()
	return s
}

// NewUnstartedServer constructs a new server but doesn't start it (compare to httptest.NewUnstartedServer).
func NewUnstartedServer(handler Handler) *Server {
	return &Server{
		handler:    handler,
		httpServer: httptest.NewUnstartedServer(&httpToHTTPMockHandler{handler: handler}),
	}
}

// Start starts an unstarted server.
func (s *Server) Start() {
	s.httpServer.Start()
}

// Close shuts down a started server.
func (s *Server) Close() {
	s.httpServer.Close()
}

// URL is the URL for the local test server, i.e. the value of httptest.Server.URL
func (s *Server) URL() string {
	return s.httpServer.URL
}

// httpToHTTPMockHandler is a normal http.Handler that converts the request into a test.Handler call and calls the
// httmock handler.
type httpToHTTPMockHandler struct {
	handler Handler
}

// ServeHTTP makes this implement http.Handler
func (h *httpToHTTPMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read HTTP body in httpmock: %v", err)
	}
	resp := h.handler.Handle(r.Method, r.URL.RequestURI(), &r.Header, body)

	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	status := resp.Status
	if status == 0 {
		status = 200
	}
	w.WriteHeader(status)
	_, err = w.Write(resp.Body)
	if err != nil {
		log.Printf("Failed to write response in httpmock: %v", err)
	}
}
