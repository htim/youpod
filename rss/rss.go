package rss

import (
	"fmt"
	"github.com/htim/youpod"
	"github.com/pkg/errors"
	"strconv"
)

type Format int

const (
	JSON Format = iota
	XML
)

type Service struct {
	rootUrl     string
	fileService youpod.MetadataService
}

func NewService(rootUrl string, fileService youpod.MetadataService) *Service {
	return &Service{rootUrl: rootUrl, fileService: fileService}
}

func (s *Service) UserFeedUrl(user youpod.User) string {
	return fmt.Sprintf("%s/feed/%s", s.rootUrl, user.Username)
}

func (s *Service) UserFeed(user youpod.User, format Format) (string, error) {

	fmm := make([]youpod.FileMetadata, 0)

	for _, fid := range user.Files {
		m, err := s.fileService.GetFileMetadata(fid)
		if err == youpod.ErrMetadataNotFound {
			continue
		}
		if err != nil {
			return "", errors.Wrap(err, "cannot get file metadata")
		}
		fmm = append(fmm, m)
	}

	feed := Feed{
		Channel: Channel{
			Title:        "YouPod feed",
			Link:         "http://youpod.io",
			Description:  "YouTube videos converted into a podcasts",
			ItunesAuthor: "YouPod Bot",
			ItunesCategory: ItunesCategory{
				Text: "Technology",
			},
			ItunesExplicit: "no",
		},
	}

	items := make([]Item, 0)

	for _, fm := range fmm {

		fileLink := fmt.Sprintf("%s/files/%s/%s", s.rootUrl, user.Username, fm.FileID)

		item := Item{
			ItunesEpisodeType: "full",
			ItunesTitle:       fm.Name,
			Description: Description{
				Content: Content{
					Text: fm.Name,
				},
			},
			Enclosure: Enclosure{
				Length: strconv.FormatInt(fm.Length, 10),
				Type:   "audio/mpeg",
				Url:    fileLink,
			},
			Guid:           fileLink,
			ItunesExplicit: "no",
		}

		items = append(items, item)
	}

	feed.Channel.Items = items

	var output string
	var err error

	output, err = feed.ToXML()
	if err != nil {
		return "", errors.Wrap(err, "cannot format rss")
	}

	return output, nil
}
