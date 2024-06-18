package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
)

func (s *NodeSetMockServer) depositDataMeta(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	args := s.processApiRequest(w, r, nil)
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}
	node := s.getNodeForSession(w, session)
	if node == nil {
		return
	}

	// Input validation
	network := args.Get("network")
	vaultAddress := common.HexToAddress(args.Get("vault"))
	vault := s.manager.GetStakeWiseVault(vaultAddress, network)
	if vault == nil {
		handleInputError(w, s.logger, fmt.Errorf("vault with address [%s] on network [%s] not found", vaultAddress.Hex(), network))
		return
	}

	// Write the response
	data := api.DepositDataMetaData{
		Version: vault.LatestDepositDataSetIndex,
	}
	handleSuccess(w, s.logger, data)
}
