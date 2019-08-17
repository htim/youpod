package rss

import (
	"fmt"
	"github.com/htim/youpod"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"strconv"
)

type service struct {
	rootUrl     string
	fileService core.MetadataRepository
}

func NewService(rootUrl string, fileService core.MetadataRepository) core.RssService {
	return &service{rootUrl: rootUrl, fileService: fileService}
}

func (s *service) UserFeedUrl(user core.User) string {
	return fmt.Sprintf("%s/feed/%s", s.rootUrl, user.Username)
}

func (s *service) UserFeed(user core.User) (string, error) {

	fmm := make([]core.Metadata, 0)

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

	feed := &Feed{
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
