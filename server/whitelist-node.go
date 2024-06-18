package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

func (s *NodeSetMockServer) whitelistNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing email query parameter"))
		return
	}
	addressString := query.Get("address")
	if addressString == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing address query parameter"))
		return
	}
	address := common.HexToAddress(addressString)

	// Whitelist the node
	err := s.manager.WhitelistNodeAccount(email, address)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Whitelisted new node account", "email", email, "address", address.Hex())
	handleSuccess(w, s.logger, "")
}
