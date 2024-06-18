package db

import (
	"crypto/ecdsa"
	"log/slog"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/db"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
	types "github.com/wealdtech/go-eth2-types/v2"
)

var (
	NodeKeys   map[uint]*ecdsa.PrivateKey    = map[uint]*ecdsa.PrivateKey{}
	BeaconKeys map[uint]*types.BLSPrivateKey = map[uint]*types.BLSPrivateKey{}
)

// Create a full database for testing
func ProvisionFullDatabase(t *testing.T, logger *slog.Logger, includeDepositDataSet bool) *db.Database {
	db := db.NewDatabase(logger)

	// Add a StakeWise vault to the database
	err := db.AddStakeWiseVault(test.StakeWiseVaultAddress, test.Network)
	if err != nil {
		t.Fatalf("Error adding StakeWise vault to database: %v", err)
	}
	t.Log("Added StakeWise vault to database")

	// Add a users to the database
	addUserToDatabase(t, db, test.User0Email)
	addUserToDatabase(t, db, test.User1Email)
	addUserToDatabase(t, db, test.User2Email)
	addUserToDatabase(t, db, test.User3Email)
	t.Log("Added users to database")

	// Add nodes to the user
	node0 := createNodeAndAddToDatabase(t, db, test.User1Email, 0)
	node1 := createNodeAndAddToDatabase(t, db, test.User2Email, 1)
	node2 := createNodeAndAddToDatabase(t, db, test.User3Email, 2)
	node3 := createNodeAndAddToDatabase(t, db, test.User3Email, 3)
	t.Log("Added nodes to users")

	// Get some deposit data
	depositData0 := GenerateDepositData(t, 0, test.StakeWiseVaultAddress)
	depositData1 := GenerateDepositData(t, 1, test.StakeWiseVaultAddress)
	depositData2 := GenerateDepositData(t, 2, test.StakeWiseVaultAddress)
	depositData3 := GenerateDepositData(t, 3, test.StakeWiseVaultAddress)
	depositData4 := GenerateDepositData(t, 4, test.StakeWiseVaultAddress)
	t.Log("Generated deposit data")

	// Handle the deposit data upload
	err = db.HandleDepositDataUpload(node0, []beacon.ExtendedDepositData{depositData0})
	if err != nil {
		t.Fatalf("Error handling deposit data upload: %v", err)
	}
	err = db.HandleDepositDataUpload(node1, []beacon.ExtendedDepositData{depositData1, depositData2})
	if err != nil {
		t.Fatalf("Error handling deposit data upload: %v", err)
	}
	err = db.HandleDepositDataUpload(node2, []beacon.ExtendedDepositData{depositData3})
	if err != nil {
		t.Fatalf("Error handling deposit data upload: %v", err)
	}
	err = db.HandleDepositDataUpload(node3, []beacon.ExtendedDepositData{depositData4})
	if err != nil {
		t.Fatalf("Error handling deposit data upload: %v", err)
	}
	t.Log("Handled deposit data upload")

	// Shortcut if skipping deposit data set generation
	if !includeDepositDataSet {
		return db
	}

	// Create a new set with 1 DD per user and verify
	depositDataSet := db.CreateNewDepositDataSet(test.Network, 1)
	require.Equal(t, []beacon.ExtendedDepositData{depositData0, depositData1, depositData3}, depositDataSet)

	// Handle the deposit data upload
	err = db.UploadDepositDataToStakeWise(test.StakeWiseVaultAddress, test.Network, depositDataSet)
	if err != nil {
		t.Fatalf("Error uploading deposit data to StakeWise: %v", err)
	}
	t.Log("Uploaded deposit data to StakeWise")

	// Finalize the upload
	err = db.MarkDepositDataSetUploaded(test.StakeWiseVaultAddress, test.Network, depositDataSet)
	if err != nil {
		t.Fatalf("Error marking deposit data set uploaded: %v", err)
	}
	t.Log("Marked deposit data set uploaded")
	return db
}

// ==========================
// === Internal Functions ===
// ==========================

// Add a user to the database
func addUserToDatabase(t *testing.T, db *db.Database, userEmail string) {
	err := db.AddUser(userEmail)
	if err != nil {
		t.Fatalf("Error adding user [%s] to database: %v", userEmail, err)
	}
}

// Create a node, register it with the user, and log it in with a new session
func createNodeAndAddToDatabase(t *testing.T, db *db.Database, userEmail string, index uint) common.Address {
	nodeKey, exists := NodeKeys[index]
	if !exists {
		var err error
		nodeKey, err = test.GetEthPrivateKey(index)
		if err != nil {
			t.Fatalf("Error getting private key for node 0: %v", err)
		}
		NodeKeys[index] = nodeKey
	}
	nodeAddress := crypto.PubkeyToAddress(nodeKey.PublicKey)

	// Whitelist the node
	err := db.WhitelistNodeAccount(userEmail, nodeAddress)
	if err != nil {
		t.Fatalf("Error authorizing node [%s] with user [%s]: %v", nodeAddress.Hex(), userEmail, err)
	}

	// Register the node
	err = db.RegisterNodeAccount(userEmail, nodeAddress)
	if err != nil {
		t.Fatalf("Error registering node [%s] with user [%s]: %v", nodeAddress.Hex(), userEmail, err)
	}

	// Create a new session for it
	session := db.CreateSession()
	err = db.Login(nodeAddress, session.Nonce)
	if err != nil {
		t.Fatalf("Error logging in node [%s]: %v", nodeAddress.Hex(), err)
	}
	return nodeAddress
}

// Generate a validator private key and deposit data for the given index
func GenerateDepositData(t *testing.T, index uint, withdrawalAddress common.Address) beacon.ExtendedDepositData {
	validatorKey, exists := BeaconKeys[index]
	if !exists {
		var err error
		validatorKey, err = test.GetBeaconPrivateKey(index)
		if err != nil {
			t.Fatalf("Error getting private key for validator %d: %v", index, err)
		}
		BeaconKeys[index] = validatorKey
	}
	depositData, err := validator.GetDepositData(
		validatorKey,
		validator.GetWithdrawalCredsFromAddress(withdrawalAddress),
		test.GenesisForkVersion,
		test.DepositAmount,
		test.Network,
	)
	if err != nil {
		t.Fatalf("Error generating deposit data for validator %d: %v", index, err)
	}
	return depositData
}

// Generate a signed exit for the given validator index
func GenerateSignedExit(t *testing.T, index uint) api.ExitData {
	// Create the exit domain
	domain, err := types.ComputeDomain(types.DomainVoluntaryExit, test.CapellaForkVersion, test.GenesisValidatorsRoot)
	if err != nil {
		t.Fatalf("Error computing domain for validator %d: %v", index, err)
	}

	// Get the validator key
	validatorKey, exists := BeaconKeys[index]
	if !exists {
		var err error
		validatorKey, err = test.GetBeaconPrivateKey(index)
		if err != nil {
			t.Fatalf("Error getting private key for validator %d: %v", index, err)
		}
		BeaconKeys[index] = validatorKey
	}

	// Get the exit signature
	validatorIndex := strconv.FormatUint(uint64(index), 10)
	exitSignature, err := validator.GetSignedExitMessage(
		validatorKey,
		validatorIndex,
		test.ExitEpoch,
		domain,
	)
	if err != nil {
		t.Fatalf("Error generating signed exit for validator %d: %v", index, err)
	}

	// Return the exit data
	pubkey := beacon.ValidatorPubkey(validatorKey.PublicKey().Marshal())
	return api.ExitData{
		Pubkey: pubkey.HexWithPrefix(),
		ExitMessage: api.ExitMessage{
			Message: api.ExitMessageDetails{
				Epoch:          strconv.FormatUint(test.ExitEpoch, 10),
				ValidatorIndex: validatorIndex,
			},
			Signature: exitSignature.HexWithPrefix(),
		},
	}
}
