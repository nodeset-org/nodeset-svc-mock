package api

// Request to register a node with the NodeSet server
type RegisterNodeRequest struct {
	Email       string `json:"email"`
	NodeAddress string `json:"node_address"`
	Signature   string `json:"signature"` // Must be 0x-prefixed hex encoded
}

// Request to log into the NodeSet server
type LoginRequest struct {
	Nonce     string `json:"nonce"`
	Address   string `json:"address"`
	Signature string `json:"signature"` // Must be 0x-prefixed hex encoded
}

// Details of an exit message
type ExitMessageDetails struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

// Voluntary exit message
type ExitMessage struct {
	Message   ExitMessageDetails `json:"message"`
	Signature string             `json:"signature"`
}

// Data for a pubkey's voluntary exit message
type ExitData struct {
	Pubkey      string      `json:"pubkey"`
	ExitMessage ExitMessage `json:"exit_message"`
}
