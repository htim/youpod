package handler

import (
	context2 "context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/htim/youpod"
	log "github.com/sirupsen/logrus"
	"net/http"
)

//GET /feed/{username}
func (h *Handler) rssFeed(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")

	user, err := h.userService.FindUserByUsername(context2.Background(), username)
	if err != nil {

		if err == youpod.ErrUserNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		log.WithError(err).Error("cannot find user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	feed, err := h.rssService.UserFeed(user)
	if err != nil {
		log.WithError(err).WithField("user", user.Username).Error("cannot generate feed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err = fmt.Fprint(w, feed); err != nil {
		log.WithError(err).WithField("user", user.Username).Error("cannot send feed")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

}
