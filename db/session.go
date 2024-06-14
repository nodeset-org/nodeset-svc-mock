package db

import (
	"crypto/md5"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/rocket-pool/node-manager-core/utils"
)

var (
	ErrUnregisteredNode error = errors.New("node hasn't been registered with the NodeSet server yet")
)

// An authorization session for access to the API
type Session struct {
	// The session nonce
	Nonce string

	// The session token
	Token string

	// The address of the node that requested this session
	NodeAddress common.Address

	// Whether or not the user for the session has logged in
	IsLoggedIn bool
}

// Creates a new session
func newSession() *Session {
	// Make a random UUID for the session token
	token := uuid.New()

	// Do a quick hash of it to act as a nonce
	nonce := md5.Sum(token[:])

	return &Session{
		Nonce:      utils.EncodeHexWithPrefix(nonce[:]),
		Token:      token.String(),
		IsLoggedIn: false,
	}
}

func (s *Session) login(nodeAddress common.Address) {
	s.NodeAddress = nodeAddress
	s.IsLoggedIn = true
}

func (s *Session) Clone() *Session {
	return &Session{
		Nonce:       s.Nonce,
		Token:       s.Token,
		NodeAddress: s.NodeAddress,
		IsLoggedIn:  s.IsLoggedIn,
	}
}
