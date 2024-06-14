package server

import (
	"net/http"

	"github.com/nodeset-org/nodeset-svc-mock/api"
)

func (s *NodeSetMockServer) getNonce(w http.ResponseWriter, r *http.Request) {
	// Create a new session
	session := s.manager.CreateSession()

	// Write the response
	data := api.NonceData{
		Nonce: session.Nonce,
		Token: session.Token,
	}
	handleSuccess(w, s.logger, data)
	s.logger.Info("Created session", "nonce", session.Nonce)
}
