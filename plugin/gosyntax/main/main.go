// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/plugin/command"
	"github.com/nelsam/vidar/plugin/gosyntax"
)

type GolangHook struct {
}

func (h GolangHook) Name() string {
	return "golang-hook"
}

func (h GolangHook) OpName() string {
	return "focus-location"
}

func (h GolangHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}
	return []bind.Bindable{
		gosyntax.New(),
	}
}

// Bindables is the main entry point to the command.
func Bindables(cmdr command.Commander, driver gxui.Driver, theme gxui.Theme) []bind.Bindable {
	return []bind.Bindable{
		GolangHook{},
	}
}
