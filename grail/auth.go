// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"fmt"
	"path/filepath"

	"github.com/coreos/go-oidc/v3/oidc"
	"goki.dev/gi/v2/gi"
	"goki.dev/goosi"
	"goki.dev/grail/xoauth2"
	"goki.dev/kid"
	"golang.org/x/oauth2"
)

// Auth authorizes access to the user's mail and sets [App.AuthClient].
// If the user does not already have a saved auth token, it calls [SignIn].
func (a *App) Auth() error {
	err := a.SignIn()
	if err != nil {
		return err
	}

	if gi.Prefs.User.Email == "" {
		return fmt.Errorf("email address not specified in preferences")
	}

	a.AuthClient = xoauth2.NewXoauth2Client(gi.Prefs.User.Email, a.AuthToken.AccessToken)
	return nil
}

// SignIn displays a dialog for the user to sign in with the platform of their choice.
func (a *App) SignIn() error {
	d := gi.NewBody().AddTitle("Sign in")
	done := make(chan struct{})
	fun := func(token *oauth2.Token, userInfo *oidc.UserInfo) {
		a.AuthToken = token
		d.Close()
		done <- struct{}{}
	}
	kid.Buttons(d, &kid.ButtonsConfig{
		SuccessFunc: fun,
		TokenFile: func(provider string) string {
			return filepath.Join(goosi.TheApp.AppPrefsDir(), provider+"-token.json")
		},
		Scopes: map[string][]string{
			"google": {"https://mail.google.com/"},
		},
	})
	d.NewDialog(a).Run()
	<-done
	return nil
}
