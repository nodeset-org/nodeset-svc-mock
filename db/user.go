package db

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrAlreadyRegistered error = errors.New("node has already been registered with the NodeSet server")
	ErrNotWhitelisted    error = errors.New("node address hasn't been whitelisted on the provided NodeSet account")
)

type User struct {
	Email            string
	WhitelistedNodes []*Node
	RegisteredNodes  []*Node
}

func newUser(email string) *User {
	return &User{
		Email:            email,
		WhitelistedNodes: []*Node{},
		RegisteredNodes:  []*Node{},
	}
}

func (u *User) WhitelistNode(nodeAddress common.Address) {
	for _, node := range u.RegisteredNodes {
		if node.Address == nodeAddress {
			return
		}
	}
	for _, node := range u.WhitelistedNodes {
		if node.Address == nodeAddress {
			return
		}
	}
	node := newNode(nodeAddress)
	u.WhitelistedNodes = append(u.WhitelistedNodes, node)
}

func (u *User) RegisterNode(nodeAddress common.Address) error {
	for _, node := range u.RegisteredNodes {
		if node.Address == nodeAddress {
			return ErrAlreadyRegistered
		}
	}
	for i, node := range u.WhitelistedNodes {
		if node.Address == nodeAddress {
			u.RegisteredNodes = append(u.RegisteredNodes, node)
			// Remove it from the whitelist
			u.WhitelistedNodes = append(u.WhitelistedNodes[:i], u.WhitelistedNodes[i+1:]...)
			return nil
		}
	}

	return ErrNotWhitelisted
}

func (u *User) Clone() *User {
	clone := newUser(u.Email)
	clone.WhitelistedNodes = make([]*Node, len(u.WhitelistedNodes))
	clone.RegisteredNodes = make([]*Node, len(u.RegisteredNodes))
	for i, node := range u.WhitelistedNodes {
		clone.WhitelistedNodes[i] = node.Clone()
	}
	for i, node := range u.RegisteredNodes {
		clone.RegisteredNodes[i] = node.Clone()
	}
	return clone
}
