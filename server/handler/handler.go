package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/htim/youpod"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/bot"
	"github.com/htim/youpod/cache"
	"github.com/htim/youpod/media"
	"github.com/htim/youpod/rss"
	"github.com/pkg/errors"
)

const (
	InternalErrorMessage = "internal error"
)

type Handler struct {
	userService     youpod.UserService
	mediaService    *media.Service
	googleDriveAuth auth.OAuth2
	bot             *bot.Telegram
	rss             *rss.Service

	responseCache *cache.LoadingCache
}

func NewHandler(userService youpod.UserService, mediaService *media.Service, googleDriveAuth auth.OAuth2, bot *bot.Telegram, rss *rss.Service) (*Handler, error) {
	rspCache, err := cache.NewLoadingCache()
	if err != nil {
		return nil, errors.Wrap(err, "cannot init response cache")
	}

	handler := &Handler{
		userService:     userService,
		mediaService:    mediaService,
		googleDriveAuth: googleDriveAuth,
		bot:             bot,
		rss:             rss,

		responseCache: rspCache,
	}

	return handler, nil
}

func (h *Handler) Routes() chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/gdrive/callback", h.gdriveAuthCallback)

	r.Head("/feed/{username}", h.headCheck)
	r.Get("/feed/{username}", h.rssFeed)

	r.Get("/files/{username}/{fileID}", h.serveFile)
	r.Get("/files/{username}/{fileID}/thumbnail", h.serveFileThumbnail)

	return r
}
