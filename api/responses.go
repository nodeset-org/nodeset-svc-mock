package api

import (
	"github.com/rocket-pool/node-manager-core/beacon"
)

// api/deposit-data/meta
type DepositDataMetaResponse struct {
	Version int `json:"version"`
}

// api/deposit-data
type DepositDataResponse struct {
	Version int                          `json:"version"`
	Data    []beacon.ExtendedDepositData `json:"data"`
}

// api/validators
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	Status              string                 `json:"status"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// api/dev/validators
type ValidatorsResponse struct {
	Data []ValidatorStatus `json:"data"`
}

// Generic response for errors
type ErrorResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// Status types for StakeWise validators
type StakeWiseStatus string

const (
	// DepositData uploaded to NodeSet, but hasn't been made part of a deposit data set yet
	StakeWiseStatus_Pending StakeWiseStatus = "PENDING"

	// DepositData uploaded to NodeSet, it's been added to a new deposit data set, but that set hasn't been uploaded to StakeWise yet
	StakeWiseStatus_Uploading StakeWiseStatus = "UPLOADING"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, but hasn't been activated on Beacon yet
	StakeWiseStatus_Uploaded StakeWiseStatus = "UPLOADED"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, and the validator is active on Beacon
	StakeWiseStatus_Registered StakeWiseStatus = "REGISTERED"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, and the validator is exited on Beacon
	StakeWiseStatus_Removed StakeWiseStatus = "REMOVED"
)
