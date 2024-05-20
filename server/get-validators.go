package server

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/nodeset-org/nodeset-svc-mock/db"
	"github.com/nodeset-org/nodeset-svc-mock/internal/test"
)

func (s *NodeSetMockServer) getValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	node, args := s.processRequest(w, r, nil)
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

	// Iterate the validators, sorted by index
	validators := make([]*db.Validator, 0, len(validatorsForNetwork))
	for _, validator := range validatorsForNetwork {
		validators = append(validators, validator)
	}
	sort.SliceStable(validators, func(i int, j int) bool {
		return validators[i].Index < validators[j].Index
	})
	for _, validator := range validators {
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
