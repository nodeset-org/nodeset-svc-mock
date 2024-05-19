package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"testing"

	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/test_utils"
	"github.com/stretchr/testify/require"
)

// Check for a 404 if requesting an unknown route
func TestUnknownRoute(t *testing.T) {
	// Create the server
	logger := slog.Default()
	server, err := NewNodeSetMockServer(logger, "localhost", 0)
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}
	t.Log("Created server")

	// Start it
	wg := &sync.WaitGroup{}
	err = server.Start(wg)
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	port := server.GetPort()
	t.Logf("Started server on port %d", port)

	// Send a message without an auth header
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/unknown_route", port), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Logf("Created request")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Check the response
	require.Equal(t, http.StatusNotFound, response.StatusCode)
	t.Logf("Received not found status code")

	// Stop the server
	server.Stop()
	wg.Wait()
	t.Log("Stopped server")
}

// Check for a 401 if the auth header's missing
func TestMissingHeader(t *testing.T) {
	// Create the server
	logger := slog.Default()
	server, err := NewNodeSetMockServer(logger, "localhost", 0)
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}
	t.Log("Created server")

	// Start it
	wg := &sync.WaitGroup{}
	err = server.Start(wg)
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	port := server.GetPort()
	t.Logf("Started server on port %d", port)

	// Send a message without an auth header
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/%s", port, api.DepositDataMetaPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Logf("Created request")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Check the response
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	t.Logf("Received unauthorized status code")

	// Stop the server
	server.Stop()
	wg.Wait()
	t.Log("Stopped server")
}

// Check for a 401 if the node isn't registered
func TestUnregisteredNode(t *testing.T) {
	// Create the server
	logger := slog.Default()
	server, err := NewNodeSetMockServer(logger, "localhost", 0)
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}
	t.Log("Created server")

	// Start it
	wg := &sync.WaitGroup{}
	err = server.Start(wg)
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	port := server.GetPort()
	t.Logf("Started server on port %d", port)

	// Send a message without an auth header
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/%s", port, api.DepositDataMetaPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Logf("Created request")

	// Add an auth header
	node0Key, err := test_utils.GetPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}
	auth.AddAuthorizationHeader(request, node0Key)
	t.Logf("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Check the response
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	t.Logf("Received unauthorized status code")

	// Stop the server
	server.Stop()
	wg.Wait()
	t.Log("Stopped server")
}
