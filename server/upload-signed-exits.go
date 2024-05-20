package server

import (
	"net/http"

	"github.com/nodeset-org/nodeset-svc-mock/api"
)

func (s *NodeSetMockServer) uploadSignedExits(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var exitData []api.ExitData
	node, args := s.processRequest(w, r, &exitData)
	if node == nil {
		return
	}

	// Handle the upload
	network := args.Get("network")
	err := s.manager.Database.HandleSignedExitUpload(node.Address, network, exitData)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	handleSuccess(w, s.logger, "")
}
