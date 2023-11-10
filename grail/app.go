// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"github.com/emersion/go-sasl"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/grr"
	"goki.dev/icons"
	"goki.dev/ki/v2"
)

// App is an email client app.
type App struct {
	gi.Frame

	// Auth is the [sasl.Client] authentication for sending messages
	Auth sasl.Client

	// Message is the current message we are editing
	Message Message

	// Messages are the messages we have fetched from the server that we can read
	Messages []*Message
}

// needed for interface import
var _ ki.Ki = (*App)(nil)

func (a *App) TopAppBar(tb *gi.TopAppBar) {
	gi.DefaultTopAppBarStd(tb)
	giv.NewFuncButton(tb, a.Compose).SetIcon(icons.Send)
}

func (a *App) ConfigWidget(sc *gi.Scene) {
	if a.HasChildren() {
		return
	}

	grr.Log0(a.GetMail())

	updt := a.UpdateStart()

	sp := gi.NewSplits(a)

	list := gi.NewFrame(sp, "list").SetLayout(gi.LayoutVert)
	for _, msg := range a.Messages {
		fr := gi.NewFrame(list).SetLayout(gi.LayoutVert)
		gi.NewLabel(fr, "subject").SetType(gi.LabelTitleMedium).SetText(msg.Subject)
		gi.NewLabel(fr, "body").SetType(gi.LabelBodyMedium).SetText(msg.Body)
	}

	mail := gi.NewFrame(sp, "mail").SetLayout(gi.LayoutVert)
	gi.NewLabel(mail).SetText("Message goes here")

	sp.SetSplits(0.3, 0.7)
	a.UpdateEndLayout(updt)
}

func (a *App) GetMail() error {
	err := a.AuthGmail()
	if err != nil {
		return err
	}
	return a.GetMessages()
}
