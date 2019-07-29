package handler

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/htim/youpod/rss"
	log "github.com/sirupsen/logrus"
	"net/http"
)

//GET /feed/{username}
func (h *Handler) rssFeed(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")

	user, err := h.UserService.FindUserByUsername(username)
	if err != nil {
		log.WithError(err).Error("cannot find user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	feed, err := h.Rss.UserFeed(*user, rss.XML)
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
