package server

import (
	"net/http"

	"github.com/nodeset-org/nodeset-svc-mock/api"
)

func (s *NodeSetMockServer) getDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	node, _ := s.processGet(w, r)
	if node == nil {
		return
	}

	// Write the response
	response := api.DepositDataResponse{
		Version: s.manager.Database.LatestDepositDataSetIndex,
		Data:    s.manager.Database.LatestDepositDataSet,
	}
	handleSuccess(w, s.logger, response)
}
