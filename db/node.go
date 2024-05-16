package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Node struct {
	Address    common.Address
	Validators map[string]map[beacon.ValidatorPubkey]*Validator
}

func NewNode(address common.Address) *Node {
	return &Node{
		Address:    address,
		Validators: map[string]map[beacon.ValidatorPubkey]*Validator{},
	}
}

func (n *Node) AddDepositData(depositData beacon.ExtendedDepositData) {
	dataMap, exists := n.Validators[depositData.NetworkName]
	if !exists {
		dataMap = map[beacon.ValidatorPubkey]*Validator{}
		n.Validators[depositData.NetworkName] = dataMap
	}

	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	validator, exists := dataMap[pubkey]
	if !exists {
		validator = NewValidator(depositData)
	}

	dataMap[pubkey] = validator
}

func (n *Node) Clone() *Node {
	clone := NewNode(n.Address)
	for network, dataMap := range n.Validators {
		validatorMap := map[beacon.ValidatorPubkey]*Validator{}
		for pubkey, validator := range dataMap {
			validatorMap[pubkey] = validator.Clone()
		}
		clone.Validators[network] = validatorMap
	}
	return clone
}
