package server

import (
	"net/http"

	"github.com/rocket-pool/node-manager-core/beacon"
)

func (s *NodeSetMockServer) uploadDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var depositData []beacon.ExtendedDepositData
	node, _ := s.processApiRequest(w, r, &depositData)
	if node == nil {
		return
	}

	// Handle the upload
	err := s.manager.HandleDepositDataUpload(node.Address, depositData)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	handleSuccess(w, s.logger, "")
}
