// Copyright (c) 2023, The Goki Authors. All rights reserved.
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
	b := gi.NewAppBody("grail").SetTitle("Grail")
	a := grail.NewApp(b, "app")
	a.Init()
	b.AddAppBar(a.AppBar)
	w := b.NewWindow().Run()
	grr.Log(a.GetMail())
	w.Wait()
}
