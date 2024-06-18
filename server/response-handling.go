package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-svc-mock/api"
	"github.com/rocket-pool/node-manager-core/log"
)

const (
	// Value of the auth response header if the node hasn't registered yet
	unregisteredAddressKey string = "unregistered_address"

	// Value of the auth response header if the login token has expired
	invalidSessionKey string = "invalid_session"

	// The node address has already been confirmed on a NodeSet account
	addressAlreadyAuthorizedKey string = "address_already_authorized"

	// The node address hasn't been whitelisted on the provided NodeSet account
	addressMissingWhitelistKey string = "address_missing_whitelist"
)

// Handle routes called with an invalid method
func handleInvalidMethod(w http.ResponseWriter, logger *slog.Logger) {
	writeResponse(w, logger, http.StatusMethodNotAllowed, []byte{})
}

// Handles an error related to parsing the input parameters of a request
func handleInputError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the auth header couldn't be decoded
func handleAuthHeaderError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the auth header is missing
func handleMissingAuthHeader(w http.ResponseWriter, logger *slog.Logger) {
	msg := "No Authorization header found"
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the session provided in the auth header is not valid
func handleInvalidSessionError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, invalidSessionKey)
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the node providing the request isn't registered
func handleUnregisteredNode(w http.ResponseWriter, logger *slog.Logger, address common.Address) {
	msg := fmt.Sprintf("No user found with authorized address %s", address.Hex())
	bytes := formatError(msg, unregisteredAddressKey)
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the node providing the request is already registered
func handleNodeNotInWhitelist(w http.ResponseWriter, logger *slog.Logger, address common.Address) {
	msg := fmt.Sprintf("Address %s is not whitelisted", address.Hex())
	bytes := formatError(msg, addressMissingWhitelistKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the node providing the request is already registered
func handleAlreadyRegisteredNode(w http.ResponseWriter, logger *slog.Logger, address common.Address) {
	msg := fmt.Sprintf("Address %s already registered", address.Hex())
	bytes := formatError(msg, addressAlreadyAuthorizedKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the auth header couldn't be decoded
func handleServerError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusInternalServerError, bytes)
}

// The request completed successfully
func handleSuccess[DataType any](w http.ResponseWriter, logger *slog.Logger, data DataType) {
	response := api.NodeSetResponse[DataType]{
		OK:      true,
		Message: "Success",
		Error:   "",
		Data:    data,
	}

	// Serialize the response
	bytes, err := json.Marshal(response)
	if err != nil {
		handleServerError(w, logger, fmt.Errorf("error serializing response: %w", err))
	}
	// Write it
	logger.Debug("Response body", slog.String(log.BodyKey, string(bytes)))
	writeResponse(w, logger, http.StatusOK, bytes)
}

// Writes a response to an HTTP request back to the client and logs it
func writeResponse(w http.ResponseWriter, logger *slog.Logger, statusCode int, message []byte) {
	// Prep the log attributes
	codeMsg := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	attrs := []any{
		slog.String(log.CodeKey, codeMsg),
	}

	// Log the response
	logMsg := "Responded with:"
	switch statusCode {
	case http.StatusOK:
		logger.Info(logMsg, attrs...)
	case http.StatusInternalServerError:
		logger.Error(logMsg, attrs...)
	default:
		logger.Warn(logMsg, attrs...)
	}

	// Write it to the client
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(message)
	if writeErr != nil {
		logger.Error("Error writing response", "error", writeErr)
	}
}

// JSONifies an error for responding to requests
func formatError(message string, errorKey string) []byte {
	msg := api.NodeSetResponse[struct{}]{
		OK:      false,
		Message: message,
		Error:   errorKey,
		Data:    struct{}{},
	}

	bytes, _ := json.Marshal(msg)
	return bytes
}
