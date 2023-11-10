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

	// Message is the current message we are editing
	Message Message

	// Auth is the [sasl.Client] authentication for sending messages
	Auth sasl.Client
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
	updt := a.UpdateStart()

	sp := gi.NewSplits(a)
	gi.NewFrame(sp, "list")

	mail := gi.NewFrame(sp, "mail").SetLayout(gi.LayoutVert)
	gi.NewLabel(mail).SetText("Message goes here")

	sp.SetSplits(0.3, 0.7)
	a.UpdateEndLayout(updt)

	err := grr.Log0(a.AuthGmail())
	if err == nil {
		grr.Log0(a.GetMessages())
	}
}
