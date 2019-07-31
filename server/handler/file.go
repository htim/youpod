package handler

import (
	"github.com/go-chi/chi"
	"github.com/htim/youpod"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

//HEAD /files/{username}/{fileID}
func (h *Handler) headCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

//GET /files/{username}/{fileID}
func (h *Handler) serveFile(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")
	fileID := chi.URLParam(r, "fileID")

	user, err := h.UserService.FindUserByUsername(username)

	if err == youpod.ErrUserNotFound {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.WithError(err).WithField("username", username).Error("failed to find user")
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	metadata, err := h.MediaService.GetFileMetadata(user, fileID)
	if err != nil {
		log.WithError(err).Error("failed to get file metadata")
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
	}

	file, err := h.MediaService.GetFileContent(user, fileID)
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, metadata.Name, time.Time{}, file)

}
