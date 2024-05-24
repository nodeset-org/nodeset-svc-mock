package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

func (s *NodeSetMockServer) addNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(s.logger, w)
		return
	}

	// Input validation
	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		handleInputError(s.logger, w, fmt.Errorf("missing email query parameter"))
		return
	}
	addressString := query.Get("address")
	if addressString == "" {
		handleInputError(s.logger, w, fmt.Errorf("missing address query parameter"))
		return
	}
	address := common.HexToAddress(addressString)

	// Create a new deposit data set
	err := s.manager.AddNodeAccount(email, address)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Added new node account", "email", email, "address", address.Hex())
	handleSuccess(w, s.logger, "")
}
