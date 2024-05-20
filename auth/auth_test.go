package auth

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-svc-mock/test_utils"
	"github.com/rocket-pool/node-manager-core/utils"
)

// =============
// === Tests ===
// =============

func TestRecoverPubkey(t *testing.T) {
	logger := slog.Default()

	// Get a private key
	privateKey, err := test_utils.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Sign the auth message
	messageHash := accounts.TextHash([]byte(nodesetAuthMessage))
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		t.Fatalf("error signing auth message: %v", err)
	}
	t.Logf("Signed auth message, signature = %x", signature)

	// Get the pubkey from the signature
	authorizer := NewAuthorizer(logger)
	recoveredPubkey, err := authorizer.getAddressFromSignature(signature)
	if err != nil {
		t.Fatalf("error getting pubkey from signature: %v", err)
	}
	t.Logf("Recovered pubkey = %s", recoveredPubkey.Hex())

	// Check the pubkey
	if pubkey != recoveredPubkey {
		t.Fatalf("pubkey mismatch: expected %s, got %s", pubkey.Hex(), recoveredPubkey.Hex())
	}
	t.Logf("Pubkey matches")
}

func TestGoodRequest(t *testing.T) {
	logger := slog.Default()

	// Get a private key
	privateKey, err := test_utils.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Create a request with the proper header
	vault := utils.RemovePrefix(test_utils.StakeWiseVaultAddressHex)
	params := map[string]string{
		"vault":   vault,
		"network": test_utils.Network,
	}
	request, err := generateRequest(privateKey, http.MethodGet, nil, params, "deposit-data", "meta")
	if err != nil {
		t.Fatalf("error generating request: %v", err)
	}
	t.Log("Generated deposit-data/meta request")

	// Verify the request
	authorizer := NewAuthorizer(logger)
	recoveredPubkey, _, err := authorizer.VerifyRequest(request)
	if err != nil {
		t.Fatalf("error verifying request: %v", err)
	}
	t.Logf("Recovered pubkey = %s", recoveredPubkey.Hex())

	// Check the pubkey
	if pubkey != recoveredPubkey {
		t.Fatalf("pubkey mismatch: expected %s, got %s", pubkey.Hex(), recoveredPubkey.Hex())
	}
	t.Logf("Pubkey matches")
}

// ==========================
// === Internal Functions ===
// ==========================

// Generate an HTTP request with the signed auth header
func generateRequest(privateKey *ecdsa.PrivateKey, method string, body io.Reader, queryParams map[string]string, subroutes ...string) (*http.Request, error) {
	// Make the request
	path, err := url.JoinPath("http://dummy", subroutes...)
	if err != nil {
		return nil, fmt.Errorf("error joining path [%v]: %w", subroutes, err)
	}
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("error generating request to [%s]: %w", path, err)
	}
	query := request.URL.Query()
	for name, value := range queryParams {
		query.Add(name, value)
	}
	request.URL.RawQuery = query.Encode()

	// Add the auth header
	AddAuthorizationHeader(request, privateKey)
	return request, nil
}
