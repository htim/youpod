package rss

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {

	f := &Feed{
		Channel: Channel{
			Title:        "test_title",
			Link:         "test_link",
			Language:     "ru",
			Copyright:    "c",
			ItunesAuthor: "email@email.com",
			Description:  "Youtube feed",
			ItunesType:   "test_type",
			ItunesImage: ItunesImage{
				Href: "test_href",
			},
		},
	}

	items := []Item{
		{
			ItunesEpisodeType: "video",
			ItunesTitle:       "test_title_item",
			Description: Description{
				Content: Content{
					Text: `<a href="http://example.org">My Example Website</a>`,
				},
			},
			Enclosure: Enclosure{
				Length: "test_length",
				Type:   "audio/mpeg",
				Url:    "url",
			},
		},
		{
			ItunesEpisodeType: "video",
			ItunesTitle:       "test_title_item",
			Description: Description{
				Content: Content{
					Text: `<a href="http://example.org">My Example Website</a>`,
				},
			},
		},
	}

	Items = items

	xml, err := ToXML()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(xml)

}
