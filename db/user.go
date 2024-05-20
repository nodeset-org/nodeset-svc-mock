package db

import (
	"github.com/ethereum/go-ethereum/common"
)

type User struct {
	Index int
	Email string
	Nodes map[common.Address]*Node

	nextNodeIndex int
}

func newUser(email string, index int) *User {
	return &User{
		Index: index,
		Email: email,
		Nodes: map[common.Address]*Node{},
	}
}

func (u *User) AddNode(nodeAddress common.Address) {
	if _, exists := u.Nodes[nodeAddress]; !exists {
		node := newNode(nodeAddress, u.nextNodeIndex)
		u.Nodes[nodeAddress] = node
		u.nextNodeIndex++
	}
}

func (u *User) Clone() *User {
	clone := newUser(u.Email, u.Index)
	clone.nextNodeIndex = u.nextNodeIndex
	for address, node := range u.Nodes {
		clone.Nodes[address] = node.Clone()
	}
	return clone
}
