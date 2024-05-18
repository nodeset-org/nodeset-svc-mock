package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/log"
)

type NodeSetMockServer struct {
	logger *slog.Logger
	ip     string
	port   uint16
	socket net.Listener
	server http.Server
	router *mux.Router
}

func NewNodeSetMockServer(logger *slog.Logger, ip string, port uint16) (*NodeSetMockServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	server := &NodeSetMockServer{
		logger: logger,
		ip:     ip,
		port:   port,
		router: router,
		server: http.Server{
			Handler: router,
		},
	}

	// Register each route
	//nmcRouter := router.PathPrefix("/api").Subrouter()
	// TODO: Register routes

	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *NodeSetMockServer) Start(wg *sync.WaitGroup) error {
	// Create the socket
	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Start listening
	wg.Add(1)
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error while listening for HTTP requests", log.Err(err))
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (s *NodeSetMockServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}
