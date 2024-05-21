package server

import (
	"fmt"
	"net/http"
)

func (s *NodeSetMockServer) addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(s.logger, w)
		return
	}

	// Input validation
	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		handleInputError(s.logger, w, fmt.Errorf("missing email query parameter"))
		return
	}

	// Create a new deposit data set
	err := s.manager.Database.AddUser(email)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Added new user", "email", email)
	handleSuccess(w, s.logger, "")
}
