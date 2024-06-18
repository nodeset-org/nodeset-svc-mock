package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// Make sure node registration works properly
func TestRegisterNode(t *testing.T) {
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

	// Create the registration request
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	require.NoError(t, err)
	regReq := api.RegisterNodeRequest{
		Email:       test.User0Email,
		NodeAddress: node0Pubkey.Hex(),
		Signature:   utils.EncodeHexWithPrefix(regSig),
	}
	body, err := json.Marshal(regReq)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/api/%s", port, api.RegisterPath), bytes.NewReader(body))
	require.NoError(t, err)
	t.Logf("Created request")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	t.Logf("Sent request")

	// Check the status code
	require.Equal(t, http.StatusOK, response.StatusCode)
	t.Logf("Received OK status code")
}
