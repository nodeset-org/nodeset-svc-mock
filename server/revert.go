package server

import (
	"fmt"
	"net/http"
)

func (s *NodeSetMockServer) revert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(s.logger, w)
		return
	}

	snapshotName := r.URL.Query().Get("name")
	if snapshotName == "" {
		handleInputError(s.logger, w, fmt.Errorf("missing snapshot name"))
		return
	}

	err := s.manager.RevertToSnapshot(snapshotName)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	handleSuccess(w, s.logger, "")
}
