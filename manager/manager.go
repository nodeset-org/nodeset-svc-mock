package manager

import (
	"errors"
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
	database *db.Database

	// Internal fields
	snapshots map[string]*db.Database
	logger    *slog.Logger
}

var (
	ErrInvalidSession error = errors.New("session token is invalid")
)

// Creates a new manager
func NewNodeSetMockManager(logger *slog.Logger) *NodeSetMockManager {
	return &NodeSetMockManager{
		database:  db.NewDatabase(logger),
		snapshots: map[string]*db.Database{},
		logger:    logger,
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

// Whitelists a node with a user
func (m *NodeSetMockManager) WhitelistNodeAccount(email string, nodeAddress common.Address) error {
	return m.database.WhitelistNodeAccount(email, nodeAddress)
}

// Registers a whitelisted node with a user
func (m *NodeSetMockManager) RegisterNodeAccount(email string, nodeAddress common.Address, signature []byte) error {
	// Verify the signature
	err := auth.VerifyRegistrationSignature(email, nodeAddress, signature)
	if err != nil {
		return err
	}

	// Try to register the node
	return m.database.RegisterNodeAccount(email, nodeAddress)
}

// Creates a new session and returns the nonce for it
func (m *NodeSetMockManager) CreateSession() *db.Session {
	return m.database.CreateSession()
}

// Logs a session in
func (m *NodeSetMockManager) Login(nonce string, nodeAddress common.Address, signature []byte) error {
	// Verify the signature
	err := auth.VerifyLoginSignature(nonce, nodeAddress, signature)
	if err != nil {
		return err
	}

	// Log the session in
	return m.database.Login(nodeAddress, nonce)
}

// Gets a session by nonce
func (m *NodeSetMockManager) GetSessionByNonce(nonce string) *db.Session {
	return m.database.GetSessionByNonce(nonce)
}

// Gets a session by token
func (m *NodeSetMockManager) GetSessionByToken(token string) *db.Session {
	return m.database.GetSessionByToken(token)
}

// Verifies a request's session and returns the node address the session belongs to
func (m *NodeSetMockManager) VerifyRequest(r *http.Request) (*db.Session, error) {
	token, err := auth.GetSessionTokenFromRequest(r)
	if err != nil {
		return nil, err
	}

	// Get the session
	session := m.database.GetSessionByToken(token)
	if session == nil {
		return nil, ErrInvalidSession
	}
	return session, nil
}

// Get a node by address - returns true if registered, false if just whitelisted
func (m *NodeSetMockManager) GetNode(address common.Address) (*db.Node, bool) {
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
		for _, node := range user.RegisteredNodes {
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
			if validator.MarkedActive {
				return api.StakeWiseStatus_Registered
			}
		}
	}

	// Check to see if the deposit data has been used
	if validator.DepositDataUsed {
		return api.StakeWiseStatus_Uploaded
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

// Call this once a deposit data set has been "registered" to StakeWise
func (m *NodeSetMockManager) MarkValidatorsRegistered(vaultAddress common.Address, network string, data []beacon.ExtendedDepositData) error {
	return m.database.MarkValidatorsRegistered(vaultAddress, network, data)
}
