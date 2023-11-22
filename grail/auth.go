// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"path/filepath"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"goki.dev/gi/v2/gi"
	"goki.dev/grail/xoauth2"
	"goki.dev/kid"
	"golang.org/x/oauth2"
)

// Auth authorizes access to the user's mail and sets [App.AuthClient].
// If the user does not already have a saved auth token, it calls [SignIn].
func (a *App) Auth() error {
	email, err := a.SignIn()
	if err != nil {
		return err
	}

	a.AuthClient[email] = xoauth2.NewXoauth2Client(email, a.AuthToken[email].AccessToken)
	return nil
}

// SignIn displays a dialog for the user to sign in with the platform of their choice.
// It returns the user's email address.
func (a *App) SignIn() (string, error) {
	d := gi.NewBody().AddTitle("Sign in")
	email := make(chan string)
	fun := func(token *oauth2.Token, userInfo *oidc.UserInfo) {
		if !slices.Contains(Prefs.Accounts, userInfo.Email) {
			Prefs.Accounts = append(Prefs.Accounts, userInfo.Email)
		}
		a.AuthToken[userInfo.Email] = token
		d.Close()
		email <- userInfo.Email
	}
	kid.Buttons(d, &kid.ButtonsConfig{
		SuccessFunc: fun,
		TokenFile: func(provider string) string {
			return filepath.Join(gi.AppPrefsDir(), "auth", provider+"-token.json")
		},
		Scopes: map[string][]string{
			"google": {"https://mail.google.com/"},
		},
	})
	d.NewDialog(a).Run()
	return <-email, nil
}
