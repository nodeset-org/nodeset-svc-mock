package db

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	stakeWiseVaultAddressHex string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	network                  string = "holesky"
	userEmail                string = "test@test.com"
	nodeAddress0Hex          string = "0x90de00eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	nodeAddress1Hex          string = "0x90de01eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	pubkeyHex                string = "0xbeac09bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestDatabaseClone(t *testing.T) {
	// Create a new database
	db := NewDatabase()

	// Add a StakeWise vault to the database
	stakeWiseVaultAddress := common.HexToAddress(stakeWiseVaultAddressHex)
	err := db.AddStakeWiseVault(stakeWiseVaultAddress, network)
	if err != nil {
		t.Fatalf("Error adding StakeWise vault to database: %v", err)
	}
	t.Log("Added StakeWise vault to database")

	// Add a user to the database
	err = db.AddUser(userEmail)
	if err != nil {
		t.Fatalf("Error adding user to database: %v", err)
	}
	t.Log("Added user to database")

	// Add nodes to the user
	errs := []error{
		addNodeToDatabase(db, userEmail, nodeAddress0Hex),
		addNodeToDatabase(db, userEmail, nodeAddress1Hex),
	}
	if err := errors.Join(errs...); err != nil {
		t.Fatalf("Error adding nodes to database: %v", err)
	}
	t.Log("Added nodes to user")

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

	// Mark deposit data uploaded for the StakeWise vault
	pubkey, err := beacon.HexToValidatorPubkey(pubkeyHex)
	if err != nil {
		t.Fatalf("Error parsing pubkey: %v", err)
	}
	db.StakeWiseVaults[network][stakeWiseVaultAddress].MarkDepositDataUploaded(pubkey)
	t.Log("Marked deposit data uploaded for StakeWise vault")

	// Make sure the clone didn't get the update
	if clone.StakeWiseVaults[network][stakeWiseVaultAddress].UploadedData[pubkey] {
		t.Fatalf("Clone got the update")
	}
	t.Log("Clone wasn't updated, as expected")
}

func addNodeToDatabase(db *Database, userEmail string, nodeAddressHex string) error {
	nodeAddress := common.HexToAddress(nodeAddressHex)
	err := db.AddNodeAccount(userEmail, nodeAddress)
	if err != nil {
		return fmt.Errorf("Error adding node [%s] to user [%s]: %w", nodeAddressHex, userEmail, err)
	}
	return nil
}

func compareDatabases(t *testing.T, db *Database, clone *Database) {
	// Compare StakeWise vault networks
	if len(clone.StakeWiseVaults) != len(db.StakeWiseVaults) {
		t.Errorf("Original: %d vault networks, Clone: %d vault networks,", len(db.StakeWiseVaults), len(clone.StakeWiseVaults))
	}
	for network, vaults := range db.StakeWiseVaults {
		// Compare vaults in this network
		cloneVaults, exists := clone.StakeWiseVaults[network]
		if !exists {
			t.Errorf("Expected vault network [%s] in clone, got none", network)
		}
		if len(cloneVaults) != len(vaults) {
			t.Errorf("Original: %d vaults in network [%s], Clone: %d vaults", len(vaults), network, len(cloneVaults))
		}
		for address, vault := range vaults {
			// Compare this vault
			cloneVault, exists := cloneVaults[address]
			if !exists {
				t.Errorf("Expected vault address [%s] in clone, got none", address.Hex())
			}
			if cloneVault == vault {
				t.Errorf("Expected vault to be different from original")
			}
			if cloneVault.Address != vault.Address {
				t.Errorf("Expected vault address [%s], got [%s]", vault.Address.Hex(), cloneVault.Address.Hex())
			}
			if len(cloneVault.UploadedData) != len(vault.UploadedData) {
				t.Errorf("Original: %d uploaded data entries, Clone: %d uploaded data entries", len(vault.UploadedData), len(cloneVault.UploadedData))
			}
			for pubkey, uploaded := range vault.UploadedData {
				// Compare this uploaded data entry
				cloneUploaded, exists := cloneVault.UploadedData[pubkey]
				if !exists {
					t.Errorf("Expected uploaded data entry for pubkey [%s] in clone, got none", pubkey.Hex())
				}
				if cloneUploaded != uploaded {
					t.Errorf("Expected uploaded data entry for pubkey [%s] to be [%t], got [%t]", pubkey.Hex(), uploaded, cloneUploaded)
				}
			}
		}
	}

	// Compare users
	if len(clone.Users) != len(db.Users) {
		t.Errorf("Original: %d users, Clone: %d users", len(db.Users), len(clone.Users))
	}
	for email, user := range db.Users {
		// Compare this user
		cloneUser, exists := clone.Users[email]
		if !exists {
			t.Errorf("Expected user with email [%s] in clone, got none", email)
		}
		if cloneUser == user {
			t.Errorf("Expected user to be different from original")
		}
		if cloneUser.Email != user.Email {
			t.Errorf("Expected user email [%s], got [%s]", user.Email, cloneUser.Email)
		}
		if len(cloneUser.Nodes) != len(user.Nodes) {
			t.Errorf("Original: %d nodes, Clone: %d nodes", len(user.Nodes), len(cloneUser.Nodes))
		}
		for address, node := range user.Nodes {
			// Compare this node
			cloneNode, exists := cloneUser.Nodes[address]
			if !exists {
				t.Errorf("Expected node address [%s] in clone, got none", address.Hex())
			}
			if cloneNode == node {
				t.Errorf("Expected node to be different from original")
			}
			if cloneNode.Address != node.Address {
				t.Errorf("Expected node address [%s], got [%s]", node.Address.Hex(), cloneNode.Address.Hex())
			}
			if len(cloneNode.Validators) != len(node.Validators) {
				t.Errorf("Original: %d validator networks, Clone: %d validator networks", len(node.Validators), len(cloneNode.Validators))
			}
			for network, validators := range node.Validators {
				// Compare validators in this network
				cloneValidators, exists := cloneNode.Validators[network]
				if !exists {
					t.Errorf("Expected validator network [%s] in clone, got none", network)
				}
				if len(cloneValidators) != len(validators) {
					t.Errorf("Original: %d validators in network [%s], Clone: %d validators", len(validators), network, len(cloneValidators))
				}
				for pubkey, validator := range validators {
					// Compare this validator
					cloneValidator, exists := cloneValidators[pubkey]
					if !exists {
						t.Errorf("Expected validator pubkey [%s] in clone, got none", pubkey.Hex())
					}
					if cloneValidator == validator {
						t.Errorf("Expected validator to be different from original")
					}
					if cloneValidator.Pubkey != validator.Pubkey {
						t.Errorf("Expected validator pubkey [%s], got [%s]", validator.Pubkey.Hex(), cloneValidator.Pubkey.Hex())
					}
					if !reflect.DeepEqual(cloneValidator.DepositData, validator.DepositData) {
						t.Errorf("Expected validator deposit data [%v], got [%v]", validator.DepositData, cloneValidator.DepositData)
					}
					if !reflect.DeepEqual(cloneValidator.SignedExit, validator.SignedExit) {
						t.Errorf("Expected validator signed exit [%v], got [%v]", validator.SignedExit, cloneValidator.SignedExit)
					}
					if cloneValidator.DepositDataUsed != validator.DepositDataUsed {
						t.Errorf("Expected validator deposit data used to be [%t], got [%t]", validator.DepositDataUsed, cloneValidator.DepositDataUsed)
					}
					if cloneValidator.ExitMessageUploaded != validator.ExitMessageUploaded {
						t.Errorf("Expected exit message uploaded to be [%t], got [%t]", validator.ExitMessageUploaded, cloneValidator.ExitMessageUploaded)
					}
				}
			}
		}
	}
}
