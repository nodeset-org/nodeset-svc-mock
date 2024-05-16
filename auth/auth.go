package auth

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// The message to sign with the node wallet when uploading deposit data
	nodesetAuthMessage string = "nodesetdev"

	// Header used for the wallet signature during a deposit data upload
	authHeader string = "Authorization"
)

type Authorizer struct {
	authMessageHash []byte
}

func NewAuthorizer() (*Authorizer, error) {
	authMessageHash := accounts.TextHash([]byte(nodesetAuthMessage))
	return &Authorizer{
		authMessageHash: authMessageHash,
	}, nil
}

// Verifies an incoming request has a valid signature, and returns the address of the signer
func (a *Authorizer) VerifyRequest(r *http.Request) (common.Address, error) {
	authHeaderVals, exists := r.Header[authHeader]
	if !exists || len(authHeaderVals) == 0 {
		return common.Address{}, fmt.Errorf("no auth header found")
	}

	authBytes, err := utils.DecodeHex(authHeaderVals[0])
	if err != nil {
		return common.Address{}, fmt.Errorf("error decoding auth header: %w", err)
	}

	return a.getAddressFromSignature(authBytes)
}

// ==========================
// === Internal Functions ===
// ==========================

// Gets the address of the private key used to sign a message from a signature
func (a *Authorizer) getAddressFromSignature(signature []byte) (common.Address, error) {
	pubkeyBytes, err := crypto.SigToPub(a.authMessageHash, signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("error recovering pubkey from signature: %w", err)
	}

	return crypto.PubkeyToAddress(*pubkeyBytes), nil
}
