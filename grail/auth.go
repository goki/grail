// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"context"
	"os"

	"goki.dev/gi/v2/gi"
	"goki.dev/goosi"
	"goki.dev/goosi/events"
	"goki.dev/goosi/mimedata"
	"goki.dev/grail/xoauth2"
	"goki.dev/grr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GRAIL_CLIENT_ID"),
	ClientSecret: os.Getenv("GRAIL_CLIENT_SECRET"),
	RedirectURL:  "",
	Scopes:       []string{"https://www.googleapis.com/auth/gmail.labels"},
	Endpoint:     google.Endpoint,
}

// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error { //gti:add
	ctx := context.Background()

	resp, err := googleOauthConfig.DeviceAuth(ctx)
	if grr.Log0(err) != nil {
		return err
	}
	a.EventMgr().ClipBoard().Write(mimedata.NewText(resp.UserCode))
	cont := make(chan struct{})
	gi.NewDialog(a).Title("Paste code").Prompt("Paste the code copied to your clipboard when prompted").Ok().
		OnAccept(func(e events.Event) {
			cont <- struct{}{}
		})
	<-cont
	goosi.TheApp.OpenURL(resp.VerificationURI)
	token, err := googleOauthConfig.DeviceAccessToken(ctx, resp)
	if err != nil {
		return err
	}

	c := xoauth2.NewXoauth2Client(a.Username, token.AccessToken)
	c.Start()
	return nil
}
