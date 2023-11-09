// Code generated by "goki generate"; DO NOT EDIT.

package grail

import (
	"goki.dev/gi/v2/gi"
	"goki.dev/gti"
	"goki.dev/ki/v2"
	"goki.dev/ordmap"
)

// AppType is the [gti.Type] for [App]
var AppType = gti.AddType(&gti.Type{
	Name:       "goki.dev/grail/grail.App",
	ShortName:  "grail.App",
	IDName:     "app",
	Doc:        "App is an email client app.",
	Directives: gti.Directives{},
	Fields:     ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Embeds: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Frame", &gti.Field{Name: "Frame", Type: "goki.dev/gi/v2/gi.Frame", LocalType: "gi.Frame", Doc: "", Directives: gti.Directives{}, Tag: ""}},
	}),
	Methods:  ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
	Instance: &App{},
})

// NewApp adds a new [App] with the given name
// to the given parent. If the name is unspecified, it defaults
// to the ID (kebab-case) name of the type, plus the
// [ki.Ki.NumLifetimeChildren] of the given parent.
func NewApp(par ki.Ki, name ...string) *App {
	return par.NewChild(AppType, name...).(*App)
}

// KiType returns the [*gti.Type] of [App]
func (t *App) KiType() *gti.Type {
	return AppType
}

// New returns a new [*App] value
func (t *App) New() ki.Ki {
	return &App{}
}

// SetTooltip sets the [App.Tooltip]
func (t *App) SetTooltip(v string) *App {
	t.Tooltip = v
	return t
}

// SetClass sets the [App.Class]
func (t *App) SetClass(v string) *App {
	t.Class = v
	return t
}

// SetCustomContextMenu sets the [App.CustomContextMenu]
func (t *App) SetCustomContextMenu(v func(m *gi.Scene)) *App {
	t.CustomContextMenu = v
	return t
}

// SetLayout sets the [App.Lay]
func (t *App) SetLayout(v gi.Layouts) *App {
	t.Lay = v
	return t
}

// SetStackTop sets the [App.StackTop]
func (t *App) SetStackTop(v int) *App {
	t.StackTop = v
	return t
}

// SetStripes sets the [App.Stripes]
func (t *App) SetStripes(v gi.Stripes) *App {
	t.Stripes = v
	return t
}
