package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/services"
)

type DeleteTextureController struct {
	deleteService *services.TextureDeleteService
}

func NewDeleteTextureController() *DeleteTextureController {
	return &DeleteTextureController{
		deleteService: services.NewTextureDeleteService(),
	}
}

type deleteTextureResponse struct {
	ID          uint   `json:"id"`
	Hash        string `json:"hash"`
	Type        string `json:"type"`
	UID         uint   `json:"uid"`
	Model       string `json:"model,omitempty"`
	Name        string `json:"name"`
	FileName    string `json:"file_name"`
	PreviewFile string `json:"preview_file"`
}

func (dc *DeleteTextureController) DeleteTexture(c *gin.Context) {
	user, err := AuthenticateUser(c)
	if err != nil {
		// AuthenticateUser already handles response and aborts
		return
	}

	textureType := strings.ToLower(strings.TrimSpace(c.PostForm("type")))
	hash := strings.ToLower(strings.TrimSpace(c.PostForm("hash")))

	if textureType == "" {
		respondError(c, http.StatusBadRequest, CodeInvalidRequest, "type is required")
		return
	}
	if hash == "" {
		respondError(c, http.StatusBadRequest, CodeInvalidRequest, "hash is required")
		return
	}

	record, err := dc.deleteService.DeleteTexture(services.DeleteTextureInput{
		Type: textureType,
		UID:  user.UID,
		Hash: hash,
	})
	if err != nil {
		status := http.StatusInternalServerError
		code := CodeInternalError
		switch {
		case errors.Is(err, services.ErrInvalidTextureType):
			status, code = http.StatusBadRequest, CodeTextureTypeInvalid
		case errors.Is(err, services.ErrInvalidTextureHash):
			status, code = http.StatusBadRequest, CodeInvalidRequest
		case errors.Is(err, services.ErrTextureNotFound):
			status, code = http.StatusNotFound, CodeTextureNotFound
		}

		respondError(c, status, code, err.Error())
		return
	}

	respondOK(c, "texture deleted successfully", deleteTextureResponse{
		ID:          record.ID,
		Hash:        record.Hash,
		Type:        record.Type,
		UID:         record.UID,
		Model:       record.Model,
		Name:        record.Name,
		FileName:    record.FileName,
		PreviewFile: record.PreviewFile,
	})
}
