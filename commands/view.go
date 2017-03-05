// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import "github.com/nelsam/vidar/commander"

type ViewHook struct{}

func (h ViewHook) Name() string {
	return "view-hook"
}

func (h ViewHook) CommandName() string {
	return "open-file"
}

func (h ViewHook) FileBindables(string) []commander.Bindable {
	return []commander.Bindable{
		NewHorizontalSplit(),
		NewVerticalSplit(),
		NewNextTab(),
		NewPrevTab(),
		NewFocusUp(),
		NewFocusDown(),
		NewFocusLeft(),
		NewFocusRight(),
	}
}