package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
)

func (s *NodeSetMockServer) getDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	node, args := s.processApiRequest(w, r, nil)
	if node == nil {
		return
	}

	// Input validation
	network := args.Get("network")
	vaultAddress := common.HexToAddress(args.Get("vault"))
	vault := s.manager.GetStakeWiseVault(vaultAddress, network)
	if vault == nil {
		handleInputError(s.logger, w, fmt.Errorf("vault with address [%s] on network [%s] not found", vaultAddress.Hex(), network))
		return
	}

	// Write the response
	response := api.DepositDataResponse{
		Version: vault.LatestDepositDataSetIndex,
		Data:    vault.LatestDepositDataSet,
	}
	handleSuccess(w, s.logger, response)
}
