package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/htim/youpod"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/media"
	"github.com/htim/youpod/rss"
	"github.com/htim/youpod/telegram"
)

const (
	InternalErrorMessage = "internal error"
)

type Handler struct {
	UserService     youpod.UserService
	MediaService    *media.Service
	GoogleDriveAuth auth.OAuth2
	Bot             *telegram.YouPod
	Rss             *rss.Service
}

func (h *Handler) Routes() chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/gdrive/callback", h.gdriveAuthCallback)

	r.Head("/feed/{username}", h.headCheck)
	r.Get("/feed/{username}", h.rssFeed)

	r.Get("/files/{username}/{fileID}", h.serveFile)

	return r
}
