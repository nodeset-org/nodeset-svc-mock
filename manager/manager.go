package manager

import (
	"fmt"
	"log/slog"

	"github.com/nodeset-org/nodeset-svc-mock/auth"
	"github.com/nodeset-org/nodeset-svc-mock/db"
)

type NodeSetMockManager struct {
	Database   *db.Database
	Authorizer *auth.Authorizer

	// Internal fields
	snapshots map[string]*db.Database
	logger    *slog.Logger
}

func NewNodeSetMockManager(logger *slog.Logger) *NodeSetMockManager {
	return &NodeSetMockManager{
		Database:   db.NewDatabase(logger),
		Authorizer: auth.NewAuthorizer(logger),
		snapshots:  map[string]*db.Database{},
		logger:     logger,
	}
}

func (m *NodeSetMockManager) TakeSnapshot(name string) {
	m.snapshots[name] = m.Database.Clone()
}

func (m *NodeSetMockManager) RevertToSnapshot(name string) error {
	snapshot, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot with name [%s] does not exist", name)
	}
	m.Database = snapshot
	return nil
}
