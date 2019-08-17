package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/htim/youpod"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"strconv"
)

//Telegram encapsulates all bot-related logic
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

func (b *Telegram) Run() {
	go func() {
		for u := range b.updates {

			telegramID := int64(u.Message.From.ID)
			telegramUserName := u.Message.From.UserName
			chatID := u.Message.Chat.ID

			user, err := b.userService.FindUserByTelegramID(telegramID)

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

					if err := b.userService.SaveUser(user); err != nil {
						log.WithError(err).Error("failed to save new user")
						b.SendInternalError(chatID)
						continue
					}
					continue
				}

				log.WithError(err).Error("failed to get user")
				b.SendInternalError(chatID)
				continue
			}

			if user.GDriveToken.AccessToken == "" {
				b.RequestGDriveAuth(u.Message.Chat.ID, b.googleDriveAuth.URL(strconv.FormatInt(telegramID, 10)))
				continue
			}

			if u.Message.Text != "" {
				file, err := b.youtubeService.Download(user, u.Message.Text)
				if err != nil {
					log.WithError(err).WithField("user", user.Username).Error("failed to download youtube video")
					b.Send(chatID, "Failed to download video. Please try again later")
					continue
				}
				id, err := b.mediaService.SaveFile(user, file)
				if err != nil {
					log.WithError(err).WithField("user", user.Username).Error("failed to save media")
					b.Send(chatID, "Failed to save file in your Google Drive. Please try again later")
					continue
				}
				if err = b.userService.AddFileToUser(user, id); err != nil {
					log.WithError(err).WithField("user", user.Username).Error("failed to update user file list")
					b.SendInternalError(chatID)
					continue
				}

				b.Send(chatID, "Alright, the podcast based on this video will be available soon")
				b.Send(chatID, b.rssService.UserFeedUrl(user))

				b.youtubeService.Cleanup(file)
			}
		}
	}()
}

func (b *Telegram) SuccessfulAuth(telegramID int64, message string, onSend func()) error {
	user, err := b.userService.FindUserByTelegramID(telegramID)
	if err != nil {
		return errors.Wrap(err, "failed to find user")
	}
	b.Send(telegramID, message)
	onSend()

	b.Send(telegramID, fmt.Sprintf("Your feed url: %s. Add it to your favourite podcast app", b.rssService.UserFeedUrl(user)))
	return nil
}

func (b *Telegram) Send(chatId int64, text string) {
	message := tgbotapi.NewMessage(chatId, text)
	if _, err := b.api.Send(message); err != nil {
		log.WithError(err).Error("failed to send message to telegram")
	}
}

func (b *Telegram) SendAndDo(chatId int64, text string, do func()) {
	message := tgbotapi.NewMessage(chatId, text)
	if _, err := b.api.Send(message); err != nil {
		log.WithError(err).Error("failed to send message to telegram")
	}
	do()
}

func (b *Telegram) SendInternalError(chatId int64) {
	message := tgbotapi.NewMessage(chatId, "Internal error. Please try again later")
	if _, err := b.api.Send(message); err != nil {
		log.WithError(err).Error("failed to send message to telegram")
	}
}

func (b *Telegram) RequestGDriveAuth(chatID int64, url string) {
	btn := tgbotapi.NewInlineKeyboardButtonURL("login at Google Drive", url)

	msg := tgbotapi.NewMessage(chatID, "Please login at Google Drive. It is used to save your uploaded audios")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})

	if _, err := b.api.Send(msg); err != nil {
		log.WithError(err).Error("failed to send login at Google Drive button")
	}
}

func (b *Telegram) generateFeedUrl(user core.User) string {
	return fmt.Sprintf("%s/%s/%s", b.rootUrl, user.Username, xid.New().String())
}
