package server

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestGetValidators(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	node0Key, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = server.manager.AddUser(test.User0Email)
	if err != nil {
		t.Fatalf("error adding user: %v", err)
	}
	err = server.manager.AddNodeAccount(test.User0Email, node0Pubkey)
	if err != nil {
		t.Fatalf("error adding node account: %v", err)
	}

	// Run a get validators request
	parsedResponse := runGetValidatorsRequest(t, node0Key)

	// Make sure the response is correct
	require.Empty(t, parsedResponse.Data)
}

func runGetValidatorsRequest(t *testing.T, nodeKey *ecdsa.PrivateKey) api.ValidatorsResponse {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/%s", port, api.ValidatorsPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("network", test.Network)
	request.URL.RawQuery = query.Encode()
	t.Logf("Created request")

	// Add the auth header
	err = auth.AddAuthorizationHeader(request, nodeKey)
	if err != nil {
		t.Fatalf("error adding auth header: %v", err)
	}
	t.Logf("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Check the status code
	require.Equal(t, http.StatusOK, response.StatusCode)
	t.Logf("Received OK status code")

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("error reading the response body: %v", err)
	}
	var parsedResponse api.ValidatorsResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}
	t.Log("Received response")
	return parsedResponse
}
