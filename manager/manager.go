package manager

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/db"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Mock manager for the nodeset.io service
type NodeSetMockManager struct {
	database   *db.Database
	authorizer *auth.Authorizer

	// Internal fields
	snapshots map[string]*db.Database
	logger    *slog.Logger
}

// Creates a new manager
func NewNodeSetMockManager(logger *slog.Logger) *NodeSetMockManager {
	return &NodeSetMockManager{
		database:   db.NewDatabase(logger),
		authorizer: auth.NewAuthorizer(logger),
		snapshots:  map[string]*db.Database{},
		logger:     logger,
	}
}

// Set the database for the manager directly if you need to custom provision it
func (m *NodeSetMockManager) SetDatabase(db *db.Database) {
	m.database = db
}

// Take a snapshot of the current database state
func (m *NodeSetMockManager) TakeSnapshot(name string) {
	m.snapshots[name] = m.database.Clone()
	m.logger.Info("Took DB snapshot", "name", name)
}

// Revert to a snapshot of the database state
func (m *NodeSetMockManager) RevertToSnapshot(name string) error {
	snapshot, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot with name [%s] does not exist", name)
	}
	m.database = snapshot
	m.logger.Info("Reverted to DB snapshot", "name", name)
	return nil
}

// ==================
// === Authorizer ===
// ==================

// Verifies a request has a valid signature, and returns the address of the signer
func (m *NodeSetMockManager) VerifyRequest(r *http.Request) (common.Address, bool, error) {
	return m.authorizer.VerifyRequest(r)
}

// ================
// === Database ===
// ================

// Adds a StakeWise vault
func (m *NodeSetMockManager) AddStakeWiseVault(address common.Address, networkName string) error {
	return m.database.AddStakeWiseVault(address, networkName)
}

// Gets a StakeWise vault
func (m *NodeSetMockManager) GetStakeWiseVault(address common.Address, networkName string) *db.StakeWiseVault {
	return m.database.GetStakeWiseVault(address, networkName)
}

// Adds a user to the database
func (m *NodeSetMockManager) AddUser(email string) error {
	return m.database.AddUser(email)
}

// Registers a node with a user
func (m *NodeSetMockManager) AddNodeAccount(email string, nodeAddress common.Address) error {
	return m.database.AddNodeAccount(email, nodeAddress)
}

// Get a node by address
func (m *NodeSetMockManager) GetNode(address common.Address) *db.Node {
	return m.database.GetNode(address)
}

// Get the StakeWise status of a validator
func (m *NodeSetMockManager) GetValidatorStatus(network string, pubkey beacon.ValidatorPubkey) api.StakeWiseStatus {
	vaults, exists := m.database.StakeWiseVaults[network]
	if !exists {
		return api.StakeWiseStatus_Pending
	}

	// Get the validator for this pubkey
	var validator *db.Validator
	for _, user := range m.database.Users {
		for _, node := range user.Nodes {
			validators, exists := node.Validators[network]
			if !exists {
				continue
			}
			for _, candidate := range validators {
				if candidate.Pubkey == pubkey {
					validator = candidate
					break
				}
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
	for _, vault := range vaults {
		if vault.Address == validator.VaultAddress && vault.UploadedData[validator.Pubkey] {
			return api.StakeWiseStatus_Uploaded
		}
	}

	// Check to see if the deposit data has been used
	if validator.DepositDataUsed {
		return api.StakeWiseStatus_Uploading
	}
	return api.StakeWiseStatus_Pending
}

// Handle a new collection of deposit data uploads from a node
func (m *NodeSetMockManager) HandleDepositDataUpload(nodeAddress common.Address, data []beacon.ExtendedDepositData) error {
	return m.database.HandleDepositDataUpload(nodeAddress, data)
}

// Handle a new collection of signed exits from a node
func (m *NodeSetMockManager) HandleSignedExitUpload(nodeAddress common.Address, network string, data []api.ExitData) error {
	return m.database.HandleSignedExitUpload(nodeAddress, network, data)
}

// Create a new deposit data set
func (m *NodeSetMockManager) CreateNewDepositDataSet(network string, validatorsPerUser int) []beacon.ExtendedDepositData {
	return m.database.CreateNewDepositDataSet(network, validatorsPerUser)
}

// Call this to "upload" a deposit data set to StakeWise
func (m *NodeSetMockManager) UploadDepositDataToStakeWise(vaultAddress common.Address, network string, data []beacon.ExtendedDepositData) error {
	return m.database.UploadDepositDataToStakeWise(vaultAddress, network, data)
}

// Call this once a deposit data set has been "uploaded" to StakeWise
func (m *NodeSetMockManager) MarkDepositDataSetUploaded(vaultAddress common.Address, network string, data []beacon.ExtendedDepositData) error {
	return m.database.MarkDepositDataSetUploaded(vaultAddress, network, data)
}
