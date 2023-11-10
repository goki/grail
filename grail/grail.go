// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/emersion/go-message/mail"

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
	giv.NewFuncButton(tb, a.SendMessage).SetIcon(icons.Send)
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
func (a *App) SendMessage() error { //gti:add
	var b bytes.Buffer

	from := []*mail.Address{{"Kai O'Reilly", "koreilly5297@gmail.com"}}
	to := []*mail.Address{{"Randall O'Reilly", "rcoreilly5@gmail.com"}}

	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetSubject("Subject")
	h.SetAddressList("To", to)

	tw, err := mail.CreateInlineWriter(&b, h)
	if err != nil {
		return err
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		return err
	}
	io.WriteString(w, "Body")
	w.Close()
	tw.Close()

	fmt.Println(b.String())

	return grr.Log0(smtp.SendMail(
		"smtp.googlemail.com:587",
		a.Auth,
		"koreilly5297@gmail.com",
		[]string{"koreilly5297@gmail.com", "rcoreilly5@gmail.com"},
		&b,
	))
}
