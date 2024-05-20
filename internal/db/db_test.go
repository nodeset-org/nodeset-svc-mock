package db

import (
	"log/slog"
	"testing"

	"github.com/nodeset-org/nodeset-svc-mock/db"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseClone(t *testing.T) {
	// Set up a database
	logger := slog.Default()
	db := ProvisionFullDatabase(t, logger, true)

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
	t.Log("Clone has identical contents to the original database but different pointers")

	// Get the first pubkey from user 2 that hasn't been uploaded yet
	user2 := db.Users[test.User2Email]
	vault := db.StakeWiseVaults[test.Network][test.StakeWiseVaultAddress]
	var pubkey beacon.ValidatorPubkey
	found := false
	for _, node := range user2.Nodes {
		validators := node.Validators[test.Network]
		for _, validator := range validators {
			if vault.UploadedData[validator.Pubkey] {
				continue
			}
			pubkey = validator.Pubkey
			found = true
			break
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatalf("Couldn't find a pubkey to test with")
	}
	t.Logf("Using pubkey %s for testing", pubkey.HexWithPrefix())

	// Mark the pubkey as uploaded in the original database
	assert.Equal(t, false, db.StakeWiseVaults[test.Network][test.StakeWiseVaultAddress].UploadedData[pubkey])
	db.StakeWiseVaults[test.Network][test.StakeWiseVaultAddress].MarkDepositDataUploaded(pubkey)
	t.Log("Marked deposit data uploaded for StakeWise vault")

	// Make sure the clone didn't get the update
	if clone.StakeWiseVaults[test.Network][test.StakeWiseVaultAddress].UploadedData[pubkey] {
		t.Fatalf("Clone got the update")
	}
	t.Log("Clone wasn't updated, as expected")
}

// ==========================
// === Internal Functions ===
// ==========================

// Compare two databases
func compareDatabases(t *testing.T, db *db.Database, clone *db.Database) {
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
	assert.Equal(t, db.Users, clone.Users)

	// Make sure the user pointers are all different
	for email, user := range db.Users {
		cloneUser := clone.Users[email]
		assert.NotSame(t, user, cloneUser)
		for address, node := range user.Nodes {
			cloneNode := cloneUser.Nodes[address]
			assert.NotSame(t, node, cloneNode)
			for network, validators := range node.Validators {
				cloneValidators := cloneNode.Validators[network]
				for pubkey, validator := range validators {
					cloneValidator := cloneValidators[pubkey]
					assert.NotSame(t, validator, cloneValidator)
				}
			}
		}
	}
}
