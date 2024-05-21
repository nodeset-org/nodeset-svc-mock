package server

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
)

func (s *NodeSetMockServer) getValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	node, args := s.processApiRequest(w, r, nil)
	if node == nil {
		return
	}

	// Get the network
	network := args.Get("network")
	if network != test.Network {
		handleInputError(s.logger, w, fmt.Errorf("unsupported network [%s]", network))
		return
	}

	// Get the registered validators
	validatorStatuses := []api.ValidatorStatus{}
	validatorsForNetwork := node.Validators[network]

	// Iterate the validators
	for _, validator := range validatorsForNetwork {
		pubkey := validator.Pubkey
		status := s.manager.Database.GetValidatorStatus(network, pubkey)
		validatorStatuses = append(validatorStatuses, api.ValidatorStatus{
			Pubkey:              pubkey,
			Status:              string(status),
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	response := api.ValidatorsResponse{
		Data: validatorStatuses,
	}
	handleSuccess(w, s.logger, response)
}
