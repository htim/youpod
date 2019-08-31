package bot

import (
	context2 "context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/htim/youpod"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"strconv"
)

type Telegram struct {
	api *tgbotapi.BotAPI

	userService    core.UserRepository
	youtubeService core.YoutubeService
	mediaService   core.MediaService
	rssService     core.RssService

	googleDriveAuth auth.OAuth2

	updates tgbotapi.UpdatesChannel

	rootUrl string
}

func NewTelegram(
	telegramToken string,

	userService core.UserRepository,
	youtubeService core.YoutubeService,
	mediaService core.MediaService,
	rssService core.RssService,

	googleDriveAuth auth.OAuth2,
	rootUrl string,
) (*Telegram, error) {

	api, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Telegram API")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := api.GetUpdatesChan(u)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get updates channel")
	}

	return &Telegram{
		api: api,

		userService:    userService,
		youtubeService: youtubeService,

		mediaService: mediaService,
		rssService:   rssService,

		updates: updates,

		googleDriveAuth: googleDriveAuth,

		rootUrl: rootUrl,
	}, nil
}

func (t *Telegram) Run() {
	go func() {
		for u := range t.updates {

			telegramID := int64(u.Message.From.ID)
			telegramUserName := u.Message.From.UserName
			chatID := u.Message.Chat.ID

			user, err := t.userService.FindUserByTelegramID(context2.Background(), telegramID)

			if err != nil {

				if err == youpod.ErrUserNotFound {
					username := telegramUserName
					if username == "" {
						username = xid.New().String() + "_telegram"
					}

					user = core.User{
						Username:   username,
						TelegramID: telegramID,
					}

					if err := t.userService.SaveUser(context.Background(), user); err != nil {
						log.WithError(err).Error("failed to save new user")
						t.SendInternalError(chatID)
						continue
					}
					continue
				}

				log.WithError(err).Error("failed to get user")
				t.SendInternalError(chatID)
				continue
			}

			if user.GDriveToken.AccessToken == "" {
				t.RequestGDriveAuth(u.Message.Chat.ID, t.googleDriveAuth.URL(strconv.FormatInt(telegramID, 10)))
				continue
			}

			if u.Message.Text != "" {
				file, err := t.youtubeService.Download(user, u.Message.Text)
				if err != nil {
					log.WithError(err).WithField("user", user.Username).Error("failed to download youtube video")
					t.Send(chatID, "Failed to download video. Please try again later")
					continue
				}
				id, err := t.mediaService.SaveFile(user, file)
				if err != nil {
					log.WithError(err).WithField("user", user.Username).Error("failed to save media")
					t.Send(chatID, "Failed to save file in your Google Drive. Please try again later")
					continue
				}
				if err = t.userService.AddFileToUser(context2.Background(), user, id); err != nil {
					log.WithError(err).WithField("user", user.Username).Error("failed to update user file list")
					t.SendInternalError(chatID)
					continue
				}

				t.Send(chatID, "Alright, the podcast based on this video will be available soon")
				t.Send(chatID, t.rssService.UserFeedUrl(user))

				t.youtubeService.Cleanup(file)
			}
		}
	}()
}

func (t *Telegram) SuccessfulAuth(telegramID int64, message string, onSend func()) error {
	user, err := t.userService.FindUserByTelegramID(context2.Background(), telegramID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	t.Send(telegramID, message)
	onSend()

	t.Send(telegramID, fmt.Sprintf("Your feed url: %s. Add it to your favourite podcast app", t.rssService.UserFeedUrl(user)))
	return nil
}

func (t *Telegram) Send(chatId int64, text string) {
	message := tgbotapi.NewMessage(chatId, text)
	if _, err := t.api.Send(message); err != nil {
		log.WithError(err).Error("failed to send message to telegram")
	}
}

func (t *Telegram) SendAndDo(chatId int64, text string, do func()) {
	message := tgbotapi.NewMessage(chatId, text)
	if _, err := t.api.Send(message); err != nil {
		log.WithError(err).Error("failed to send message to telegram")
	}
	do()
}

func (t *Telegram) SendInternalError(chatId int64) {
	message := tgbotapi.NewMessage(chatId, "Internal error. Please try again later")
	if _, err := t.api.Send(message); err != nil {
		log.WithError(err).Error("failed to send message to telegram")
	}
}

func (t *Telegram) RequestGDriveAuth(chatID int64, url string) {
	btn := tgbotapi.NewInlineKeyboardButtonURL("login at Google Drive", url)

	msg := tgbotapi.NewMessage(chatID, "Please login at Google Drive. It is used to save your uploaded audios")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})

	if _, err := t.api.Send(msg); err != nil {
		log.WithError(err).Error("failed to send login at Google Drive button")
	}
}
