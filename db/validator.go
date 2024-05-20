package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Validator struct {
	Index               int
	Pubkey              beacon.ValidatorPubkey
	VaultAddress        common.Address
	DepositData         beacon.ExtendedDepositData
	SignedExit          api.ExitMessage
	ExitMessageUploaded bool
	DepositDataUsed     bool
}

func newValidator(depositData beacon.ExtendedDepositData, index int, vaultAddress common.Address) *Validator {
	return &Validator{
		Index:        index,
		Pubkey:       beacon.ValidatorPubkey(depositData.PublicKey),
		VaultAddress: vaultAddress,
		DepositData:  depositData,
	}
}

func (v *Validator) UseDepositData() {
	v.DepositDataUsed = true
}

func (v *Validator) SetExitMessage(exitMessage api.ExitMessage) {
	// Normally this is where validation would occur
	v.SignedExit = exitMessage
	v.ExitMessageUploaded = true
}

func (v *Validator) Clone() *Validator {
	return &Validator{
		Index:               v.Index,
		Pubkey:              v.Pubkey,
		VaultAddress:        v.VaultAddress,
		DepositData:         v.DepositData,
		SignedExit:          v.SignedExit,
		ExitMessageUploaded: v.ExitMessageUploaded,
		DepositDataUsed:     v.DepositDataUsed,
	}
}
