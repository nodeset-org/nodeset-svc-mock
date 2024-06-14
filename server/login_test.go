package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// Make sure logging in works properly
func TestLogin(t *testing.T) {
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
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = server.manager.AddUser(test.User0Email)
	require.NoError(t, err)
	err = server.manager.WhitelistNodeAccount(test.User0Email, node0Pubkey)
	require.NoError(t, err)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = server.manager.RegisterNodeAccount(test.User0Email, node0Pubkey, regSig)
	require.NoError(t, err)

	// Create a session
	session := server.manager.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)

	// Create the request
	loginReq := api.LoginRequest{
		Nonce:     session.Nonce,
		Address:   node0Pubkey.Hex(),
		Signature: utils.EncodeHexWithPrefix(loginSig),
	}
	body, err := json.Marshal(loginReq)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/api/%s", port, api.LoginPath), bytes.NewReader(body))
	require.NoError(t, err)
	t.Logf("Created request")

	// Add the auth header
	auth.AddAuthorizationHeader(request, session)
	t.Logf("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	t.Logf("Sent request")

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	var parsedResponse api.NodeSetResponse[api.LoginData]
	err = json.Unmarshal(bytes, &parsedResponse)
	require.NoError(t, err)

	// Check the status code
	require.Equal(t, http.StatusOK, response.StatusCode)
	t.Logf("Received OK status code")

	// Make sure the response is correct
	require.Equal(t, session.Token, parsedResponse.Data.Token)
	t.Logf("Received correct response - session token = %s", session.Token)
}
