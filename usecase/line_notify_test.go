package usecase_test

import (
	"testing"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
	"github.com/utahta/momoclo-channel/adapter/gateway/api/linenotify"
	"github.com/utahta/momoclo-channel/container"
	"github.com/utahta/momoclo-channel/domain/model"
	"github.com/utahta/momoclo-channel/infrastructure/event/eventtest"
	"github.com/utahta/momoclo-channel/lib/aetestutil"
	"github.com/utahta/momoclo-channel/usecase"
	"google.golang.org/appengine/aetest"
)

func TestLineNotify_Do(t *testing.T) {
	ctx, done, err := aetestutil.NewContex(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	taskQueue := eventtest.NewTaskQueue()
	repo := container.Repository(ctx).LineNotificationRepository()
	u := usecase.NewLineNotify(container.Logger(ctx).AE(), taskQueue, linenotify.NewNop(), repo)

	validationTests := []struct {
		params usecase.LineNotifyParams
	}{
		{usecase.LineNotifyParams{Request: model.LineNotifyRequest{ID: "id-1"}}},
		{usecase.LineNotifyParams{Request: model.LineNotifyRequest{AccessToken: "token"}}},
		{usecase.LineNotifyParams{Request: model.LineNotifyRequest{
			ID: "id-2", AccessToken: "token",
		}}},
		{usecase.LineNotifyParams{Request: model.LineNotifyRequest{
			ID: "id-3", AccessToken: "token", Messages: []model.LineNotifyMessage{
				{Text: ""},
			},
		}}},
		{usecase.LineNotifyParams{Request: model.LineNotifyRequest{
			ID: "id-4", AccessToken: "token", Messages: []model.LineNotifyMessage{
				{Text: "hello", ImageURL: "unknown"},
			},
		}}},
	}

	for _, test := range validationTests {
		err = u.Do(test.params)
		if errs, ok := errors.Cause(err).(validator.ValidationErrors); !ok {
			t.Errorf("Expected validation error, got %v. params:%v", errs, test.params)
		}
	}

	err = u.Do(usecase.LineNotifyParams{Request: model.LineNotifyRequest{
		ID: "id-1", AccessToken: "token", Messages: []model.LineNotifyMessage{
			{Text: "hello"},
			{Text: " ", ImageURL: "http://localhost/a"},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	if len(taskQueue.Tasks) != 1 {
		t.Errorf("Expected taskqueue length 1, got %v", len(taskQueue.Tasks))
	}
}