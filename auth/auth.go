package auth

import (
	"crypto/ecdsa"
	"fmt"
	"log/slog"
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

// Authorizer is used to verify incoming requests have a valid signature and add signatures to outgoing requests
type Authorizer struct {
	authMessageHash []byte
	logger          *slog.Logger
}

// Creates a new authorizer
func NewAuthorizer(logger *slog.Logger) *Authorizer {
	authMessageHash := accounts.TextHash([]byte(nodesetAuthMessage))
	return &Authorizer{
		authMessageHash: authMessageHash,
		logger:          logger,
	}
}

// Verifies an incoming request has a valid signature, and returns the address of the signer
func (a *Authorizer) VerifyRequest(r *http.Request) (common.Address, bool, error) {
	authHeaderVals, exists := r.Header[authHeader]
	if !exists || len(authHeaderVals) == 0 {
		return common.Address{}, false, nil
	}

	authBytes, err := utils.DecodeHex(authHeaderVals[0])
	if err != nil {
		return common.Address{}, false, fmt.Errorf("error decoding auth header: %w", err)
	}

	address, err := a.getAddressFromSignature(authBytes)
	return address, true, err
}

// =============
// === Utils ===
// =============

// Adds an authorization header to an HTTP request
func AddAuthorizationHeader(request *http.Request, privateKey *ecdsa.PrivateKey) error {
	// Sign the auth message
	messageHash := accounts.TextHash([]byte(nodesetAuthMessage))
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		return fmt.Errorf("error signing auth message: %v", err)
	}
	// fix the ECDSA 'v' (see https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7#:~:text=The%20version%20number,2%E2%80%9D%20was%20introduced)
	signature[crypto.RecoveryIDOffset] += 27
	request.Header.Set(authHeader, utils.EncodeHexWithPrefix(signature))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return nil
}

// ==========================
// === Internal Functions ===
// ==========================

// Gets the address of the private key used to sign a message from a signature
func (a *Authorizer) getAddressFromSignature(signature []byte) (common.Address, error) {
	// fix the ECDSA 'v' (see https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7#:~:text=The%20version%20number,2%E2%80%9D%20was%20introduced)
	if signature[crypto.RecoveryIDOffset] >= 4 {
		signature[crypto.RecoveryIDOffset] -= 27
	}
	pubkeyBytes, err := crypto.SigToPub(a.authMessageHash, signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("error recovering pubkey from signature: %w", err)
	}

	return crypto.PubkeyToAddress(*pubkeyBytes), nil
}
