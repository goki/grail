// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"context"
	"fmt"
	"os"

	"goki.dev/goosi"
	"goki.dev/grail/xoauth2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GRAIL_CLIENT_ID"),
	ClientSecret: os.Getenv("GRAIL_CLIENT_SECRET"),
	RedirectURL:  "http://127.0.0.1:3000",
	Scopes:       []string{"https://www.googleapis.com/auth/gmail.labels"},
	Endpoint:     google.Endpoint,
}

// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error { //gti:add
	ctx := context.Background()

	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	goosi.TheApp.OpenURL(url)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return err
	}
	token, err := googleOauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return err
	}

	c := xoauth2.NewXoauth2Client(a.Username, token.AccessToken)
	c.Start()
	return nil
}
