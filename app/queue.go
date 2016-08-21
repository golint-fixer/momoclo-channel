package app

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
	"sync"

	"github.com/pkg/errors"
	"github.com/utahta/momoclo-channel/crawler"
	"github.com/utahta/momoclo-channel/log"
	"github.com/utahta/momoclo-channel/model"
	"github.com/utahta/momoclo-channel/twitter"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// Queue for crawler.Channel
type QueueHandler struct {
}

func (h *QueueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var err *Error
	defer err.Handle(ctx, w)

	ch, err := h.parseParams(r)
	if err != nil {
		return
	}

	switch r.URL.Path {
	case "/queue/tweet":
		err = h.serveTweet(ctx, ch)
	case "/queue/line":
		err = h.serveLine(ctx, ch)
	default:
		http.NotFound(w, r)
	}
}

func (h *QueueHandler) parseParams(r *http.Request) (*crawler.Channel, *Error) {
	var ch crawler.Channel
	if err := json.Unmarshal([]byte(r.FormValue("channel")), &ch); err != nil {
		return nil, newError(errors.Wrapf(err, "Failed to unmarshal."), http.StatusInternalServerError)
	}
	return &ch, nil
}

func (h *QueueHandler) serveTweet(ctx context.Context, ch *crawler.Channel) *Error {
	var wg sync.WaitGroup
	for _, item := range ch.Items {
		wg.Add(1)
		go func(item *crawler.ChannelItem) {
			defer wg.Done()

			if err := model.NewTweetItem(item).Put(ctx); err != nil {
				return
			}
			ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
			defer cancel()

			tw := twitter.NewTwitterClient(
				os.Getenv("TWITTER_CONSUMER_KEY"),
				os.Getenv("TWITTER_CONSUMER_SECRET"),
				os.Getenv("TWITTER_ACCESS_TOKEN"),
				os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
			)
			tw.Log = log.NewGaeLogger(ctx)
			tw.Api.HttpClient.Transport = &urlfetch.Transport{Context: ctx}

			tw.TweetItem(ch.Title, item)
		}(item)
	}
	wg.Wait()

	return nil
}

func (h *QueueHandler) serveLine(ctx context.Context, ch *crawler.Channel) *Error {
	return nil
}
