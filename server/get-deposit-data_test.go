package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/db"
	idb "github.com/nodeset-org/nodeset-svc-mock/internal/db"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestGetDepositData(t *testing.T) {
	// Take a snapshot
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	db := idb.ProvisionFullDatabase(t, logger, true)
	server.manager.SetDatabase(db)

	// Run a get deposit data request
	parsedResponse := runGetDepositDataRequest(t, db.Sessions[0])

	// Make sure the response is correct
	vault := db.StakeWiseVaults[test.Network][0]
	require.Equal(t, vault.LatestDepositDataSetIndex, parsedResponse.Data.Version)
	require.Equal(t, vault.LatestDepositDataSet, parsedResponse.Data.DepositData)
	require.Greater(t, len(parsedResponse.Data.DepositData), 0)
	t.Logf("Received correct response - version = %d, deposit data matches", parsedResponse.Data.Version)
}

func runGetDepositDataRequest(t *testing.T, session *db.Session) api.NodeSetResponse[api.DepositDataData] {
	// Create the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/%s", port, api.DepositDataPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("vault", utils.RemovePrefix(strings.ToLower(test.StakeWiseVaultAddressHex)))
	query.Add("network", test.Network)
	request.URL.RawQuery = query.Encode()
	t.Logf("Created request")

	// Add the auth header
	auth.AddAuthorizationHeader(request, session)
	t.Logf("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("error reading the response body: %v", err)
	}
	var parsedResponse api.NodeSetResponse[api.DepositDataData]
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}
	t.Log("Received response")
	return parsedResponse
}
