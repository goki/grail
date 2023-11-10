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
	"goki.dev/grr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GRAIL_CLIENT_ID"),
	ClientSecret: os.Getenv("GRAIL_CLIENT_SECRET"),
	RedirectURL:  "",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error { //gti:add
	ctx := context.Background()

	resp, err := googleOauthConfig.DeviceAuth(ctx)
	if grr.Log0(err) != nil {
		return err
	}
	fmt.Println(resp.UserCode)
	goosi.TheApp.OpenURL(resp.VerificationURI)
	token, err := googleOauthConfig.DeviceAccessToken(ctx, resp)
	if err != nil {
		return err
	}

	c := xoauth2.NewXoauth2Client(a.Username, token.AccessToken)
	c.Start()
	return nil
}
