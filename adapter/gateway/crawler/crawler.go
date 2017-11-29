package crawler

import (
	"context"

	"github.com/pkg/errors"
	"github.com/utahta/momoclo-channel/domain/model"
	"github.com/utahta/momoclo-crawler"
	"google.golang.org/appengine/urlfetch"
)

type handler struct {
	ctx context.Context
}

// New returns model.FeedFetcher that wraps momoclo-crawler
func New(ctx context.Context) model.FeedFetcher {
	return &handler{ctx}
}

func (c *handler) Fetch(code model.FeedCode, maxItemNum int, latestURL string) ([]model.FeedItem, error) {
	const errTag = "client.Fetch failed"
	var (
		cli  *crawler.ChannelClient
		err  error
		opts = crawler.WithHTTPClient(urlfetch.Client(c.ctx))
	)

	switch code {
	case model.FeedCodeTamai:
		cli, err = crawler.NewTamaiBlogChannelClient(maxItemNum, latestURL, opts)
	case model.FeedCodeMomota:
		cli, err = crawler.NewMomotaBlogChannelClient(maxItemNum, latestURL, opts)
	case model.FeedCodeAriyasu:
		cli, err = crawler.NewAriyasuBlogChannelClient(maxItemNum, latestURL, opts)
	case model.FeedCodeSasaki:
		cli, err = crawler.NewSasakiBlogChannelClient(maxItemNum, latestURL, opts)
	case model.FeedCodeTakagi:
		cli, err = crawler.NewTakagiBlogChannelClient(maxItemNum, latestURL, opts)
	case model.FeedCodeHappyclo:
		cli, err = crawler.NewHappycloChannelClient(latestURL, opts)
	case model.FeedCodeAeNews:
		cli, err = crawler.NewAeNewsChannelClient(opts)
	case model.FeedCodeYoutube:
		cli, err = crawler.NewYoutubeChannelClient(opts)
	default:
		err = errors.Errorf("code:%s did not support", code)
	}

	if err != nil {
		return nil, errors.Wrap(err, errTag)
	}

	channel, err := cli.Fetch()
	if err != nil {
		return nil, errors.Wrap(err, errTag)
	}

	var items = make([]model.FeedItem, len(channel.Items))
	for i, feed := range channel.Items {
		item := model.FeedItem{
			Title:       channel.Title,
			URL:         channel.URL,
			EntryTitle:  feed.Title,
			EntryURL:    feed.URL,
			PublishedAt: feed.PublishedAt,
		}

		item.ImageURLs = make([]string, len(feed.Images))
		for i, image := range feed.Images {
			item.ImageURLs[i] = image.URL
		}

		item.VideoURLs = make([]string, len(feed.Videos))
		for i, video := range feed.Videos {
			item.VideoURLs[i] = video.URL
		}

		items[i] = item
	}
	return items, nil
}