package test_utils

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/tyler-smith/go-bip39"
)

const (
	DerivationPath           string = "m/44'/60'/0'/0/%d"
	Mnemonic                 string = "test test test test test test test test test test test junk"
	StakeWiseVaultAddressHex string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	Network                  string = "holesky"

	UserEmail       string = "test@test.com"
	NodeAddress0Hex string = "0x90de00eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	NodeAddress1Hex string = "0x90de01eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	PubkeyHex       string = "0xbeac09bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

var (
	StakeWiseVaultAddress common.Address         = common.HexToAddress(StakeWiseVaultAddressHex)
	Pubkey                beacon.ValidatorPubkey = parsePubkey(PubkeyHex)
)

// Get the private key from the account recovery info
func GetPrivateKey(index uint) (*ecdsa.PrivateKey, error) {
	// Check the mnemonic
	if !bip39.IsMnemonicValid(Mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic '%s'", Mnemonic)
	}

	// Generate the seed
	seed := bip39.NewSeed(Mnemonic, "")
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("error creating wallet master key: %w", err)
	}

	// Get the derived key
	derivedKey, _, err := getDerivedKey(masterKey, DerivationPath, index)
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

// ==========================
// === Internal Functions ===
// ==========================

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

func parsePubkey(pubkeyHex string) beacon.ValidatorPubkey {
	pubkey, err := beacon.HexToValidatorPubkey(pubkeyHex)
	if err != nil {
		panic(fmt.Sprintf("error parsing validator pubkey [%s]: %s", pubkeyHex, err.Error()))
	}
	return pubkey
}
