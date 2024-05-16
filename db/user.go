package db

import (
	"github.com/ethereum/go-ethereum/common"
)

type User struct {
	Email string
	Nodes map[common.Address]*Node
}

func NewUser(email string) *User {
	return &User{
		Email: email,
		Nodes: map[common.Address]*Node{},
	}
}

func (u *User) AddNode(nodeAddress common.Address) {
	if _, exists := u.Nodes[nodeAddress]; exists {
		node := NewNode(nodeAddress)
		u.Nodes[nodeAddress] = node
	}
}

func (u *User) Clone() *User {
	clone := NewUser(u.Email)
	for address, node := range u.Nodes {
		clone.Nodes[address] = node
	}
	return clone
}
