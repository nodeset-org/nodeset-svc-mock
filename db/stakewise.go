package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Info for StakeWise vaults
type StakeWiseVault struct {
	// The vault address
	Address common.Address

	// The map of pubkeys that have been uploaded to StakeWise
	UploadedData map[beacon.ValidatorPubkey]bool

	// Index of the latest deposit data set uploaded to StakeWise
	LatestDepositDataSetIndex int

	// Latest deposit data set uploaded to StakeWise
	LatestDepositDataSet []beacon.ExtendedDepositData
}

func NewStakeWiseVaultInfo(address common.Address) *StakeWiseVault {
	return &StakeWiseVault{
		Address:                   address,
		UploadedData:              map[beacon.ValidatorPubkey]bool{},
		LatestDepositDataSet:      []beacon.ExtendedDepositData{},
		LatestDepositDataSetIndex: 0,
	}
}

func (v *StakeWiseVault) MarkDepositDataUploaded(pubkey beacon.ValidatorPubkey) {
	v.UploadedData[pubkey] = true
}

func (v *StakeWiseVault) Clone() *StakeWiseVault {
	clone := NewStakeWiseVaultInfo(v.Address)
	clone.LatestDepositDataSetIndex = v.LatestDepositDataSetIndex
	clone.LatestDepositDataSet = make([]beacon.ExtendedDepositData, len(v.LatestDepositDataSet))
	copy(clone.LatestDepositDataSet, v.LatestDepositDataSet)
	for pubkey, uploaded := range v.UploadedData {
		clone.UploadedData[pubkey] = uploaded
	}
	return clone
}
