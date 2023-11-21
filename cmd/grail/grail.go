// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/gimain"
	"goki.dev/grail/grail"
	"goki.dev/grr"
)

func main() { gimain.Run(app) }

func app() {
	gi.SetAppName("grail")
	b := gi.NewBody().SetTitle("Grail")
	app := grail.NewApp(b, "app")
	b.AddTopBar(func(pw gi.Widget) {
		app.TopAppBar(b.TopAppBar(pw))
	})
	w := b.NewWindow().Run()
	grr.Log0(app.GetMail())
	w.Wait()
}
