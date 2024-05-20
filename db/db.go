package db

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Mock database for storing nodeset.io info
type Database struct {
	// Collection of StakeWise vaults
	StakeWiseVaults map[string]map[common.Address]*StakeWiseVault

	// Collection of users
	Users map[string]*User

	// Internal fields
	logger        *slog.Logger
	nextUserIndex int
}

// Creates a new database
func NewDatabase(logger *slog.Logger) *Database {
	return &Database{
		StakeWiseVaults: map[string]map[common.Address]*StakeWiseVault{},
		Users:           map[string]*User{},
		logger:          logger,
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

	user := newUser(email, d.nextUserIndex)
	d.Users[email] = user
	d.nextUserIndex++
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
	clone.nextUserIndex = d.nextUserIndex

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

// ===============
// === Getters ===
// ===============

func (d *Database) GetNode(address common.Address) *Node {
	for _, user := range d.Users {
		for candidateAddress, candidate := range user.Nodes {
			if candidateAddress == address {
				return candidate
			}
		}
	}
	return nil
}

// Get the StakeWise status of a validator
func (d *Database) GetValidatorStatus(network string, pubkey beacon.ValidatorPubkey) api.StakeWiseStatus {
	vaults, exists := d.StakeWiseVaults[network]
	if !exists {
		return api.StakeWiseStatus_Pending
	}

	// Get the validator for this pubkey
	var validator *Validator
	for _, user := range d.Users {
		for _, node := range user.Nodes {
			validators, exists := node.Validators[network]
			if !exists {
				continue
			}
			candidate, exists := validators[pubkey]
			if exists {
				validator = candidate
				break
			}
		}
		if validator != nil {
			break
		}
	}
	if validator == nil {
		return api.StakeWiseStatus_Pending
	}

	// Check if the StakeWise vault has already seen it
	uploadedToStakewise := vaults[validator.VaultAddress].UploadedData[validator.Pubkey]
	if uploadedToStakewise {
		return api.StakeWiseStatus_Uploaded
	}

	// Check to see if the deposit data has been used
	if validator.DepositDataUsed {
		return api.StakeWiseStatus_Uploading
	}
	return api.StakeWiseStatus_Pending
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
		vaultAddress := common.BytesToAddress(depositData.WithdrawalCredentials)
		vaults, exists := d.StakeWiseVaults[depositData.NetworkName]
		if !exists {
			return fmt.Errorf("network [%s] not found in StakeWise vaults", depositData.NetworkName)
		}
		_, exists = vaults[vaultAddress]
		if !exists {
			return fmt.Errorf("vault with address [%s] not found", vaultAddress.Hex())
		}

		node.AddDepositData(depositData, vaultAddress)
	}

	return nil
}

// Handle a new collection of signed exits from a node
func (d *Database) HandleSignedExitUpload(nodeAddress common.Address, network string, data []api.ExitData) error {
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

// Create a new deposit data set
func (d *Database) CreateNewDepositDataSet(network string, validatorsPerUser int) []beacon.ExtendedDepositData {
	depositData := []beacon.ExtendedDepositData{}

	// Iterate the users, sorted by index
	users := make([]*User, 0, len(d.Users))
	for _, user := range d.Users {
		users = append(users, user)
	}
	sort.SliceStable(users, func(i int, j int) bool {
		return users[i].Index < users[j].Index
	})
	for _, user := range users {
		userCount := 0

		// Iterate the nodes, sorted by index
		nodes := make([]*Node, 0, len(user.Nodes))
		for _, node := range user.Nodes {
			nodes = append(nodes, node)
		}
		sort.SliceStable(nodes, func(i int, j int) bool {
			return nodes[i].Index < nodes[j].Index
		})
		for _, node := range nodes {
			validatorsForNetwork, exists := node.Validators[network]
			if !exists {
				continue
			}

			// Iterate the validators, sorted by index
			validators := make([]*Validator, 0, len(validatorsForNetwork))
			for _, validator := range validatorsForNetwork {
				validators = append(validators, validator)
			}
			sort.SliceStable(validators, func(i int, j int) bool {
				return validators[i].Index < validators[j].Index
			})
			for _, validator := range validators {
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
	vault, exists := vaults[vaultAddress]
	if !exists {
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
	vault, exists := vaults[vaultAddress]
	if !exists {
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
