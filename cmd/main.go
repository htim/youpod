package main

import (
	gdrive "github.com/htim/youpod/google_drive"
	"github.com/htim/youpod/rss"
	"github.com/htim/youpod/server"
	"github.com/htim/youpod/server/handler"
	"github.com/htim/youpod/store/bolt"
	"github.com/htim/youpod/telegram"
	"github.com/htim/youpod/youtube"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	client := bolt.NewClient("youpod.db")

	if err := client.Open(); err != nil {
		log.WithError(err).Fatal("cannot open bolt client")
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.WithError(err).Error("cannot close bolt client")
		}
	}()

	userService, err := bolt.NewUserService(client)
	if err != nil {
		log.WithError(err).Fatal("cannot init UserService")
	}

	clientID := "101230901477-0g3vef4t1f4rc2urrf0sqcc0cu0l36qd.apps.googleusercontent.com"
	clientSecret := "sVnwe-EtkOUSaUVwV9vKAtKn"
	redirectUrl := "http://localhost:9000/gdrive/callback"

	googleDriveClient := gdrive.NewClient(
		userService,
		clientID,
		clientSecret,
		redirectUrl,
	)

	fileService := bolt.NewFileService(client, userService, googleDriveClient, "YouPod")

	youtubeService, err := youtube.NewService()
	if err != nil {
		log.WithError(err).Fatal("cannot init youtube service")
	}

	rssService := rss.NewService("http://localhost:9000", fileService)

	youPod, err := telegram.NewYouPod("963489157:AAH_ua7yUh4EkOvuHT1TNZmwda6joGQ7gQ4",
		userService,
		fileService,
		youtubeService,
		rssService,
		googleDriveClient,
		"http://localhost:9000",
	)

	if err != nil {
		log.WithError(err).Fatal("cannot init telegram bot")
	}

	youPod.Run()

	h := handler.Handler{
		UserService:     userService,
		GoogleDriveAuth: googleDriveClient,
		Bot:             youPod,
		Rss:             rssService,
	}

	srv := server.Server{
		Handler: h,
	}

	if err := srv.Run(":9000"); err != nil {
		log.WithError(err).Fatal("server fatal error")
	}

}
