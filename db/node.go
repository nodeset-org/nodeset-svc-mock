package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Node struct {
	Index              int
	Address            common.Address
	Validators         map[string]map[beacon.ValidatorPubkey]*Validator
	nextValidatorIndex int
}

func newNode(address common.Address, index int) *Node {
	return &Node{
		Index:      index,
		Address:    address,
		Validators: map[string]map[beacon.ValidatorPubkey]*Validator{},
	}
}

func (n *Node) AddDepositData(depositData beacon.ExtendedDepositData, vaultAddress common.Address) {
	dataMap, exists := n.Validators[depositData.NetworkName]
	if !exists {
		dataMap = map[beacon.ValidatorPubkey]*Validator{}
		n.Validators[depositData.NetworkName] = dataMap
	}

	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	validator, exists := dataMap[pubkey]
	if !exists {
		validator = newValidator(depositData, n.nextValidatorIndex, vaultAddress)
		n.nextValidatorIndex++
	}

	dataMap[pubkey] = validator
}

func (n *Node) Clone() *Node {
	clone := newNode(n.Address, n.Index)
	clone.nextValidatorIndex = n.nextValidatorIndex
	for network, dataMap := range n.Validators {
		validatorMap := map[beacon.ValidatorPubkey]*Validator{}
		for pubkey, validator := range dataMap {
			validatorMap[pubkey] = validator.Clone()
		}
		clone.Validators[network] = validatorMap
	}
	return clone
}
