package api

import (
	"github.com/rocket-pool/node-manager-core/beacon"
)

// All responses from the NodeSet API will have this format
// `message` may or may not be populated (but should always be populated if `ok` is false)
// `data` should be populated if `ok` is true, and will be omitted if `ok` is false
type NodeSetResponse[DataType any] struct {
	OK      bool     `json:"ok"`
	Message string   `json:"message,omitempty"`
	Data    DataType `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Response to a login request
type LoginData struct {
	Token string `json:"token"`
}

// Data used returned from nonce requests
type NonceData struct {
	Nonce string `json:"nonce"`
	Token string `json:"token"`
}

// Response to a deposit data meta request
type DepositDataMetaData struct {
	Version int `json:"version"`
}

// Response to a deposit data request
type DepositDataData struct {
	Version     int                          `json:"version"`
	DepositData []beacon.ExtendedDepositData `json:"depositData"`
}

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	Status              string                 `json:"status"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// Status types for StakeWise validators
type StakeWiseStatus string

const (
	// DepositData hasn't been uploaded to NodeSet yet
	StakeWiseStatus_Unknown StakeWiseStatus = "UNKNOWN"

	// DepositData uploaded to NodeSet, but hasn't been made part of a deposit data set yet
	StakeWiseStatus_Pending StakeWiseStatus = "PENDING"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, but hasn't been activated on Beacon yet
	StakeWiseStatus_Uploaded StakeWiseStatus = "UPLOADED"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, and the validator is active on Beacon
	StakeWiseStatus_Registered StakeWiseStatus = "REGISTERED"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, and the validator is exited on Beacon
	StakeWiseStatus_Removed StakeWiseStatus = "REMOVED"
)
