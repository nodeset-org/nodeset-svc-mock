package db

import (
	"log/slog"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/test_utils"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ()

func TestDatabaseClone(t *testing.T) {
	logger := slog.Default()

	// Create a new database
	db := NewDatabase(logger)

	// Add a StakeWise vault to the database
	err := db.AddStakeWiseVault(test_utils.StakeWiseVaultAddress, test_utils.Network)
	if err != nil {
		t.Fatalf("Error adding StakeWise vault to database: %v", err)
	}
	t.Log("Added StakeWise vault to database")

	// Add a users to the database
	addUserToDatabase(t, db, test_utils.User0Email)
	addUserToDatabase(t, db, test_utils.User1Email)
	addUserToDatabase(t, db, test_utils.User2Email)
	addUserToDatabase(t, db, test_utils.User3Email)
	t.Log("Added users to database")

	// Add nodes to the user
	node0 := createNodeAndAddToDatabase(t, db, test_utils.User1Email, 0)
	node1 := createNodeAndAddToDatabase(t, db, test_utils.User2Email, 1)
	node2 := createNodeAndAddToDatabase(t, db, test_utils.User3Email, 2)
	node3 := createNodeAndAddToDatabase(t, db, test_utils.User3Email, 3)
	t.Log("Added nodes to users")

	// Get some deposit data
	depositData0 := generateDepositData(t, 0, node0)
	depositData1 := generateDepositData(t, 1, node1)
	depositData2 := generateDepositData(t, 2, node1)
	depositData3 := generateDepositData(t, 3, node2)
	depositData4 := generateDepositData(t, 4, node3)
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

	// Create a new set with 1 DD per user and verify (note since we're just using regular maps and not preserving
	// upload order, we have to make sure there's only one deposit data per user but don't know which we'll get)
	depositDataSet := db.CreateNewDepositDataSet(test_utils.Network, 1)
	require.Equal(t, 3, len(depositDataSet))
	contains := []bool{false, false, false, false, false}
	for _, dd := range depositDataSet {
		pubkey := beacon.ValidatorPubkey(dd.PublicKey)
		switch pubkey {
		case beacon.ValidatorPubkey(depositData0.PublicKey):
			contains[0] = true
		case beacon.ValidatorPubkey(depositData1.PublicKey):
			contains[1] = true
		case beacon.ValidatorPubkey(depositData2.PublicKey):
			contains[2] = true
		case beacon.ValidatorPubkey(depositData3.PublicKey):
			contains[3] = true
		case beacon.ValidatorPubkey(depositData4.PublicKey):
			contains[4] = true
		}
	}
	assert.Equal(t, true, contains[0])
	assert.Equal(t, true, contains[1] != contains[2]) // 1 or 2, but not both
	assert.Equal(t, true, contains[3] != contains[4]) // 3 or 4, but not both

	// Handle the deposit data upload
	err = db.UploadDepositDataToStakeWise(test_utils.StakeWiseVaultAddress, test_utils.Network, depositDataSet)
	if err != nil {
		t.Fatalf("Error uploading deposit data to StakeWise: %v", err)
	}
	t.Log("Uploaded deposit data to StakeWise")

	// Finalize the upload
	db.MarkDepositDataSetUploaded(depositDataSet)
	t.Log("Marked deposit data set uploaded")

	// Clone the database
	clone := db.Clone()
	t.Log("Cloned database")

	// Check the clone is not the same as the original
	if clone == db {
		t.Fatalf("Clone is the same as the original database")
	}
	compareDatabases(t, db, clone)
	if t.Failed() {
		return
	}

	// Mark an unused deposit data as uploaded for the StakeWise vault
	pubkey := beacon.ValidatorPubkey(depositData2.PublicKey)
	if db.StakeWiseVaults[test_utils.Network][test_utils.StakeWiseVaultAddress].UploadedData[pubkey] {
		// Swap to pubkey 1 if 2 is used
		pubkey = beacon.ValidatorPubkey(depositData1.PublicKey)
	}
	assert.Equal(t, false, db.StakeWiseVaults[test_utils.Network][test_utils.StakeWiseVaultAddress].UploadedData[pubkey])
	db.StakeWiseVaults[test_utils.Network][test_utils.StakeWiseVaultAddress].MarkDepositDataUploaded(pubkey)
	t.Log("Marked deposit data uploaded for StakeWise vault")

	// Make sure the clone didn't get the update
	if clone.StakeWiseVaults[test_utils.Network][test_utils.StakeWiseVaultAddress].UploadedData[pubkey] {
		t.Fatalf("Clone got the update")
	}
	t.Log("Clone wasn't updated, as expected")
}

func addUserToDatabase(t *testing.T, db *Database, userEmail string) {
	err := db.AddUser(userEmail)
	if err != nil {
		t.Fatalf("Error adding user [%s] to database: %v", userEmail, err)
	}
}

func createNodeAndAddToDatabase(t *testing.T, db *Database, userEmail string, index uint) common.Address {
	nodeKey, err := test_utils.GetEthPrivateKey(index)
	if err != nil {
		t.Fatalf("Error getting private key for node 0: %v", err)
	}
	nodeAddress := crypto.PubkeyToAddress(nodeKey.PublicKey)
	err = db.AddNodeAccount(userEmail, nodeAddress)
	if err != nil {
		t.Fatalf("Error adding node [%s] to user [%s]: %v", nodeAddress.Hex(), userEmail, err)
	}
	return nodeAddress
}

func generateDepositData(t *testing.T, index uint, withdrawalAddress common.Address) beacon.ExtendedDepositData {
	validatorKey, err := test_utils.GetBeaconPrivateKey(index)
	if err != nil {
		t.Fatalf("Error getting private key for validator %d: %v", index, err)
	}
	depositData, err := validator.GetDepositData(
		validatorKey,
		validator.GetWithdrawalCredsFromAddress(withdrawalAddress),
		test_utils.GenesisForkVersion,
		test_utils.DepositAmount,
		test_utils.Network,
	)
	if err != nil {
		t.Fatalf("Error generating deposit data for validator %d: %v", index, err)
	}
	return depositData
}

func compareDatabases(t *testing.T, db *Database, clone *Database) {
	// Compare DD sets
	assert.Equal(t, db.LatestDepositDataSetIndex, clone.LatestDepositDataSetIndex)
	assert.Equal(t, db.LatestDepositDataSet, clone.LatestDepositDataSet)
	assert.NotSame(t, &db.LatestDepositDataSet, &clone.LatestDepositDataSet)

	// Compare StakeWise vault networks
	assert.Equal(t, len(db.StakeWiseVaults), len(clone.StakeWiseVaults))
	for network, vaults := range db.StakeWiseVaults {
		// Compare vaults in this network
		cloneVaults, exists := clone.StakeWiseVaults[network]
		if !exists {
			t.Errorf("Expected vault network [%s] in clone, got none", network)
		}
		assert.Equal(t, len(vaults), len(cloneVaults))
		for address, vault := range vaults {
			// Compare this vault
			cloneVault, exists := cloneVaults[address]
			if !exists {
				t.Errorf("Expected vault address [%s] in clone, got none", address.Hex())
			}
			assert.NotSame(t, vault, cloneVault)
			assert.Equal(t, vault, cloneVault)
		}
	}

	// Compare users
	assert.Equal(t, len(db.Users), len(clone.Users))
	for email, user := range db.Users {
		// Compare this user
		cloneUser, exists := clone.Users[email]
		if !exists {
			t.Errorf("Expected user with email [%s] in clone, got none", email)
		}
		assert.NotSame(t, user, cloneUser)
		assert.Equal(t, user.Email, cloneUser.Email)
		assert.Equal(t, len(user.Nodes), len(cloneUser.Nodes))
		for address, node := range user.Nodes {
			// Compare this node
			cloneNode, exists := cloneUser.Nodes[address]
			if !exists {
				t.Errorf("Expected node address [%s] in clone, got none", address.Hex())
			}
			assert.NotSame(t, node, cloneNode)
			assert.Equal(t, node.Address, cloneNode.Address)
			assert.Equal(t, len(node.Validators), len(cloneNode.Validators))
			for network, validators := range node.Validators {
				// Compare validators in this network
				cloneValidators, exists := cloneNode.Validators[network]
				if !exists {
					t.Errorf("Expected validator network [%s] in clone, got none", network)
				}
				assert.Equal(t, len(validators), len(cloneValidators))
				for pubkey, validator := range validators {
					// Compare this validator
					cloneValidator, exists := cloneValidators[pubkey]
					if !exists {
						t.Errorf("Expected validator pubkey [%s] in clone, got none", pubkey.Hex())
					}
					assert.NotSame(t, validator, cloneValidator)
					assert.Equal(t, validator, cloneValidator)
				}
			}
		}
	}
}
