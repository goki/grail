// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"bytes"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-maildir"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"github.com/yuin/goldmark"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
	"goki.dev/grows/jsons"
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

// CacheData contains the data stored for a cached message in the cached messages file.
// It contains basic information about the message so that it can be displayed in the
// mail list in the GUI.
type CacheData struct {
	imap.Envelope
	UID      uint32
	Filename string
}

// CacheMessages caches all of the messages from the server that
// have not already been cached. It caches them using maildir in
// the app's prefs directory.
func (a *App) CacheMessages() error {
	for _, account := range Prefs.Accounts {
		err := a.CacheMessagesForAccount(account)
		if err != nil {
			return err
		}
	}
	return nil
}

// CacheMessages caches all of the messages from the server that
// have not already been cached for the given email account. It
// caches them using maildir in the app's prefs directory.
func (a *App) CacheMessagesForAccount(email string) error {
	c, err := imapclient.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return fmt.Errorf("TLS dialing: %w", err)
	}
	defer c.Logout()

	err = c.Authenticate(a.AuthClient[email])
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}

	mailboxes, err := c.List("", "*", nil).Collect()
	if err != nil {
		return fmt.Errorf("getting mailboxes: %w", err)
	}

	fmt.Printf("Found %d mailboxes\n", len(mailboxes))
	for _, mbox := range mailboxes {
		fmt.Println("* " + mbox.Mailbox)
	}

	a.CurMailbox = "INBOX"

	return a.CacheMessagesForMailbox(c, email, "INBOX")
}

// CacheMessagesForMailbox caches all of the messages from the server
// that have not already been cached for the given email account and mailbox.
// It caches them using maildir in the app's prefs directory.
func (a *App) CacheMessagesForMailbox(c *imapclient.Client, email string, mailbox string) error {
	hemail := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(email))
	dir := maildir.Dir(filepath.Join(gi.AppPrefsDir(), "mail", hemail, mailbox))
	err := os.MkdirAll(string(dir), 0700)
	if err != nil {
		return err
	}
	err = dir.Init()
	if err != nil {
		return fmt.Errorf("initializing maildir: %w", err)
	}

	cachedFile := filepath.Join(gi.AppPrefsDir(), "caching", hemail, mailbox, "cached-messages.json")
	err = os.MkdirAll(filepath.Dir(cachedFile), 0700)
	if err != nil {
		return err
	}
	var cached []*CacheData
	err = jsons.Open(&cached, cachedFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, io.EOF) {
		return fmt.Errorf("opening cache list: %w", err)
	}

	mbox, err := c.Select(mailbox, nil).Wait()
	if err != nil {
		return fmt.Errorf("opening mailbox: %w", err)
	}
	_ = mbox

	// we want messages with UIDs not in the list we already cached
	criteria := &imap.SearchCriteria{}
	if len(cached) > 0 {
		seqset := imap.SeqSet{}
		for _, c := range cached {
			seqset.AddNum(c.UID)
		}

		nc := imap.SearchCriteria{}
		nc.UID = []imap.SeqSet{seqset}
		criteria.Not = append(criteria.Not, nc)
	}

	// these are the UIDs of the new messages
	uidsData, err := c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("searching for uids: %w", err)
	}

	uids := uidsData.AllNums()
	if len(uids) == 0 {
		return nil
	}

	// we only fetch the new messages
	fseqset := imap.SeqSet{}
	fseqset.AddNum(uids...)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		UID:      true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierHeader},
			{Specifier: imap.PartSpecifierText},
		},
	}

	mcmd := c.Fetch(fseqset, fetchOptions)

	for {
		msg := mcmd.Next()
		if msg == nil {
			break
		}

		mdata, err := msg.Collect()
		if err != nil {
			return err
		}

		key, w, err := dir.Create([]maildir.Flag{})
		if err != nil {
			return fmt.Errorf("making maildir file: %w", err)
		}

		var header, text []byte

		for k, v := range mdata.BodySection {
			if k.Specifier == imap.PartSpecifierHeader {
				header = v
			} else if k.Specifier == imap.PartSpecifierText {
				text = v
			}
		}

		_, err = w.Write(append(header, text...))
		if err != nil {
			return fmt.Errorf("writing message: %w", err)
		}

		err = w.(*os.File).Sync()
		if err != nil {
			return fmt.Errorf("saving message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("closing message: %w", err)
		}

		cd := &CacheData{
			Envelope: *mdata.Envelope,
			UID:      mdata.UID,
			Filename: key,
		}

		// we need to save the list of cached messages every time in case
		// we get interrupted or have an error
		cached = append(cached, cd)
		err = jsons.Save(&cached, cachedFile)
		if err != nil {
			return fmt.Errorf("saving cache list: %w", err)
		}
	}

	err = mcmd.Close()
	if err != nil {
		return fmt.Errorf("fetching messages: %w", err)
	}

	return nil
}
