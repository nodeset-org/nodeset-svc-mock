package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/test_utils"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestDepositDataMeta(t *testing.T) {
	depositDataSet := 192

	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer server.manager.RevertToSnapshot("test")

	// Provision the database
	node0Key, err := test_utils.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = server.manager.Database.AddUser(test_utils.User0Email)
	if err != nil {
		t.Fatalf("error adding user: %v", err)
	}
	err = server.manager.Database.AddNodeAccount(test_utils.User0Email, node0Pubkey)
	if err != nil {
		t.Fatalf("error adding node account: %v", err)
	}
	server.manager.Database.LatestDepositDataSetIndex = depositDataSet

	// Send a message without an auth header
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/%s", port, api.DepositDataMetaPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	t.Logf("Created request")

	// Add an auth header
	auth.AddAuthorizationHeader(request, node0Key)
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
	var parsedResponse api.DepositDataMetaResponse
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	// Make sure the response is correct
	require.Equal(t, depositDataSet, parsedResponse.Version)
	t.Logf("Received correct response - version = %d", parsedResponse.Version)
}
