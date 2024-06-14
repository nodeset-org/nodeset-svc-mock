package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

func (s *NodeSetMockServer) addStakeWiseVault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	network := query.Get("network")
	if network == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing network query parameter"))
		return
	}
	addressString := query.Get("address")
	if addressString == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing address query parameter"))
		return
	}
	address := common.HexToAddress(addressString)

	// Create a new deposit data set
	err := s.manager.AddStakeWiseVault(address, network)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Added new stakewise vault", "address", address.Hex(), "network", network)
	handleSuccess(w, s.logger, "")
}
