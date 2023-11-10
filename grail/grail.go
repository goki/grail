// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"bytes"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"

	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/grr"
	"goki.dev/icons"
	"goki.dev/ki/v2"
)

// App is an email client app.
type App struct {
	gi.Frame
	Username string
	Auth     sasl.Client
}

// needed for interface import
var _ ki.Ki = (*App)(nil)

func (a *App) TopAppBar(tb *gi.TopAppBar) {
	gi.DefaultTopAppBarStd(tb)
	giv.NewFuncButton(tb, a.AuthGmail).SetIcon(icons.Mail)
}

func (a *App) ConfigWidget(sc *gi.Scene) {
	if a.HasChildren() {
		return
	}
	updt := a.UpdateStart()

	sp := gi.NewSplits(a)
	gi.NewFrame(sp, "list")

	mail := gi.NewFrame(sp, "mail").SetLayout(gi.LayoutVert)
	gi.NewLabel(mail).SetText("Message goes here")

	sp.SetSplits(0.3, 0.7)
	a.UpdateEndLayout(updt)
}

// SendMessage sends the current message
func (a *App) SendMessage() {
	grr.Log0(smtp.SendMail("smtp.gmail.com:587", a.Auth, "test@example.com", []string{"dst@example.com"}, &bytes.Buffer{}))
}
