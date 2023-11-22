// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"io"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"goki.dev/cursors"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/girl/abilities"
	"goki.dev/girl/styles"
	"goki.dev/glide/gidom"
	"goki.dev/goosi/events"
	"goki.dev/grr"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"golang.org/x/oauth2"
)

// App is an email client app.
type App struct {
	gi.Frame

	// AuthToken is [oauth2.Token] for the signed in user.
	AuthToken *oauth2.Token

	// AuthClient is the [sasl.Client] authentication for sending messages
	AuthClient sasl.Client

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

	updt := a.UpdateStart()

	sp := gi.NewSplits(a)

	var ml *gi.Frame
	var msv *giv.StructView
	var mb *gi.Frame

	list := gi.NewFrame(sp, "list").Style(func(s *styles.Style) {
		s.Direction = styles.Column
	})
	for _, msg := range a.Messages {
		msg := msg
		fr := gi.NewFrame(list).Style(func(s *styles.Style) {
			s.Direction = styles.Column
		})

		fr.Style(func(s *styles.Style) {
			s.SetAbilities(true, abilities.Activatable, abilities.Hoverable)
			s.Cursor = cursors.Pointer
		})
		fr.OnClick(func(e events.Event) {
			a.ReadMessage = msg
			a.UpdateReadMessage(ml, msv, mb)
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

	ml = gi.NewFrame(sp, "mail")
	ml.Style(func(s *styles.Style) {
		s.Direction = styles.Column
	})
	msv = giv.NewStructView(ml, "msv")
	mb = gi.NewFrame(ml, "mb")
	mb.Style(func(s *styles.Style) {
		s.Direction = styles.Column
	})

	a.UpdateReadMessage(ml, msv, mb)

	sp.SetSplits(0.3, 0.7)
	a.UpdateEndLayout(updt)
}

// UpdateReadMessage updates the view of the message currently being read.
func (a *App) UpdateReadMessage(ml *gi.Frame, msv *giv.StructView, mb *gi.Frame) {
	if a.ReadMessage == nil {
		return
	}

	msv.SetStruct(a.ReadMessage)

	updt := mb.UpdateStart()
	if mb.HasChildren() {
		mb.DeleteChildren(true)
	}

	mr := grr.Log(mail.CreateReader(a.ReadMessage.BodyReader))
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			grr.Log0(err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			ct, _ := grr.Log2(h.ContentType())
			switch ct {
			case "text/plain":
				grr.Log0(gidom.ReadMD(gidom.BaseContext(), mb, grr.Log(io.ReadAll(p.Body))))
			case "text/html":
				grr.Log0(gidom.ReadHTML(gidom.BaseContext(), mb, p.Body))
			}
		}
	}
	mb.Update()
	mb.UpdateEndLayout(updt)
}

func (a *App) GetMail() error {
	err := a.Auth()
	if err != nil {
		return err
	}
	err = a.GetMessages()
	if err != nil {
		return err
	}
	err = a.CacheMessages()
	if err != nil {
		return err
	}
	updt := a.UpdateStart()
	a.DeleteChildren(true)
	a.Update()
	a.UpdateEndLayout(updt)
	return nil
}
