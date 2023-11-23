// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"github.com/yuin/goldmark"
	"goki.dev/cursors"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/girl/abilities"
	"goki.dev/girl/styles"
	"goki.dev/glide/gidom"
	"goki.dev/goosi/events"
	"goki.dev/grr"
)

// Message contains the relevant information for an email message.
type Message struct {
	From    []*mail.Address `view:"inline"`
	To      []*mail.Address `view:"inline"`
	Subject string
	// only for sending
	Body string `view:"text-editor" viewif:"BodyReader==nil"`
	// only for receiving
	BodyReader imap.LiteralReader `view:"-"`
}

// Compose pulls up a dialog to send a new message
func (a *App) Compose() { //gti:add
	a.ComposeMessage = &Message{}
	a.ComposeMessage.From = []*mail.Address{{Address: Prefs.Accounts[0]}}
	b := gi.NewBody().AddTitle("Send message")
	giv.NewStructView(b).SetStruct(a.ComposeMessage)
	b.AddBottomBar(func(pw gi.Widget) {
		b.AddCancel(pw)
		b.AddOk(pw).SetText("Send").OnClick(func(e events.Event) {
			a.SendMessage()
		})
	})
	b.NewFullDialog(a).Run()
}

// SendMessage sends the current message
func (a *App) SendMessage() error { //gti:add
	if len(a.ComposeMessage.From) != 1 {
		return fmt.Errorf("expected 1 sender, but got %d", len(a.ComposeMessage.From))
	}
	email := a.ComposeMessage.From[0].Address

	var b bytes.Buffer

	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", a.ComposeMessage.From)
	h.SetAddressList("To", a.ComposeMessage.To)
	h.SetSubject(a.ComposeMessage.Subject)

	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		return err
	}

	tw, err := mw.CreateInline()
	if err != nil {
		return err
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		return err
	}
	err = goldmark.Convert([]byte(a.ComposeMessage.Body), w)
	if err != nil {
		return err
	}
	w.Close()
	tw.Close()

	to := make([]string, len(a.ComposeMessage.To))
	for i, t := range a.ComposeMessage.To {
		to[i] = t.Address
	}

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		a.AuthClient[email],
		email,
		to,
		&b,
	)
	if err != nil {
		se := err.(*smtp.SMTPError)
		slog.Error("error sending message: SMTP error:", "code", se.Code, "enhancedCode", se.EnhancedCode, "message", se.Message)
	}
	return err
}

// UpdateMessageList updates the message list from [App.Cache].
func (a *App) UpdateMessageList() {
	cached := a.Cache[a.CurEmail][a.CurMailbox]

	list := a.FindPath("splits/list").(*gi.Frame)

	if list.NumChildren() > 100 {
		return
	}

	updt := list.UpdateStartAsync()

	list.DeleteChildren(true)

	for i, cd := range cached {
		cd := cd

		if i > 100 {
			break
		}

		fr := gi.NewFrame(list).Style(func(s *styles.Style) {
			s.Direction = styles.Column
		})

		fr.Style(func(s *styles.Style) {
			s.SetAbilities(true, abilities.Activatable, abilities.Hoverable)
			s.Cursor = cursors.Pointer
		})
		fr.OnClick(func(e events.Event) {
			a.ReadMessage = cd
			grr.Log0(a.UpdateReadMessage())
		})

		ftxt := ""
		for _, f := range cd.From {
			ftxt += f.Name + " "
		}

		gi.NewLabel(fr, "from").SetType(gi.LabelTitleMedium).SetText(ftxt).
			Style(func(s *styles.Style) {
				s.SetAbilities(false, abilities.Selectable, abilities.DoubleClickable)
				s.Cursor = cursors.None
			})
		gi.NewLabel(fr, "subject").SetType(gi.LabelBodyMedium).SetText(cd.Subject).
			Style(func(s *styles.Style) {
				s.SetAbilities(false, abilities.Selectable, abilities.DoubleClickable)
				s.Cursor = cursors.None
			})
	}

	list.Update()
	list.UpdateEndAsyncLayout(updt)
}

// UpdateReadMessage updates the view of the message currently being read.
func (a *App) UpdateReadMessage() error {
	msv := a.FindPath("splits/mail/msv").(*giv.StructView)
	msv.SetStruct(a.ReadMessage)

	mb := a.FindPath("splits/mail/mb").(*gi.Frame)
	updt := mb.UpdateStart()
	mb.DeleteChildren(true)

	bemail := FilenameBase32(a.CurEmail)
	// there can be flags at the end of the filename, so we have to glob it
	glob := filepath.Join(gi.AppPrefsDir(), "mail", bemail, a.CurMailbox, "cur", a.ReadMessage.Filename+"*")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(matches) != 1 {
		return fmt.Errorf("expected 1 match for filepath glob but got %d: %s", len(matches), glob)
	}

	f, err := os.Open(matches[0])
	if err != nil {
		return err
	}
	defer f.Close()

	mr, err := mail.CreateReader(f)
	if err != nil {
		return err
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			ct, _, err := h.ContentType()
			if err != nil {
				return err
			}
			switch ct {
			case "text/plain":
				err := gidom.ReadMD(gidom.BaseContext(), mb, grr.Log(io.ReadAll(p.Body)))
				if err != nil {
					return err
				}
			case "text/html":
				err := gidom.ReadHTML(gidom.BaseContext(), mb, p.Body)
				if err != nil {
					return err
				}
			}
		}
	}
	mb.Update()
	mb.UpdateEndLayout(updt)
	return nil
}

/*
// GetMessages fetches the messages from the server
func (a *App) GetMessages() error { //gti:add
	c, err := imapclient.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	err = c.Authenticate(a.AuthClient)
	if err != nil {
		return err
	}

	ibox, err := c.Select("INBOX", false)
	if err != nil {
		return err
	}

	// Get the last 40 messages
	from := uint32(1)
	to := ibox.Messages
	if ibox.Messages > 39 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = ibox.Messages - 39
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	var sect imap.BodySectionName

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, sect.FetchItem()}, messages)
	}()

	a.Messages = make([]*Message, 0)
	for msg := range messages {

		from := make([]*mail.Address, len(msg.Envelope.From))
		for i, fr := range msg.Envelope.From {
			from[i] = &mail.Address{Name: fr.PersonalName, Address: fr.Address()}
		}
		to := make([]*mail.Address, len(msg.Envelope.To))
		for i, fr := range msg.Envelope.To {
			to[i] = &mail.Address{Name: fr.PersonalName, Address: fr.Address()}
		}

		m := &Message{
			From:       from,
			To:         to,
			Subject:    msg.Envelope.Subject,
			BodyReader: msg.GetBody(&sect),
		}
		a.Messages = append(a.Messages, m)
	}
	slices.Reverse(a.Messages)

	if err := <-done; err != nil {
		return err
	}
	return nil
}
*/
