package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
)

func (s *NodeSetMockServer) getDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	node, args := s.processRequest(w, r, nil)
	if node == nil {
		return
	}

	// Input validation
	network := args.Get("network")
	vaults, exists := s.manager.Database.StakeWiseVaults[network]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("unsupported network [%s]", network))
		return
	}
	vaultAddress := common.HexToAddress(args.Get("vault"))
	vault, exists := vaults[vaultAddress]
	if !exists {
		handleInputError(s.logger, w, fmt.Errorf("vault with address [%s] not found", vaultAddress.Hex()))
		return
	}

	// Write the response
	response := api.DepositDataResponse{
		Version: vault.LatestDepositDataSetIndex,
		Data:    vault.LatestDepositDataSet,
	}
	handleSuccess(w, s.logger, response)
}
