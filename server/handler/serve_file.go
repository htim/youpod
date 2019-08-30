package handler

import (
	"bytes"
	context2 "context"
	"encoding/base64"
	"github.com/go-chi/chi"
	"github.com/htim/youpod"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"time"
)

type fileKey struct {
	username string
	fileID   string
}

type fileValue struct {
	rs   io.ReadSeeker
	name string
}

//HEAD /files/{username}/{fileID}.mp3
func (h *Handler) headCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

//GET /files/{username}/{fileID}.mp3
func (h *Handler) serveFile(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")
	fileID := chi.URLParam(r, "fileID")

	fk := fileKey{
		username: username,
		fileID:   fileID,
	}

	var f fileValue

	cached, ok := h.responseCache.Get(fk)

	if ok {

		f = cached.(fileValue)

	} else {

		user, err := h.userService.FindUserByUsername(context2.Background(), username)

		if err != nil {

			if err == youpod.ErrUserNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}

			log.WithError(err).WithField("username", username).Error("failed to find user")
			http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
			return
		}

		metadata, err := h.mediaService.GetFileMetadata(user, fileID, context.Background())
		if err != nil {
			log.WithError(err).Error("failed to get rs metadata")
			http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
			return
		}

		rs, err := h.mediaService.GetFileContent(user, fileID)

		if err != nil {
			log.WithError(err).Error("failed to get rs content")
			http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
			return
		}

		f = fileValue{
			rs:   rs,
			name: metadata.Name,
		}

		h.responseCache.Add(fk, f)
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, f.name, time.Time{}, f.rs)

}

//GET /files/{username}/{fileID}/thumbnail.jpg
func (h *Handler) serveFileThumbnail(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	fileID := chi.URLParam(r, "fileID")

	user, err := h.userService.FindUserByUsername(context2.Background(), username)

	if err != nil {

		if err == youpod.ErrUserNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		log.WithError(err).WithField("username", username).Error("failed to find user")
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	metadata, err := h.mediaService.GetFileMetadata(user, fileID, context.Background())
	if err != nil {
		log.WithError(err).Error("failed to get file metadata")
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	bb, err := base64.StdEncoding.DecodeString(metadata.Picture)
	if err != nil {
		log.WithError(err).Error("failed to decode image")
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")

	reader := bytes.NewReader(bb)

	if _, err = io.Copy(w, reader); err != nil {
		log.WithError(err).Error("cannot serve thumbnail")
	}

}
