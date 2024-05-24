package server

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	idb "github.com/nodeset-org/nodeset-svc-mock/internal/db"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

// Make sure signed exits are uploaded correctly
func TestUploadSignedExits(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	db := idb.ProvisionFullDatabase(t, logger, false)
	server.manager.SetDatabase(db)

	// Run a get deposit data request to make sure it's empty
	parsedResponse := runGetDepositDataRequest(t)
	require.Equal(t, 0, parsedResponse.Version)
	require.Empty(t, parsedResponse.Data)

	// Generate new deposit data
	nodeKey := idb.NodeKeys[0]
	depositData := []beacon.ExtendedDepositData{
		idb.GenerateDepositData(t, 0, test.StakeWiseVaultAddress),
		idb.GenerateDepositData(t, 1, test.StakeWiseVaultAddress),
	}
	t.Log("Generated deposit data")

	// Run an upload deposit data request
	runUploadDepositDataRequest(t, nodeKey, depositData)

	// Run a get deposit data request to make sure it's uploaded
	validatorsResponse := runGetValidatorsRequest(t, nodeKey)
	expectedData := []api.ValidatorStatus{
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[0].PublicKey),
			Status:              string(api.StakeWiseStatus_Pending),
			ExitMessageUploaded: false,
		},
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[1].PublicKey),
			Status:              string(api.StakeWiseStatus_Pending),
			ExitMessageUploaded: false,
		},
	}
	require.Equal(t, expectedData, validatorsResponse.Data)
	t.Logf("Received matching response")

	// Generate a signed exit for validator 1
	signedExit1 := idb.GenerateSignedExit(t, 1)
	t.Log("Generated signed exit")

	// Upload it
	runUploadSignedExitsRequest(t, nodeKey, []api.ExitData{signedExit1})
	t.Logf("Uploaded signed exit")

	// Get the validator status again
	validatorsResponse = runGetValidatorsRequest(t, nodeKey)
	expectedData = []api.ValidatorStatus{
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[0].PublicKey),
			Status:              string(api.StakeWiseStatus_Pending),
			ExitMessageUploaded: false,
		},
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[1].PublicKey),
			Status:              string(api.StakeWiseStatus_Pending),
			ExitMessageUploaded: true, // This should be true now
		},
	}
	require.Equal(t, expectedData, validatorsResponse.Data)
	t.Logf("Received matching response")
}

func runUploadSignedExitsRequest(t *testing.T, nodeKey *ecdsa.PrivateKey, signedExits []api.ExitData) {
	// Marshal the deposit data
	body, err := json.Marshal(signedExits)
	if err != nil {
		t.Fatalf("error marshalling signed exits: %v", err)
	}

	// Create the request
	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:%d/api/%s", port, api.ValidatorsPath), strings.NewReader(string(body)))
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
	t.Logf("Received an OK status code")
}
