package manager

import (
	"fmt"

	"github.com/nodeset-org/nodeset-svc-mock/db"
)

type NodeSetMockManager struct {
	Database *db.Database

	// Internal fields
	snapshots map[string]*db.Database
}

func NewNodesetMockManager() *NodeSetMockManager {
	return &NodeSetMockManager{
		Database:  db.NewDatabase(),
		snapshots: map[string]*db.Database{},
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
