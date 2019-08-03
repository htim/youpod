package main

import (
	"github.com/htim/youpod"
	"github.com/htim/youpod/bot"
	"github.com/htim/youpod/media"
	gdrive "github.com/htim/youpod/media/google_drive"
	"github.com/htim/youpod/rss"
	"github.com/htim/youpod/server"
	"github.com/htim/youpod/server/handler"
	"github.com/htim/youpod/store/bolt"
	"github.com/htim/youpod/youtube"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"os"
)

var opts struct {
	ClientID     string `long:"client_id" env:"CLIENT_ID" description:"Google Drive client_id" required:"true"`
	ClientSecret string `long:"client_secret" env:"CLIENT_SECRET" description:"Google Drive client secret" required:"true"`

	TelegramBotApiKey string `long:"tg_bot_api_key" env:"TG_BOT_API_KEY" description:"Telegram Bot API Key" required:"true"`

	BaseURL string `long:"base_url" env:"BASE_URL" description:"app base url" required:"true"`
}

func main() {

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		log.WithError(err).Fatal("cannot parse options")
	}

	boltClient := bolt.NewClient("youpod.db")

	if err := boltClient.Open(); err != nil {
		log.WithError(err).Fatal("cannot open bolt client")
	}
	defer func() {
		if err := boltClient.Close(); err != nil {
			log.WithError(err).Error("cannot close bolt client")
		}
	}()

	userService, err := bolt.NewUserService(boltClient)
	if err != nil {
		log.WithError(err).Fatal("cannot init UserService")
	}

	googleDriveClient := gdrive.NewClient(
		userService,
		opts.ClientID,
		opts.ClientSecret,
		"http://localhost:9000"+"/gdrive/callback",
	)

	metadataService := bolt.NewMetadataService(boltClient, userService, googleDriveClient, "YouPod")

	youtubeService, err := youtube.NewService()
	if err != nil {
		log.WithError(err).Fatal("cannot init youtube service")
	}

	rssService := rss.NewService(opts.BaseURL, metadataService)

	mediaService := media.NewService(
		metadataService,
		map[youpod.StoreType]media.Store{
			youpod.GoogleDrive: googleDriveClient,
		},
	)

	tgBot, err := bot.NewTelegram(opts.TelegramBotApiKey,
		userService,
		youtubeService,
		mediaService,
		rssService,
		googleDriveClient,
		opts.BaseURL,
	)

	if err != nil {
		log.WithError(err).Fatal("cannot init telegram bot")
	}

	tgBot.Run()

	h, err := handler.NewHandler(userService,
		mediaService,
		googleDriveClient,
		tgBot,
		rssService,
	)

	if err != nil {
		log.WithError(err).Fatal("cannot init server handler")
	}

	srv := server.Server{
		Handler: h,
	}

	if err := srv.Run(":9000"); err != nil {
		log.WithError(err).Fatal("server fatal error")
	}

}
