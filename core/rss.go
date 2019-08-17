package core

type (
	RssService interface {
		UserFeedUrl(user User) string
		UserFeed(user User) (string, error)
	}
)
