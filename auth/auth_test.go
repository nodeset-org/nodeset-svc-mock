package auth

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/tyler-smith/go-bip39"
)

const (
	derivationPath string = "m/44'/60'/0'/0/%d"
	mnemonic       string = "test test test test test test test test test test test junk"
	goodVault      string = "0x1234567890123456789012345678901234567890"
	network        string = "holesky"
)

// =============
// === Tests ===
// =============

func TestRecoverPubkey(t *testing.T) {
	// Get a private key
	privateKey, err := getPrivateKey(derivationPath, 0, mnemonic)
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
	authorizer, err := NewAuthorizer()
	if err != nil {
		t.Fatalf("error creating authorizer: %v", err)
	}
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
	// Get a private key
	privateKey, err := getPrivateKey(derivationPath, 0, mnemonic)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Create a request with the proper header
	vault := utils.RemovePrefix(goodVault)
	params := map[string]string{
		"vault":   vault,
		"network": network,
	}
	request, err := generateRequest(privateKey, http.MethodGet, nil, params, "deposit-data", "meta")
	if err != nil {
		t.Fatalf("error generating request: %v", err)
	}
	t.Log("Generated deposit-data/meta request")

	// Verify the request
	authorizer, err := NewAuthorizer()
	if err != nil {
		t.Fatalf("error creating authorizer: %v", err)
	}
	recoveredPubkey, err := authorizer.VerifyRequest(request)
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

// Get the private key from the account recovery info
func getPrivateKey(derivationPath string, index uint, mnemonic string) (*ecdsa.PrivateKey, error) {
	// Check the mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic '%s'", mnemonic)
	}

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("error creating wallet master key: %w", err)
	}

	// Get the derived key
	derivedKey, _, err := getDerivedKey(masterKey, derivationPath, index)
	if err != nil {
		return nil, fmt.Errorf("error getting node wallet derived key: %w", err)
	}

	// Get the private key from it
	privateKey, err := derivedKey.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("error getting node wallet private key: %w", err)
	}
	privateKeyECDSA := privateKey.ToECDSA()
	return privateKeyECDSA, nil
}

// Get the derived key & derivation path for the account at the index
func getDerivedKey(masterKey *hdkeychain.ExtendedKey, derivationPath string, index uint) (*hdkeychain.ExtendedKey, uint, error) {
	formattedDerivationPath := fmt.Sprintf(derivationPath, index)

	// Parse derivation path
	path, err := accounts.ParseDerivationPath(formattedDerivationPath)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid node key derivation path '%s': %w", formattedDerivationPath, err)
	}

	// Follow derivation path
	key := masterKey
	for i, n := range path {
		key, err = key.Derive(n)
		if err == hdkeychain.ErrInvalidChild {
			// Start over with the next index
			return getDerivedKey(masterKey, derivationPath, index+1)
		} else if err != nil {
			return nil, 0, fmt.Errorf("invalid child key at depth %d: %w", i, err)
		}
	}

	// Return
	return key, index, nil
}

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

	// Sign the auth message
	messageHash := accounts.TextHash([]byte(nodesetAuthMessage))
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error signing auth message: %v", err)
	}
	request.Header.Set(authHeader, utils.EncodeHexWithPrefix(signature))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return request, nil
}
