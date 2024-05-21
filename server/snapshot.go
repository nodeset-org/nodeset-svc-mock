package server

import (
	"fmt"
	"net/http"
)

func (s *NodeSetMockServer) snapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(s.logger, w)
		return
	}

	snapshotName := r.URL.Query().Get("name")
	if snapshotName == "" {
		handleInputError(s.logger, w, fmt.Errorf("missing snapshot name"))
		return
	}
	s.manager.TakeSnapshot(snapshotName)
	handleSuccess(w, s.logger, "")
}
