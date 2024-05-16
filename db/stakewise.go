package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Info for StakeWise vaults
type StakeWiseVault struct {
	Address      common.Address
	UploadedData map[beacon.ValidatorPubkey]bool
}

func NewStakeWiseVaultInfo(address common.Address) *StakeWiseVault {
	return &StakeWiseVault{
		Address:      address,
		UploadedData: map[beacon.ValidatorPubkey]bool{},
	}
}

func (v *StakeWiseVault) MarkDepositDataUploaded(pubkey beacon.ValidatorPubkey) {
	v.UploadedData[pubkey] = true
}

func (v *StakeWiseVault) Clone() *StakeWiseVault {
	clone := NewStakeWiseVaultInfo(v.Address)
	for pubkey, uploaded := range v.UploadedData {
		clone.UploadedData[pubkey] = uploaded
	}
	return clone
}
