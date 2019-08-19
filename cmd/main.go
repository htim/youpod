package main

import (
	"github.com/htim/youpod/bot"
	"github.com/htim/youpod/server"
	"github.com/htim/youpod/server/handler"
	"github.com/htim/youpod/service/media"
	gdrive "github.com/htim/youpod/service/media/google_drive"
	"github.com/htim/youpod/service/rss"
	"github.com/htim/youpod/service/youtube"
	"github.com/htim/youpod/store/bolt"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"os"
)

var opts struct {
	ClientID     string `long:"client_id" env:"CLIENT_ID" description:"Google Drive client_id" required:"true"`
	ClientSecret string `long:"client_secret" env:"CLIENT_SECRET" description:"Google Drive client secret" required:"true"`

	TelegramBotApiKey string `long:"tg_bot_api_key" env:"TG_BOT_API_KEY" description:"Telegram Bot API Key" required:"true"`

	BaseURL          string `long:"base_url" env:"BASE_URL" description:"app base url" required:"true"`
	BoltRootDir      string `long:"bolt_root_dir" env:"BOLT_ROOT_DIR" description:"directory for boltdb" required:"true"`
	YoutubeOutputDir string `long:"youtube_output_dir" env:"YOUTUBE_OUTPUT_DIR" description:"tmp directory for youtube" required:"true"`
}

func main() {

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		log.WithError(err).Fatal("cannot parse options")
	}

	boltClient := bolt.NewClient(opts.BoltRootDir + "/youpod.db")

	if err := boltClient.Open(); err != nil {
		log.WithError(err).Fatal("cannot open bolt client")
	}
	defer func() {
		if err := boltClient.Close(); err != nil {
			log.WithError(err).Error("cannot close bolt client")
		}
	}()

	userRepository, err := bolt.NewUserRepository(boltClient)
	if err != nil {
		log.WithError(err).Fatal("cannot init userRepository")
	}

	googleDriveClient := gdrive.NewClient(
		userRepository,
		opts.ClientID,
		opts.ClientSecret,
		"http://localhost:9000"+"/gdrive/callback",
	)

	metadataRepository := bolt.NewMetadataRepository(boltClient, userRepository, "YouPod")

	youtubeService, err := youtube.NewService(opts.BoltRootDir)
	if err != nil {
		log.WithError(err).Fatal("cannot init youtube service")
	}

	rssService := rss.NewService(opts.BaseURL, metadataRepository)

	mediaService := media.NewService(
		metadataRepository,
		googleDriveClient,
	)

	tgBot, err := bot.NewTelegram(opts.TelegramBotApiKey,
		userRepository,
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

	h, err := handler.NewHandler(userRepository,
		mediaService,
		rssService,
		googleDriveClient,
		tgBot,
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
