package backend

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/utahta/go-linenotify/auth"
	"github.com/utahta/momoclo-channel/adapter/handler"
	"github.com/utahta/momoclo-channel/container"
	"github.com/utahta/momoclo-channel/domain/event"
	"github.com/utahta/momoclo-channel/domain/model"
	"github.com/utahta/momoclo-channel/lib/config"
	"github.com/utahta/momoclo-channel/usecase"
)

// LineNotifyOn redirect to LINE Notify connection page
func LineNotifyOn(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	c, err := auth.New(config.C.LineNotify.ClientID, config.LineNotifyCallbackURL())
	if err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "state", Value: c.State, Expires: time.Now().Add(300 * time.Second), Secure: true})

	container.Logger(ctx).AE().Info("Redirect to LINE Notify connection page")

	err = c.Redirect(w, req)
	if err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}
}

// LineNotifyOff redirect to LINE Notify revoking page
func LineNotifyOff(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	container.Logger(ctx).AE().Info("Redirect to LINE Notify revoking page")

	// official url
	http.Redirect(w, req, "https://notify-bot.line.me/my/", http.StatusFound)
}

// LineNotifyCallback stores LINE Notify token
func LineNotifyCallback(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	params, err := auth.ParseRequest(req)
	if err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	state, err := req.Cookie("state")
	if err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	if params.State != state.Value {
		handler.Fail(ctx, w, errors.New("invalid csrf token"), http.StatusBadRequest)
		return
	}

	if err := container.Usecase(ctx).AddLineNotification().Do(usecase.AddLineNotificationParams{Code: params.Code}); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	t, err := template.New("callback").Parse("<html><body><h1>通知ノフ設定オンにしました（・Θ・）</h1></body></html>")
	if err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}
}

// LineNotifyBroadcast invokes broadcast line notification event
func LineNotifyBroadcast(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if err := req.ParseForm(); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	var messages []model.LineNotifyMessage
	if err := event.ParseTask(req.Form, &messages); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	params := usecase.LineNotifyBroadcastParams{Messages: messages}
	if err := container.Usecase(ctx).LineNotifyBroadcast().Do(params); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}
}

// LineNotify notify users of messages
func LineNotify(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()

	if err := req.ParseForm(); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	var request model.LineNotifyRequest
	if err := event.ParseTask(req.Form, &request); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}

	params := usecase.LineNotifyParams{Request: request}
	if err := container.Usecase(ctx).LineNotify().Do(params); err != nil {
		handler.Fail(ctx, w, err, http.StatusInternalServerError)
		return
	}
}