package db

import (
	"github.com/nodeset-org/nodeset-svc-mock/types"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Validator struct {
	Pubkey              beacon.ValidatorPubkey
	DepositData         beacon.ExtendedDepositData
	SignedExit          types.ExitMessage
	ExitMessageUploaded bool
	DepositDataUsed     bool
}

func NewValidator(depositData beacon.ExtendedDepositData) *Validator {
	return &Validator{
		Pubkey:      beacon.ValidatorPubkey(depositData.PublicKey),
		DepositData: depositData,
	}
}

func (v *Validator) UseDepositData() {
	v.DepositDataUsed = true
}

func (v *Validator) SetExitMessage(exitMessage types.ExitMessage) {
	// Normally this is where validation would occur
	v.SignedExit = exitMessage
	v.ExitMessageUploaded = true
}

func (v *Validator) Clone() *Validator {
	return &Validator{
		Pubkey:              v.Pubkey,
		DepositData:         v.DepositData,
		SignedExit:          v.SignedExit,
		ExitMessageUploaded: v.ExitMessageUploaded,
		DepositDataUsed:     v.DepositDataUsed,
	}
}
