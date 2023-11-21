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
	"goki.dev/grows/jsons"
	"goki.dev/kid"
	"golang.org/x/oauth2"
)

// Auth authorizes access to the user's mail and sets [App.AuthClient].
// If the user does not already have a saved auth token, it calls [SignIn].
func (a *App) Auth() error {
	tpath := filepath.Join(goosi.TheApp.AppPrefsDir(), "gmail-token.json")
	// exists, err := dirs.FileExists(tpath)
	// if err != nil {
	// 	return err
	// }

	// if !exists {
	err := a.SignIn()
	if err != nil {
		return err
	}
	// }

	err = jsons.Open(&a.AuthToken, tpath)
	if err != nil {
		return err
	}

	if gi.Prefs.User.Email == "" {
		return fmt.Errorf("email address not specified in preferences")
	}

	// if a.AuthToken.Expiry.Before(time.Now()) {
	// 	google.
	// }

	a.AuthClient = xoauth2.NewXoauth2Client(gi.Prefs.User.Email, a.AuthToken.AccessToken)
	return nil
}

// SignIn displays a dialog for the user to sign in with the platform of their choice.
func (a *App) SignIn() error {
	d := gi.NewBody().AddTitle("Sign in")
	ec := make(chan error)
	fun := func(token *oauth2.Token, userInfo *oidc.UserInfo) {
		tpath := filepath.Join(goosi.TheApp.AppPrefsDir(), "gmail-token.json")
		// TODO(kai/grail): figure out a more secure way to save the token
		err := jsons.Save(token, tpath)
		ec <- err
	}
	kid.Buttons(d, fun, "https://mail.google.com/")
	d.NewDialog(a).Run()
	return <-ec
}

/*
// AuthGmail authenticates the user with gmail.
func (a *App) AuthGmail() error { //gti:add

	if !exists {
		err := a.GetGmailRefreshToken()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetGmailRefreshToken uses the Google Oauth2 system to get and save a long-lived
// refresh token for the user that grants access to the gmail SMTP and IMAP servers.
func (a *App) GetGmailRefreshToken() error {
	ctx := context.Background()

	b, err := secretJSON.ReadFile("secret.json")
	if err != nil {
		return err
	}
	config, err := google.ConfigFromJSON(b, "https://mail.google.com/")
	if err != nil {
		return err
	}
	config.RedirectURL += ":5556"

	code := make(chan string)
	sm := http.NewServeMux()
	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code <- r.URL.Query().Get("code")
		w.Write([]byte("<h1>Authentication Successful</h1><p>You can close this browser tab and return to Grail</p>"))
	})
	// TODO(kai/grail): more graceful closing / error handling
	go http.ListenAndServe("127.0.0.1:5556", sm)

	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	// TODO(kai/grail): state
	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	goosi.TheApp.OpenURL(url)

	cs := <-code
	token, err := config.Exchange(ctx, cs, oauth2.VerifierOption(verifier))
	if err != nil {
		return err
	}

	return nil
}
*/
