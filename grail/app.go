// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"github.com/emersion/go-sasl"
	"goki.dev/cursors"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/girl/abilities"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/grr"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"goki.dev/mat32/v2"
)

// App is an email client app.
type App struct {
	gi.Frame

	// Auth is the [sasl.Client] authentication for sending messages
	Auth sasl.Client

	// ComposeMessage is the current message we are editing
	ComposeMessage *Message

	// ReadMessage is the current message we are reading
	ReadMessage *Message

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

	var msv *giv.StructView

	list := gi.NewFrame(sp, "list").Style(func(s *styles.Style) {
		s.MainAxis = mat32.Y
	})
	for _, msg := range a.Messages {
		msg := msg
		fr := gi.NewFrame(list).Style(func(s *styles.Style) {
			s.MainAxis = mat32.Y
		})

		fr.Style(func(s *styles.Style) {
			s.SetAbilities(true, abilities.Activatable, abilities.Hoverable)
			s.Cursor = cursors.Pointer
		})
		fr.OnClick(func(e events.Event) {
			a.ReadMessage = msg
			msv.SetStruct(a.ReadMessage)
		})

		gi.NewLabel(fr, "subject").SetType(gi.LabelTitleMedium).SetText(msg.Subject).
			Style(func(s *styles.Style) {
				s.SetAbilities(false, abilities.Selectable, abilities.DoubleClickable)
				s.Cursor = cursors.None
			})
		gi.NewLabel(fr, "body").SetType(gi.LabelBodyMedium).SetText(msg.Body).
			Style(func(s *styles.Style) {
				s.SetAbilities(false, abilities.Selectable, abilities.DoubleClickable)
				s.Cursor = cursors.None
			})
	}

	mail := gi.NewFrame(sp, "mail").Style(func(s *styles.Style) {
		s.MainAxis = mat32.Y
	})
	msv = giv.NewStructView(mail).SetStruct(a.ReadMessage)

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
