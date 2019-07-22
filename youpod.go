package youpod

import "os"

//Source is an interface for a content source (e.g. YouTube video)
type Source interface {
	Download(url string, out os.File) error
}

//Publisher is an interface for a publishing platform (e.g. SoundCloud)
type Publisher interface {

}

