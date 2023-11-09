// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grail implements a GUI email client.
package grail

import (
	"goki.dev/gi/v2/gi"
)

// App is an email client app.
type App struct {
	gi.Frame
}

func (app *App) ConfigWidget(sc *gi.Scene) {
	if app.HasChildren() {
		return
	}
	updt := app.UpdateStart()

	sp := gi.NewSplits(app).SetSplits(0.3, 0.7)
	gi.NewFrame(sp, "list")

	mail := gi.NewFrame(sp, "mail")
	gi.NewLabel(mail).SetText("Message goes here")

	app.UpdateEndLayout(updt)
}