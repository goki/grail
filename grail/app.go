// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"github.com/emersion/go-sasl"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/girl/styles"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"golang.org/x/oauth2"
)

// App is an email client app.
type App struct {
	gi.Frame

	// AuthToken contains the [oauth2.Token] for each account.
	AuthToken map[string]*oauth2.Token

	// AuthClient contains the [sasl.Client] authentication for sending messages for each account.
	AuthClient map[string]sasl.Client

	// ComposeMessage is the current message we are editing
	ComposeMessage *Message

	// Cache contains the cache data, keyed by account and then mailbox.
	Cache map[string]map[string][]*CacheData

	// ReadMessage is the current message we are reading
	ReadMessage *CacheData

	// The current email account
	CurEmail string

	// The current mailbox
	CurMailbox string
}

// needed for interface import
var _ ki.Ki = (*App)(nil)

func (a *App) OnInit() {
	a.AuthToken = map[string]*oauth2.Token{}
	a.AuthClient = map[string]sasl.Client{}
	a.AppStyles()
}

func (a *App) AppStyles() {
	a.Style(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
}

func (a *App) TopAppBar(tb *gi.TopAppBar) {
	gi.DefaultTopAppBarStd(tb)
	giv.NewFuncButton(tb, a.Compose).SetIcon(icons.Send)
}

func (a *App) ConfigWidget(sc *gi.Scene) {
	if a.HasChildren() {
		return
	}

	updt := a.UpdateStart()

	sp := gi.NewSplits(a, "splits")

	mbox := giv.NewTreeView(sp, "mbox")
	mbox.SetRootView(mbox)

	gi.NewFrame(sp, "list").Style(func(s *styles.Style) {
		s.Direction = styles.Column
	})

	ml := gi.NewFrame(sp, "mail")
	ml.Style(func(s *styles.Style) {
		s.Direction = styles.Column
	})
	giv.NewStructView(ml, "msv").SetReadOnly(true)
	gi.NewFrame(ml, "mb").Style(func(s *styles.Style) {
		s.Direction = styles.Column
	})

	// a.UpdateReadMessage(ml, msv, mb)

	sp.SetSplits(0.1, 0.2, 0.7)
	a.UpdateEndLayout(updt)
}

// // UpdateReadMessage updates the view of the message currently being read.
// func (a *App) UpdateReadMessage(ml *gi.Frame, msv *giv.StructView, mb *gi.Frame) {
// 	if a.ReadMessage == nil {
// 		return
// 	}

// 	msv.SetStruct(a.ReadMessage)

// 	updt := mb.UpdateStart()
// 	if mb.HasChildren() {
// 		mb.DeleteChildren(true)
// 	}

// 	mr := grr.Log(mail.CreateReader(a.ReadMessage.BodyReader))
// 	for {
// 		p, err := mr.NextPart()
// 		if err == io.EOF {
// 			break
// 		} else if err != nil {
// 			grr.Log0(err)
// 		}

// 		switch h := p.Header.(type) {
// 		case *mail.InlineHeader:
// 			ct, _ := grr.Log2(h.ContentType())
// 			switch ct {
// 			case "text/plain":
// 				grr.Log0(gidom.ReadMD(gidom.BaseContext(), mb, grr.Log(io.ReadAll(p.Body))))
// 			case "text/html":
// 				grr.Log0(gidom.ReadHTML(gidom.BaseContext(), mb, p.Body))
// 			}
// 		}
// 	}
// 	mb.Update()
// 	mb.UpdateEndLayout(updt)
// }

func (a *App) GetMail() error {
	err := a.Auth()
	if err != nil {
		return err
	}
	go func() {
		err = a.CacheMessages()
		if err != nil {
			gi.ErrorDialog(a, err, "Error caching messages").Run()
		}
	}()
	updt := a.UpdateStart()
	a.DeleteChildren(true)
	a.Update()
	a.UpdateEndLayout(updt)
	return nil
}
