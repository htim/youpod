package handler

import (
	"bytes"
	"encoding/base64"
	"github.com/go-chi/chi"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	if err != nil {
		log.WithError(err).WithField("username", username).Error("cannot find user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	file, err := h.FileService.GetFile(fileID, *user)
	if err != nil {
		log.WithError(err).WithField("user", user.Username).WithField("fileID", fileID).Error("cannot get file")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if file == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	defer func() {
		if err := file.Content.Close(); err != nil {
			log.WithError(err).Error("cannot close file content")
		}
	}()

	tempDir := os.TempDir()

	tmpFilename := tempDir + "/" + xid.New().String()

	tmp, err := os.Create(tmpFilename)
	if err != nil {
		log.WithError(err).Error("cannot create tmp file")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err = io.Copy(tmp, file.Content); err != nil {
		log.WithError(err).Error("cannot copy content to tmp file")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := tmp.Close(); err != nil {
			log.WithError(err).Error("cannot close tmp file")
		}
	}()

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, file.Name, time.Time{}, tmp)

}

//GET /files/{username}/{fileID}/thumbnail
func (h *Handler) serveFileThumbnail(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	fileID := chi.URLParam(r, "fileID")

	user, err := h.UserService.FindUserByUsername(username)
	if err != nil {
		log.WithError(err).WithField("username", username).Error("cannot find user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	file, err := h.FileService.GetFile(fileID, *user)
	if err != nil {
		log.WithError(err).WithField("user", user.Username).WithField("fileID", fileID).Error("cannot get file")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if file == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	bytesReader := bytes.NewReader(file.Picture)
	decoder := base64.NewDecoder(base64.StdEncoding, bytesReader)

	picture, err := ioutil.ReadAll(decoder)
	if err != nil {
		log.WithField("user", user.Username).WithField("file", fileID).WithError(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, file.Name, time.Time{}, bytes.NewReader(picture))

}
