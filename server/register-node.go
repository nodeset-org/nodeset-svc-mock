package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) registerNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Get the requesting node
	var request api.RegisterNodeRequest
	_ = s.processApiRequest(w, r, &request)

	// Get the node
	address := common.HexToAddress(request.NodeAddress)
	node, isRegistered := s.manager.GetNode(address)
	if node == nil {
		handleNodeNotInWhitelist(w, s.logger, address)
		return
	}
	if isRegistered {
		handleAlreadyRegisteredNode(w, s.logger, address)
		return
	}

	// Register the node
	sig, err := utils.DecodeHex(request.Signature)
	if err != nil {
		handleInputError(w, s.logger, fmt.Errorf("invalid signature"))
		return
	}
	err = s.manager.RegisterNodeAccount(request.Email, address, sig)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Registered new node account", "email", request.Email, "address", address.Hex())
	handleSuccess(w, s.logger, "")
}
