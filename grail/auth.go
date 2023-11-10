// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"goki.dev/goosi"
	"goki.dev/grail/xoauth2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error { //gti:add
	ctx := context.Background()

	b, err := os.ReadFile("../../grail/secret.json")
	if err != nil {
		return err
	}
	config, err := google.ConfigFromJSON(b, "https://mail.google.com/")
	if err != nil {
		return err
	}
	port := rand.Intn(10_000)
	config.RedirectURL += ":" + strconv.Itoa(port)

	code := make(chan string)
	sm := http.NewServeMux()
	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code <- r.URL.Query().Get("code")
	})
	go http.ListenAndServe(config.RedirectURL, sm)

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

	c := xoauth2.NewXoauth2Client(a.Username, token.AccessToken)
	c.Start()
	return nil
}
