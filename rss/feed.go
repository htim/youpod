package rss

import (
	"encoding/xml"
)

const (
	itunesHeader = `<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:content="http://purl.org/rss/1.0/modules/content/">` + "\n"
	itunesFooter = "\n" + `</rss>`
)

type Feed struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	XMLName        xml.Name       `xml:"channel"`
	Title          string         `xml:"title"`
	Link           string         `xml:"link"`
	Language       string         `xml:"language"`
	Copyright      string         `xml:"copyright"`
	ItunesAuthor   string         `xml:"itunes:author"`
	Description    string         `xml:"description"`
	ItunesType     string         `xml:"itunes:type"`
	ItunesOwner    ItunesOwner    `xml:"itunes:owner"`
	ItunesImage    ItunesImage    `xml:"itunes:image"`
	ItunesCategory ItunesCategory `xml:"itunes:category"`
	ItunesExplicit string         `xml:"itunes:explicit"`
	Items          []Item
}

type ItunesOwner struct {
	ItunesName  string `xml:"itunes:name"`
	ItunesEmail string `xml:"itunes:email"`
}

type ItunesImage struct {
	Href string `xml:"href,attr"`
}

type ItunesCategory struct {
	Text string `xml:"text,attr"`
}

type Item struct {
	XMLName           xml.Name    `xml:"item"`
	ItunesEpisodeType string      `xml:"itunes:episodeType"`
	ItunesTitle       string      `xml:"itunes:title"`
	Description       Description `xml:"description"`
	Enclosure         Enclosure   `xml:"enclosure"`
	Guid              string      `xml:"guid"`
	PubDate           string      `xml:"pubDate"`
	ItunesDuration    string      `xml:"itunes:duration"`
	ItunesExplicit    string      `xml:"itunes:explicit"`
}

type Description struct {
	Content Content `xml:"content:encoded"`
}

type Content struct {
	Text string `xml:",cdata"`
}

type Enclosure struct {
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
	Url    string `xml:"url,attr"`
}

func (f *Feed) ToXML() (string, error) {
	marshalIndent, err := xml.MarshalIndent(f.Channel, "", "   ")
	if err != nil {
		return "", err
	}
	return xml.Header + itunesHeader + string(marshalIndent) + itunesFooter, nil
}
