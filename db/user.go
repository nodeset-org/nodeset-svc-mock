package db

import (
	"github.com/ethereum/go-ethereum/common"
)

type User struct {
	Email string
	Nodes []*Node
}

func newUser(email string) *User {
	return &User{
		Email: email,
		Nodes: []*Node{},
	}
}

func (u *User) AddNode(nodeAddress common.Address) {
	for _, node := range u.Nodes {
		if node.Address == nodeAddress {
			return
		}
	}
	node := newNode(nodeAddress)
	u.Nodes = append(u.Nodes, node)
}

func (u *User) Clone() *User {
	clone := newUser(u.Email)
	clone.Nodes = make([]*Node, len(u.Nodes))
	for i, node := range u.Nodes {
		clone.Nodes[i] = node.Clone()
	}
	return clone
}
