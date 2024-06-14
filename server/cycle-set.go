package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

func (s *NodeSetMockServer) cycleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	networkName := query.Get("network")
	if networkName == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing network query parameter"))
		return
	}
	vaultAddressString := query.Get("vault")
	if vaultAddressString == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing vault query parameter"))
		return
	}
	vaultAddress := common.HexToAddress(vaultAddressString)
	userLimit := query.Get("user-limit")
	if userLimit == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing user-limit query parameter"))
		return
	}
	validatorsPerUser, err := strconv.ParseInt(userLimit, 10, 32)
	if err != nil {
		handleInputError(w, s.logger, fmt.Errorf("error parsing user-limit: %w", err))
		return
	}

	// Create a new deposit data set
	set := s.manager.CreateNewDepositDataSet(networkName, int(validatorsPerUser))
	s.logger.Info("Created new deposit data set", "network", networkName, "user-limit", validatorsPerUser)

	err = s.manager.UploadDepositDataToStakeWise(vaultAddress, networkName, set)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Uploaded deposit data set", "vault", vaultAddress.Hex())

	err = s.manager.MarkDepositDataSetUploaded(vaultAddress, networkName, set)
	if err != nil {
		handleServerError(w, s.logger, err)
		return
	}

	vault := s.manager.GetStakeWiseVault(vaultAddress, networkName)
	if vault != nil {
		s.logger.Info("Marked deposit data set as uploaded", "version", vault.LatestDepositDataSetIndex)
	}
	handleSuccess(w, s.logger, "")
}
