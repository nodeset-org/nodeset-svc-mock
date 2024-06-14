package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/utils"
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
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

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
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

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
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Create a session
	session := server.manager.CreateSession()
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)
	t.Log("Created session and login signature")

	// Create a login request
	sig := utils.EncodeHexWithPrefix(loginSig)
	loginReq := api.LoginRequest{
		Nonce:     session.Nonce,
		Address:   node0Pubkey.Hex(),
		Signature: sig,
	}
	body, err := json.Marshal(loginReq)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/api/%s", port, api.LoginPath), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Log("Created request")

	// Add an auth header
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}
	auth.AddAuthorizationHeader(request, session)
	if err != nil {
		t.Fatalf("error adding auth header: %v", err)
	}
	t.Log("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	defer response.Body.Close()
	t.Log("Sent request")

	// Check the status code
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	t.Log("Received unauthorized status code")

	// Unmarshal into a response to make sure it returns the correct error key
	var nodesetResponse api.NodeSetResponse[api.LoginData]
	bodyBytes, err := io.ReadAll(response.Body)
	t.Logf("Read response body: %s", string(bodyBytes))

	require.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &nodesetResponse)
	require.NoError(t, err)
	require.Equal(t, unregisteredAddressKey, nodesetResponse.Error)
	t.Logf("Received correct error key (%s)", unregisteredAddressKey)
}
