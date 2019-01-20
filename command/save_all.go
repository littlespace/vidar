// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package command

import (
	"fmt"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/bind"
)

type AllSaver interface {
	SaveAll()
}

type SaveAll struct{}

func NewSaveAll(theme gxui.Theme) *SaveAll {
	return &SaveAll{}
}

func (s *SaveAll) Name() string {
	return "save-all-files"
}

func (s *SaveAll) Menu() string {
	return "File"
}

func (s *SaveAll) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyS,
	}}
}

func (s *SaveAll) Exec(target interface{}) bind.Status {
	saver, ok := target.(AllSaver)
	if !ok {
		return bind.Waiting
	}
	saver.SaveAll()
	return bind.Done
}
