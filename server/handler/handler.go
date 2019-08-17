package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/bot"
	"github.com/htim/youpod/cache"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"net/http"
)

const (
	InternalErrorMessage = "internal error"
)

type Handler struct {
	userService  core.UserRepository
	rssService   core.RssService
	mediaService core.MediaService

	googleDriveAuth auth.OAuth2
	bot             *bot.Telegram

	responseCache *cache.LoadingCache
}

func NewHandler(
	userService core.UserRepository,
	mediaService core.MediaService,
	rss core.RssService,

	googleDriveAuth auth.OAuth2,
	bot *bot.Telegram,
) (*Handler, error) {
	rspCache, err := cache.NewLoadingCache()
	if err != nil {
		return nil, errors.Wrap(err, "cannot init response cache")
	}

	handler := &Handler{
		userService:     userService,
		mediaService:    mediaService,
		googleDriveAuth: googleDriveAuth,
		bot:             bot,
		rssService:      rss,

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

	r.Get("/files/{username}/{fileID}.mp3", h.serveFile)
	r.Get("/files/{username}/{fileID}/thumbnail.jpg", h.serveFileThumbnail)

	r.Mount("/", http.FileServer(http.Dir("./assets")))

	return r
}
