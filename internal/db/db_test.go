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
	user2 := db.Users[2]
	vault := db.StakeWiseVaults[test.Network][0]
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
	assert.Equal(t, false, db.StakeWiseVaults[test.Network][0].UploadedData[pubkey])
	db.StakeWiseVaults[test.Network][0].MarkDepositDataUploaded(pubkey)
	t.Log("Marked deposit data uploaded for StakeWise vault")

	// Make sure the clone didn't get the update
	if clone.StakeWiseVaults[test.Network][0].UploadedData[pubkey] {
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
	assert.Equal(t, db.StakeWiseVaults, clone.StakeWiseVaults)
	for network, vaults := range db.StakeWiseVaults {
		cloneVaults := clone.StakeWiseVaults[network]
		for i, vault := range vaults {
			cloneVault := cloneVaults[i]
			assert.NotSame(t, vault, cloneVault)
		}
	}

	// Compare users
	assert.Equal(t, db.Users, clone.Users)

	// Make sure the user pointers are all different
	for i, user := range db.Users {
		cloneUser := clone.Users[i]
		assert.NotSame(t, user, cloneUser)
		for j, node := range user.Nodes {
			cloneNode := cloneUser.Nodes[j]
			assert.NotSame(t, node, cloneNode)
			for k, validators := range node.Validators {
				cloneValidators := cloneNode.Validators[k]
				for l, validator := range validators {
					cloneValidator := cloneValidators[l]
					assert.NotSame(t, validator, cloneValidator)
				}
			}
		}
	}
}
