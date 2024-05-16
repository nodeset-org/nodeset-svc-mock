package types

type ExitMessageDetails struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

type ExitMessage struct {
	Message   ExitMessageDetails `json:"message"`
	Signature string             `json:"signature"`
}

type ExitData struct {
	Pubkey      string      `json:"pubkey"`
	ExitMessage ExitMessage `json:"exit_message"`
}
