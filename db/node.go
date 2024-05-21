package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Node struct {
	Address    common.Address
	Validators map[string][]*Validator
}

func newNode(address common.Address) *Node {
	return &Node{
		Address:    address,
		Validators: map[string][]*Validator{},
	}
}

func (n *Node) AddDepositData(depositData beacon.ExtendedDepositData, vaultAddress common.Address) {
	validatorsForNetwork, exists := n.Validators[depositData.NetworkName]
	if !exists {
		validatorsForNetwork = []*Validator{}
		n.Validators[depositData.NetworkName] = validatorsForNetwork
	}

	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	for _, validator := range validatorsForNetwork {
		if validator.Pubkey == pubkey {
			// Already present
			return
		}
	}

	validator := newValidator(depositData, vaultAddress)
	validatorsForNetwork = append(validatorsForNetwork, validator)
	n.Validators[depositData.NetworkName] = validatorsForNetwork
}

func (n *Node) Clone() *Node {
	clone := newNode(n.Address)
	for network, validatorsForNetwork := range n.Validators {
		cloneSlice := make([]*Validator, len(validatorsForNetwork))
		for i, validator := range validatorsForNetwork {
			cloneSlice[i] = validator.Clone()
		}
		clone.Validators[network] = cloneSlice
	}
	return clone
}
