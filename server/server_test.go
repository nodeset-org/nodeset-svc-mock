package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/stretchr/testify/require"
)

// Various singleton variables used for testing
var (
	logger *slog.Logger       = slog.Default()
	server *NodeSetMockServer = nil
	wg     *sync.WaitGroup    = nil
	port   uint16             = 0
)

// Initialize a common server used by all tests
func TestMain(m *testing.M) {
	// Create the server
	var err error
	server, err = NewNodeSetMockServer(logger, "localhost", 0)
	if err != nil {
		fail("error creating server: %v", err)
	}
	logger.Info("Created server")

	// Start it
	wg = &sync.WaitGroup{}
	err = server.Start(wg)
	if err != nil {
		fail("error starting server: %v", err)
	}
	port = server.GetPort()
	logger.Info(fmt.Sprintf("Started server on port %d", port))

	// Run tests
	code := m.Run()

	// Revert to the baseline after testing is done
	cleanup()

	// Done
	os.Exit(code)
}

func fail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logger.Error(msg)
	cleanup()
	os.Exit(1)
}

func cleanup() {
	if server != nil {
		_ = server.Stop()
		wg.Wait()
		logger.Info("Stopped server")
	}
}

// =============
// === Tests ===
// =============

// Check for a 404 if requesting an unknown route
func TestUnknownRoute(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer server.manager.RevertToSnapshot("test")

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
}

// Check for a 401 if the auth header's missing
func TestMissingHeader(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer server.manager.RevertToSnapshot("test")

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
}

// Check for a 401 if the node isn't registered
func TestUnregisteredNode(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer server.manager.RevertToSnapshot("test")

	// Send a message without an auth header
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/%s", port, api.DepositDataMetaPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Logf("Created request")

	// Add an auth header
	node0Key, err := test.GetEthPrivateKey(0)
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
}
