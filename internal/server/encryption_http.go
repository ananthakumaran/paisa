package server

import (
	"errors"
	"net/http"

	"github.com/ananthakumaran/paisa/internal/encryption"
	"github.com/gin-gonic/gin"
)

func ledgerReadErrHTTP(err error) (code int, body gin.H) {
	if err == nil {
		return 0, nil
	}
	if errors.Is(err, encryption.ErrNoPassword) {
		return http.StatusPreconditionRequired, gin.H{
			"error": err.Error(),
			"code":  "encryption_password_required",
		}
	}
	if errors.Is(err, encryption.ErrDecryptFailed) {
		return http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "encryption_decrypt_failed",
		}
	}
	return http.StatusInternalServerError, gin.H{"error": err.Error()}
}
