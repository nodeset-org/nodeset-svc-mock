package db

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/types"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Mock database for storing nodeset.io info
type Database struct {
	// Collection of StakeWise vaults
	StakeWiseVaults map[string]map[common.Address]*StakeWiseVault

	// Collection of users
	Users map[string]*User

	// Latest deposit data set uploaded to StakeWise
	LatestDepositDataSet uint
}

// Creates a new database
func NewDatabase() *Database {
	return &Database{
		StakeWiseVaults: map[string]map[common.Address]*StakeWiseVault{},
		Users:           map[string]*User{},
	}
}

// Adds a StakeWise vault to the database
func (d *Database) AddStakeWiseVault(address common.Address, networkName string) error {
	network, exists := d.StakeWiseVaults[networkName]
	if !exists {
		network = map[common.Address]*StakeWiseVault{}
		d.StakeWiseVaults[networkName] = network
	}

	if _, exists := network[address]; exists {
		return fmt.Errorf("stakewise vault with address [%s] already exists", address.Hex())
	}

	vault := NewStakeWiseVaultInfo(address)
	network[address] = vault
	return nil
}

// Adds a user to the database
func (d *Database) AddUser(email string) error {
	if _, exists := d.Users[email]; exists {
		return fmt.Errorf("user with email [%s] already exists", email)
	}

	user := NewUser(email)
	d.Users[email] = user
	return nil
}

// Registers a node with a user
func (d *Database) AddNodeAccount(email string, nodeAddress common.Address) error {
	for _, user := range d.Users {
		if user.Email != email {
			continue
		}
		user.AddNode(nodeAddress)
		return nil
	}

	return fmt.Errorf("user with email [%s] not found", email)
}

// Clones the database
func (d *Database) Clone() *Database {
	clone := NewDatabase()
	clone.LatestDepositDataSet = d.LatestDepositDataSet

	// Copy StakeWise vaults
	for network, vaults := range d.StakeWiseVaults {
		for address, vault := range vaults {
			clone.AddStakeWiseVault(address, network)
			clone.StakeWiseVaults[network][address] = vault.Clone()
		}
	}

	// Copy users
	for email, user := range d.Users {
		clone.AddUser(email)
		clone.Users[email] = user.Clone()
	}
	return clone
}

// ==========================

// Handle a new collection of deposit data uploads from a node
func (d *Database) HandleDepositDataUpload(nodeAddress common.Address, data []beacon.ExtendedDepositData) error {
	// Get the node
	var node *Node
	for _, user := range d.Users {
		for candidateAddress, candidate := range user.Nodes {
			if candidateAddress == nodeAddress {
				node = candidate
				break
			}
		}
		if node != nil {
			break
		}
	}
	if node == nil {
		return fmt.Errorf("node with address [%s] not found", nodeAddress.Hex())
	}

	// Add the deposit data
	for _, depositData := range data {
		node.AddDepositData(depositData)
	}
	return nil
}

// Handle a new collection of signed exits from a node
func (d *Database) HandleSignedExitUpload(nodeAddress common.Address, network string, data []types.ExitData) error {
	// Get the node
	var node *Node
	for _, user := range d.Users {
		for candidateAddress, candidate := range user.Nodes {
			if candidateAddress == nodeAddress {
				node = candidate
				break
			}
		}
		if node != nil {
			break
		}
	}
	if node == nil {
		return fmt.Errorf("node with address [%s] not found", nodeAddress.Hex())
	}

	// Add the signed exits
	for _, signedExit := range data {
		pubkey, err := beacon.HexToValidatorPubkey(signedExit.Pubkey)
		if err != nil {
			return fmt.Errorf("error parsing validator pubkey [%s]: %w", signedExit.Pubkey, err)
		}

		// Get the validator
		validatorMap, exists := node.Validators[network]
		if !exists {
			return fmt.Errorf("network [%s] is not used by node [%s]", network, nodeAddress.Hex())
		}
		validator, exists := validatorMap[pubkey]
		if !exists {
			return fmt.Errorf("node [%s] doesn't have validator [%s]", nodeAddress.Hex(), pubkey.Hex())
		}

		validator.SetExitMessage(signedExit.ExitMessage)
	}
	return nil
}

func (d *Database) UploadDepositDataToStakeWise(vaultAddress common.Address, network string, data []beacon.ExtendedDepositData) error {
	vault, exists := d.StakeWiseVaults[network][vaultAddress]
	if !exists {
		return fmt.Errorf("vault with address [%s] not found", vaultAddress.Hex())
	}

	for _, depositData := range data {
		pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
		vault.MarkDepositDataUploaded(pubkey)
	}
	return nil
}
