package test

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/tyler-smith/go-bip39"
	types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	EthDerivationPath           string = "m/44'/60'/0'/0/%d"
	BeaconDerivationPath        string = "m/12381/3600/%d/0/0"
	Mnemonic                    string = "test test test test test test test test test test test junk"
	StakeWiseVaultAddressHex    string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	Network                     string = "holesky"
	GenesisForkVersionString    string = "0x01017000"
	CapellaForkVersionString    string = "0x04017000"
	GenesisValidatorsRootString string = "0x9143aa7c615a7f7115e2b6aac319c03529df8242ae705fba9df39b79c59fa8b1"
	User0Email                  string = "user_0@test.com"
	User1Email                  string = "user_1@test.com"
	User2Email                  string = "user_2@test.com"
	User3Email                  string = "user_3@test.com"
	DepositAmount               uint64 = 32e9
	ExitEpoch                   uint64 = 100
)

var (
	StakeWiseVaultAddress common.Address = common.HexToAddress(StakeWiseVaultAddressHex)
	GenesisForkVersion    []byte         = common.FromHex(GenesisForkVersionString)
	CapellaForkVersion    []byte         = common.FromHex(CapellaForkVersionString)
	GenesisValidatorsRoot []byte         = common.FromHex(GenesisValidatorsRootString)
)

// Get the EL private key for the given index
func GetEthPrivateKey(index uint) (*ecdsa.PrivateKey, error) {
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
	derivedKey, _, err := getDerivedKey(masterKey, EthDerivationPath, index)
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

// Get the BLS private key for the given index
func GetBeaconPrivateKey(index uint) (*types.BLSPrivateKey, error) {
	path := fmt.Sprintf(BeaconDerivationPath, index)
	return validator.GetPrivateKey(Mnemonic, path)
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
