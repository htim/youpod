package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/htim/youpod"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/rss"
	"github.com/htim/youpod/telegram"
)

type Handler struct {
	UserService     youpod.UserService
	GoogleDriveAuth auth.OAuth2
	Bot             *telegram.YouPod
	Rss             *rss.Service
}

func (h *Handler) Routes() chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/gdrive/callback", h.gdriveAuthCallback)
	r.Get("/feed/{username}", h.rssFeed)

	return r
}
