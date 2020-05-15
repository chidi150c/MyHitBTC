package app

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

// Server represents an HTTP server.
type Server struct {
	ln net.Listener

	// Handler to serve.
	Handler TradeHandler

	// Bind address to open.
	Addr string
}

func NewServer(addre string, h TradeHandler) *Server {
	return &Server{
		Handler: h,
		Addr:    ":" + addre,
	}
}

// Open opens a socket and serves the HTTP server.
//,
func (s *Server) Open() error {
	//_, _ = done, sigs

	//[to make an app engine app uncomment the next two lines]
	// http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, s.Handler))
	// appengine.Main()

	// Open socket.
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.ln = ln
	fmt.Println("Listening on: ", s.Addr)
	log.Fatal(http.Serve(s.ln, s.Handler))
	//log.Fatal(http.Serve(s.ln, handlers.CombinedLoggingHandler(os.Stderr, s.Handler)))
	return nil
}

// Close closes the socket.
func (s *Server) Close() error {
	if s.ln != nil {
		s.ln.Close()
	}
	return nil
}
