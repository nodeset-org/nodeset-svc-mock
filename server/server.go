package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/db"
	"github.com/nodeset-org/nodeset-svc-mock/manager"
	"github.com/rocket-pool/node-manager-core/log"
)

type NodeSetMockServer struct {
	logger  *slog.Logger
	ip      string
	port    uint16
	socket  net.Listener
	server  http.Server
	router  *mux.Router
	manager *manager.NodeSetMockManager
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
		manager: manager.NewNodeSetMockManager(logger),
	}

	// Register each route
	nmcRouter := router.PathPrefix("/api").Subrouter()
	server.registerRoutes(nmcRouter)
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

	// Get the port if random
	if s.port == 0 {
		s.port = uint16(socket.Addr().(*net.TCPAddr).Port)
	}

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

// Get the port the server is listening on
func (s *NodeSetMockServer) GetPort() uint16 {
	return s.port
}

// Register all of the routes
func (s *NodeSetMockServer) registerRoutes(nmcRouter *mux.Router) {
	nmcRouter.HandleFunc("/"+api.DepositDataMetaPath, s.depositDataMeta)
	nmcRouter.HandleFunc("/"+api.DepositDataPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getDepositData(w, r)
		case http.MethodPost:
			s.uploadDepositData(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
}

// =============
// === Utils ===
// =============

func (s *NodeSetMockServer) processGet(w http.ResponseWriter, r *http.Request) (*db.Node, url.Values) {
	args := r.URL.Query()
	s.logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
	s.logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

	// Check the method
	if r.Method != http.MethodGet {
		handleInvalidMethod(s.logger, w)
		return nil, nil
	}

	return s.processAuthHeader(w, r), args
}

func (s *NodeSetMockServer) processPost(w http.ResponseWriter, r *http.Request, requestBody any) *db.Node {
	s.logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))

	// Check the method
	if r.Method != http.MethodPost {
		handleInvalidMethod(s.logger, w)
		return nil
	}

	// Read the body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("error reading request body: %w", err))
		return nil
	}
	s.logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

	// Deserialize the body
	err = json.Unmarshal(bodyBytes, &requestBody)
	if err != nil {
		handleInputError(s.logger, w, fmt.Errorf("error deserializing request body: %w", err))
		return nil
	}

	return s.processAuthHeader(w, r)
}

func (s *NodeSetMockServer) processAuthHeader(w http.ResponseWriter, r *http.Request) *db.Node {
	// Get the auth header
	nodeAddress, hasHeader, err := s.manager.Authorizer.VerifyRequest(r)
	if err != nil {
		handleAuthHeaderError(w, s.logger, err)
		return nil
	}
	if !hasHeader {
		handleMissingAuthHeader(w, s.logger)
		return nil
	}

	// Get the node
	node := s.manager.Database.GetNode(nodeAddress)
	if node == nil {
		handleUnregisteredNode(w, s.logger, nodeAddress)
		return nil
	}

	return node
}
