package rss

import (
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/htim/youpod"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Format int

const (
	JSON Format = iota
	XML
)

type Service struct {
	rootUrl     string
	fileService youpod.FileService
}

func NewService(rootUrl string, fileService youpod.FileService) *Service {
	return &Service{rootUrl: rootUrl, fileService: fileService}
}

func (s *Service) UserFeedUrl(user youpod.User) string {
	return fmt.Sprintf("%s/feed/%s", s.rootUrl, user.Username)
}

func (s *Service) UserFeed(user youpod.User, format Format) (string, error) {

	fmm := make([]youpod.FileMetadata, 0)

	for _, fid := range user.Files {
		m, err := s.fileService.GetFileMetadata(fid)
		if err != nil {
			return "", errors.Wrap(err, "cannot get file metadata")
		}
		if m == nil {
			log.
				WithField("file", fid).
				WithField("user", user.Username).
				Debug("no file metadata")
			continue
		}
		fmm = append(fmm, *m)
	}

	feed := &feeds.Feed{
		Title:       "YouPod feed",
		Link:        &feeds.Link{Href: "http://jmoiron.net/blog"},
		Description: "YouTube videos converted into podcasts",
		Created:     time.Now(),
	}

	items := make([]*feeds.Item, 0)

	for _, fm := range fmm {

		if fm.ID == "" {
			continue
		}

		fileLink := fmt.Sprintf("%s/files/%s/%s", s.rootUrl, user.Username, fm.ID)

		link := &feeds.Link{
			Href: fileLink,
		}
		enclosure := &feeds.Enclosure{
			Url:    fileLink,
			Length: strconv.FormatInt(fm.Length, 10),
			Type:   "audio/mpeg",
		}

		item := &feeds.Item{
			Id:        fileLink,
			Title:     fm.Name,
			Created:   time.Now(),
			Enclosure: enclosure,
			Link:      link,
		}

		items = append(items, item)
	}

	feed.Items = items

	var output string
	var err error

	switch format {
	case JSON:
		output, err = feed.ToJSON()
		if err != nil {
			return "", errors.Wrap(err, "cannot format rss")
		}

	case XML:
		output, err = feed.ToAtom()
		if err != nil {
			return "", errors.Wrap(err, "cannot format rss")
		}

	}

	return output, nil
}
