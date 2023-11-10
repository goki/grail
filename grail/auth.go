// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"context"
	"embed"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"

	"goki.dev/goosi"
	"goki.dev/grail/xoauth2"
	"goki.dev/grows/jsons"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//go:embed secret.json
var secretJSON embed.FS

// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error { //gti:add
	ctx := context.Background()

	b, err := secretJSON.ReadFile("secret.json")
	if err != nil {
		return err
	}
	config, err := google.ConfigFromJSON(b, "https://mail.google.com/")
	if err != nil {
		return err
	}
	port := rand.Intn(10_000)
	sport := ":" + strconv.Itoa(port)
	config.RedirectURL += sport

	code := make(chan string)
	sm := http.NewServeMux()
	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code <- r.URL.Query().Get("code")
		w.Write([]byte("<h1>Authentication Successful</h1><p>You can close this browser tab and return to Grail</p>"))
	})
	// TODO(kai/grail): more graceful closing / error handling
	go http.ListenAndServe(sport, sm)

	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	goosi.TheApp.OpenURL(url)

	cs := <-code
	token, err := config.Exchange(ctx, cs, oauth2.VerifierOption(verifier))
	if err != nil {
		return err
	}

	tpath := filepath.Join(goosi.TheApp.AppPrefsDir(), "token.json")
	// TODO(kai/grail): figure out a more secure way to save the token
	err = jsons.Save(token, tpath)
	if err != nil {
		return err
	}

	a.Username = "koreilly5297@gmail.com"
	a.Auth = xoauth2.NewXoauth2Client(a.Username, token.AccessToken)

	return nil
}
