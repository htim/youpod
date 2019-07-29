package handler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func (h *Handler) gdriveAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	token, err := h.GoogleDriveAuth.Exchange(code)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	telegramID, err := strconv.Atoi(state)
	if err != nil {
		http.Error(w, "unparseable state", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.FindUserByTelegramID(int64(telegramID))
	if err != nil {
		log.WithError(err).Error("cannot find user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	user.GDriveToken = token

	userInfo, err := h.GoogleDriveAuth.GetUserInfo(user.GDriveToken)
	if err != nil {
		log.WithError(err).Error("cannot retrieve google drive user info")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err = h.UserService.SaveUser(*user); err != nil {
		log.WithError(err).Error("cannot update user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err = h.Bot.SuccessfulAuth(user.TelegramID, userInfo.Email, func() {
		if _, err := fmt.Fprint(w, "Successfully logged in, back to Telegram"); err != nil {
			log.WithError(err).Error("cannot write response")
		}
	}); err != nil {
		log.WithError(err).Error("cannot handle successful auth")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

}
