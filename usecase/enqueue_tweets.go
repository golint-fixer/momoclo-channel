package usecase

import (
	"github.com/pkg/errors"
	"github.com/utahta/momoclo-channel/domain"
	"github.com/utahta/momoclo-channel/domain/core"
	"github.com/utahta/momoclo-channel/domain/event"
	"github.com/utahta/momoclo-channel/domain/model"
	"github.com/utahta/momoclo-channel/domain/service/eventtask"
	"github.com/utahta/momoclo-channel/domain/service/feeditem"
	"github.com/utahta/momoclo-channel/log"
)

type (
	// EnqueueTweets use case
	EnqueueTweets struct {
		log        log.Logger
		taskQueue  event.TaskQueue
		transactor model.Transactor
		repo       model.TweetItemRepository
	}

	// EnqueueTweetsParams input parameters
	EnqueueTweetsParams struct {
		FeedItem model.FeedItem
	}
)

// NewEnqueueTweets returns EnqueueTweets use case
func NewEnqueueTweets(
	log log.Logger,
	taskQueue event.TaskQueue,
	transactor model.Transactor,
	repo model.TweetItemRepository) *EnqueueTweets {
	return &EnqueueTweets{
		log:        log,
		taskQueue:  taskQueue,
		transactor: transactor,
		repo:       repo,
	}
}

// Do converts feeds to tweet requests and enqueue it
func (use *EnqueueTweets) Do(params EnqueueTweetsParams) error {
	const errTag = "EnqueueTweets.Do failed"

	if err := core.Validate(params); err != nil {
		return errors.Wrap(err, errTag)
	}

	item := model.NewTweetItem(params.FeedItem)
	if use.repo.Exists(item.ID) {
		return nil // already enqueued
	}

	err := use.transactor.RunInTransaction(func(h model.PersistenceHandler) error {
		repo := use.repo.Tx(h)
		if _, err := repo.Find(item.ID); err != domain.ErrNoSuchEntity {
			return err
		}
		return repo.Save(item)
	}, nil)
	if err != nil {
		use.log.Errorf("%v: enqueue tweets feedItem:%v", errTag, params.FeedItem)
		return errors.Wrap(err, errTag)
	}

	requests := feeditem.ToTweetRequests(params.FeedItem)
	if len(requests) == 0 {
		use.log.Errorf("%v: invalid enqueue tweets feedItem:%v", errTag, params.FeedItem)
		return errors.Errorf("%v: invalid enqueue tweets", errTag)
	}

	task := eventtask.NewTweets(requests)
	if err := use.taskQueue.Push(task); err != nil {
		return errors.Wrap(err, errTag)
	}
	use.log.Infof("enqueue tweet requests:%#v", requests)

	return nil
}
