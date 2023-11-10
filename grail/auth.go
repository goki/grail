// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"os"

	"goki.dev/grail/xoauth2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GRAIL_CLIENT_ID"),
	ClientSecret: os.Getenv("GRAIL_CLIENT_SECRET"),
	RedirectURL:  "",
	Scopes:       []string{"https://mail.google.com/"},
	Endpoint:     google.Endpoint,
}

// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error {
	c := xoauth2.NewXoauth2Client(a.Username, "")
	c.Start()
	return nil
}
