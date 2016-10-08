package controller

import (
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/utahta/momoclo-channel/appengine/model"
	"github.com/utahta/momoclo-channel/linenotify"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

// LINE Notify と連携する
func LinenotifyOn(ctx context.Context, w http.ResponseWriter, req *http.Request) *Error {
	reqAuth, err := linenotify.NewRequestAuthorization(os.Getenv("LINENOTIFY_CLIENT_ID"), buildURL(req.URL, "/linenotify/callback"))
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{Name: "state", Value: reqAuth.State, Expires: time.Now().Add(60 * time.Second), Secure: true})

	err = reqAuth.Redirect(w, req)
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}
	return nil
}

// LINE Notify の連携を解除する
func LinenotifyOff(ctx context.Context, w http.ResponseWriter, req *http.Request) *Error {
	// Using feature that provided in official.
	http.Redirect(w, req, "https://notify-bot.line.me/my/", http.StatusFound)
	return nil
}

func LinenotifyCallback(ctx context.Context, w http.ResponseWriter, req *http.Request) *Error {
	params, err := linenotify.ParseCallbackParameters(req)
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}

	state, err := req.Cookie("state")
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}

	if params.State != state.Value {
		return newError(errors.New("Invalid csrf token."), http.StatusBadRequest)
	}

	reqToken := linenotify.NewRequestToken(
		params.Code,
		buildURL(req.URL, "/linenotify/callback"),
		os.Getenv("LINENOTIFY_CLIENT_ID"),
		os.Getenv("LINENOTIFY_CLIENT_SECRET"),
	)
	reqToken.Client = urlfetch.Client(ctx)

	token, err := reqToken.Get()
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}

	ln, err := model.NewLineNotification(token)
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}
	ln.Put(ctx) // save to datastore

	t, err := template.New("callback").Parse("<html><body><h1>通知ノフ設定オンにしました（・Θ・）</h1></body></html>")
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}
	err = t.Execute(w, nil)
	if err != nil {
		return newError(err, http.StatusInternalServerError)
	}
	return nil
}