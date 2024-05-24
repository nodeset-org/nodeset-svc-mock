package db

import (
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Mock database for storing nodeset.io info
type Database struct {
	// Collection of StakeWise vaults
	StakeWiseVaults map[string][]*StakeWiseVault

	// Collection of users
	Users []*User

	// Internal fields
	logger *slog.Logger
}

// Creates a new database
func NewDatabase(logger *slog.Logger) *Database {
	return &Database{
		StakeWiseVaults: map[string][]*StakeWiseVault{},
		Users:           []*User{},
		logger:          logger,
	}
}

// Adds a StakeWise vault to the database
func (d *Database) AddStakeWiseVault(address common.Address, networkName string) error {
	networkVaults, exists := d.StakeWiseVaults[networkName]
	if !exists {
		networkVaults = []*StakeWiseVault{}
		d.StakeWiseVaults[networkName] = networkVaults
	}

	for _, vault := range networkVaults {
		if vault.Address == address {
			return fmt.Errorf("stakewise vault with address [%s] already exists", address.Hex())
		}
	}

	vault := NewStakeWiseVaultInfo(address)
	networkVaults = append(networkVaults, vault)
	d.StakeWiseVaults[networkName] = networkVaults
	return nil
}

// Adds a user to the database
func (d *Database) AddUser(email string) error {
	for _, user := range d.Users {
		if user.Email == email {
			return fmt.Errorf("user with email [%s] already exists", email)
		}
	}

	user := newUser(email)
	d.Users = append(d.Users, user)
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
	clone := NewDatabase(d.logger)

	// Copy StakeWise vaults
	for network, vaults := range d.StakeWiseVaults {
		networkVaults := make([]*StakeWiseVault, len(vaults))
		for i, vault := range vaults {
			networkVaults[i] = vault.Clone()
		}
		clone.StakeWiseVaults[network] = networkVaults
	}

	// Copy users
	for _, user := range d.Users {
		clone.Users = append(clone.Users, user.Clone())
	}
	return clone
}

// ===============
// === Getters ===
// ===============

// Get a node by address
func (d *Database) GetNode(address common.Address) *Node {
	for _, user := range d.Users {
		for _, candidate := range user.Nodes {
			if candidate.Address == address {
				return candidate
			}
		}
	}
	return nil
}

// Get the StakeWise status of a validator
func (d *Database) GetStakeWiseVault(address common.Address, networkName string) *StakeWiseVault {
	vaults, exists := d.StakeWiseVaults[networkName]
	if !exists {
		return nil
	}
	for _, vault := range vaults {
		if vault.Address == address {
			return vault
		}
	}
	return nil
}

// Handle a new collection of deposit data uploads from a node
func (d *Database) HandleDepositDataUpload(nodeAddress common.Address, data []beacon.ExtendedDepositData) error {
	// Get the node
	var node *Node
	for _, user := range d.Users {
		for _, candidate := range user.Nodes {
			if candidate.Address == nodeAddress {
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
		vaultAddress := common.BytesToAddress(depositData.WithdrawalCredentials)
		vaults, exists := d.StakeWiseVaults[depositData.NetworkName]
		if !exists {
			return fmt.Errorf("network [%s] not found in StakeWise vaults", depositData.NetworkName)
		}
		found := false
		for _, vault := range vaults {
			if vault.Address == vaultAddress {
				found = true
				node.AddDepositData(depositData, vaultAddress)
				break
			}
		}
		if !found {
			return fmt.Errorf("vault with address [%s] not found", vaultAddress.Hex())
		}
	}

	return nil
}

// Handle a new collection of signed exits from a node
func (d *Database) HandleSignedExitUpload(nodeAddress common.Address, network string, data []api.ExitData) error {
	// Get the node
	var node *Node
	for _, user := range d.Users {
		for _, candidate := range user.Nodes {
			if candidate.Address == nodeAddress {
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
		validators, exists := node.Validators[network]
		if !exists {
			return fmt.Errorf("network [%s] is not used by node [%s]", network, nodeAddress.Hex())
		}
		found := false
		for _, validator := range validators {
			if validator.Pubkey == pubkey {
				validator.SetExitMessage(signedExit.ExitMessage)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("node [%s] doesn't have validator [%s]", nodeAddress.Hex(), pubkey.Hex())
		}

	}
	return nil
}

// Create a new deposit data set
func (d *Database) CreateNewDepositDataSet(network string, validatorsPerUser int) []beacon.ExtendedDepositData {
	depositData := []beacon.ExtendedDepositData{}

	// Iterate the users
	for _, user := range d.Users {
		userCount := 0
		for _, node := range user.Nodes {
			validatorsForNetwork, exists := node.Validators[network]
			if !exists {
				continue
			}
			for _, validator := range validatorsForNetwork {
				// Add this deposit data if it hasn't been used
				if !validator.DepositDataUsed {
					depositData = append(depositData, validator.DepositData)
					userCount++
					if userCount >= validatorsPerUser {
						break
					}
				}
			}
			if userCount >= validatorsPerUser {
				break
			}
		}
	}

	return depositData
}

// Call this to "upload" a deposit data set to StakeWise
func (d *Database) UploadDepositDataToStakeWise(vaultAddress common.Address, network string, data []beacon.ExtendedDepositData) error {
	vaults, exists := d.StakeWiseVaults[network]
	if !exists {
		return fmt.Errorf("network [%s] not found in StakeWise vaults", network)
	}
	var vault *StakeWiseVault
	for _, candidate := range vaults {
		if candidate.Address == vaultAddress {
			vault = candidate
			break
		}
	}
	if vault == nil {
		return fmt.Errorf("vault with address [%s] not found", vaultAddress.Hex())
	}

	for _, depositData := range data {
		pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
		vault.MarkDepositDataUploaded(pubkey)
	}
	return nil
}

// Call this once a deposit data set has been "uploaded" to StakeWise
func (d *Database) MarkDepositDataSetUploaded(vaultAddress common.Address, network string, data []beacon.ExtendedDepositData) error {
	vaults, exists := d.StakeWiseVaults[network]
	if !exists {
		return fmt.Errorf("network [%s] not found in StakeWise vaults", network)
	}

	var vault *StakeWiseVault
	for _, candidate := range vaults {
		if candidate.Address == vaultAddress {
			vault = candidate
			break
		}
	}
	if vault == nil {
		return fmt.Errorf("vault with address [%s] not found", vaultAddress.Hex())
	}

	// Flag each deposit data as uploaded
	for _, depositData := range data {
		network := depositData.NetworkName
		for _, user := range d.Users {
			for _, node := range user.Nodes {
				validators, exists := node.Validators[network]
				if !exists {
					continue
				}
				for _, validator := range validators {
					if validator.Pubkey == beacon.ValidatorPubkey(depositData.PublicKey) {
						validator.DepositData = depositData
						validator.UseDepositData()
					}
				}
			}
		}
	}

	// Increment the index
	vault.LatestDepositDataSet = data
	vault.LatestDepositDataSetIndex++
	return nil
}
