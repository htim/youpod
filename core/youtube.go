package core

type (
	YoutubeService interface {
		Download(owner User, link string) (File, error)
		Cleanup(f File)
	}
)
