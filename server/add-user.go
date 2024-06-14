package server

import (
	"fmt"
	"net/http"
)

func (s *NodeSetMockServer) addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing email query parameter"))
		return
	}

	// Create a new deposit data set
	err := s.manager.AddUser(email)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Added new user", "email", email)
	handleSuccess(w, s.logger, "")
}
