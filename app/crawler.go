package app

import (
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/utahta/momoclo-channel/crawler"
	"github.com/utahta/momoclo-channel/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

type Crawler struct {
	context context.Context
	log     log.Logger
}

func newCrawler(ctx context.Context) *Crawler {
	return &Crawler{context: ctx, log: log.NewGaeLogger(ctx)}
}

func (c *Crawler) Crawl() error {
	var workQueue = make(chan bool, 20)
	defer close(workQueue)

	var wg sync.WaitGroup
	for _, cli := range c.crawlChannelClients() {
		workQueue <- true
		wg.Add(1)
		go func(ctx context.Context, cli *crawler.ChannelClient) {
			defer func() {
				<-workQueue
				wg.Done()
			}()

			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			cli.Channel.Client = urlfetch.Client(ctx)

			ch, err := cli.Fetch()
			if err != nil {
				c.log.Errorf("Failed to fetcc. error:%v", err)
				return
			}

			bin, err := json.Marshal(ch)
			if err != nil {
				c.log.Errorf("Failed to encode to json. error:%v", err)
				return
			}
			params := url.Values{"channel": {string(bin)}}

			c.pushTweetQueue(params)
			c.pushLineQueue(params)
		}(c.context, cli)
	}
	wg.Wait()

	return nil
}

func (c *Crawler) pushTweetQueue(params url.Values) {
	task := taskqueue.NewPOSTTask("/queue/tweet", params)
	_, err := taskqueue.Add(c.context, task, "queue-tweet")
	if err != nil {
		c.log.Errorf("Failed to add taskqueue for tweet. error:%v", err)
	}
}

func (c *Crawler) pushLineQueue(params url.Values) {
	task := taskqueue.NewPOSTTask("/queue/line", params)
	_, err := taskqueue.Add(c.context, task, "queue-line")
	if err != nil {
		c.log.Errorf("Failed to add taskqueue for line. error:%v", err)
	}
}

func (c *Crawler) crawlChannelClients() []*crawler.ChannelClient {
	return []*crawler.ChannelClient{
		crawler.NewTamaiBlogChannelClient(nil),
		crawler.NewMomotaBlogChannelClient(nil),
		crawler.NewAriyasuBlogChannelClient(nil),
		crawler.NewSasakiBlogChannelClient(nil),
		crawler.NewTakagiBlogChannelClient(nil),
		crawler.NewAeNewsChannelClient(),
		crawler.NewYoutubeChannelClient(),
	}
}