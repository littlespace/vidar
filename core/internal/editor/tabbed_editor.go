// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/theme"
	"github.com/nelsam/vidar/ui"
)

type refocuser interface {
	ReFocus()
}

type TabbedEditor struct {
	layout ui.Layout

	editors map[string]*CodeEditor

	runner ui.Runner
	cur    string
}

func NewTabbedEditor(c ui.Creator, cmdr Commander, syntaxTheme theme.Theme) (*TabbedEditor, error) {
	l, err := c.TabbedLayout()
	if err != nil {
		return nil, fmt.Errorf("could not create tabbed layout: %s", err)
	}

	editor := &TabbedEditor{
		layout:  l,
		editors: make(map[string]input.Editor),
		runner:  c.Runner(),
	}
}

func (e *TabbedEditor) Has(name string) bool {
	_, ok := e.editors[name]
	return ok
}

func (e *TabbedEditor) Open(name, path string, environ []string) (editor input.Editor, existed bool) {
	if editor, ok := e.editors[name]; ok {
		e.Select(e.PanelIndex(editor.(gxui.Control)))
		gxui.SetFocus(editor.(gxui.Focusable))
		return editor, true
	}
	ce := &CodeEditor{}
	editor = ce
	// We want the OnRename trigger set up before the editor opens the file
	// in its Init method.
	ce.OnRename(func(newPath string) {
		e.driver.Call(func() {
			delete(e.editors, name)
			newName := relPath(hiddenPrefix, newPath)
			focused := e.SelectedPanel()
			e.editors[newName] = editor
			idx := e.PanelIndex(ce)
			if idx == -1 {
				return
			}
			e.RemovePanel(ce)
			e.AddPanelAt(ce, newName, idx)
			e.Select(e.PanelIndex(focused))
			gxui.SetFocus(focused.(gxui.Focusable))
		})
	})
	ce.Init(e.driver, e.theme, e.syntaxTheme, e.font, path, headerText)
	ce.SetTabWidth(4)
	e.Add(name, editor)
	return editor, false
}

func (e *TabbedEditor) Add(name string, editor input.Editor) {
	e.editors[name] = editor
	ec := editor.(gxui.Control)
	e.AddPanel(ec, name)
	e.Select(e.PanelIndex(ec))
	gxui.SetFocus(editor.(gxui.Focusable))
}

func (e *TabbedEditor) AddPanelAt(c gxui.Control, n string, i int) {
	e.PanelHolder.AddPanelAt(c, n, i)
	e.editors[n] = c.(input.Editor)
}

func (e *TabbedEditor) RemovePanel(panel gxui.Control) {
	toRemove := panel.(input.Editor)
	for name, editor := range e.editors {
		if editor == toRemove {
			delete(e.editors, name)
			break
		}
	}
	e.PanelHolder.RemovePanel(panel)
	if ed := e.CurrentEditor(); ed != nil {
		opener := e.cmdr.Bindable("focus-location").(Opener)
		e.cmdr.Execute(opener.For(focus.Path(ed.Filepath())))
	}
}

func (e *TabbedEditor) Files() []string {
	files := make([]string, 0, len(e.editors))
	for file := range e.editors {
		files = append(files, file)
	}
	return files
}

func (e *TabbedEditor) Editors() uint {
	return uint(len(e.editors))
}

func (e *TabbedEditor) CreatePanelTab() mixins.PanelTab {
	tab := basic.CreatePanelTab(e.theme)
	tab.OnMouseDown(func(ev gxui.MouseEvent) {
		if e.CurrentEditor() != nil {
			e.cur = e.CurrentEditor().Filepath()
		}
	})
	tab.OnMouseUp(func(gxui.MouseEvent) {
		if e.CurrentEditor() == nil {
			if len(e.editors) <= 1 {
				e.purgeSelf()
			} else {
				delete(e.editors, e.cur)
			}
		}
	})

	return tab
}

func (e *TabbedEditor) purgeSelf() {
	// Because of the order of events in gxui when a mouse drag happens,
	// the tab will move to a separate split *after* the SplitEditor's
	// MouseUp method is called, so the SplitEditor has no idea that
	// we're now empty.  We have to purge ourselves from the SplitEditor.
	parent := e.Parent()
	parent.(gxui.Container).RemoveChild(e)
	parent.(refocuser).ReFocus()
}

func (e *TabbedEditor) EditorAt(d ui.Direction) input.Editor {
	panels := e.PanelCount()
	if panels < 2 {
		return e.CurrentEditor()
	}
	idx := e.PanelIndex(e.SelectedPanel())
	switch d {
	case ui.Right:
		idx++
		if idx == panels {
			idx = 0
		}
	case ui.Left:
		idx--
		if idx < 0 {
			idx = panels - 1
		}
	}
	return e.Panel(idx).(input.Editor)
}

func (e *TabbedEditor) CloseCurrentEditor() (name string, editor input.Editor) {
	toRemove := e.CurrentEditor()
	if toRemove == nil {
		return "", nil
	}
	name = ""
	for key, panel := range e.editors {
		if panel == toRemove {
			name = key
			break
		}
	}
	e.RemovePanel(toRemove.(gxui.Control))
	if name == "" {
		return "", nil
	}
	return name, toRemove
}

func (e *TabbedEditor) SaveAll() {
	for name, editor := range e.editors {
		f, err := os.Create(name)
		if err != nil {
			log.Printf("Could not save %s : %s", name, err)
		}
		defer f.Close()
		if _, err := f.WriteString(editor.Text()); err != nil {
			log.Printf("Could not write to file %s: %s", name, err)
		}
	}
}

func (e *TabbedEditor) CurrentEditor() input.Editor {
	if e.SelectedPanel() == nil {
		return nil
	}
	return e.SelectedPanel().(input.Editor)
}

func (e *TabbedEditor) CurrentFile() string {
	if e.SelectedPanel() == nil {
		return ""
	}
	return e.SelectedPanel().(input.Editor).Filepath()
}

func (e *TabbedEditor) Elements() []interface{} {
	if e.SelectedPanel() == nil {
		return nil
	}
	return []interface{}{e.SelectedPanel()}
}

func relPath(from, path string) string {
	rel := strings.TrimPrefix(path, from)
	if rel[0] == filepath.Separator {
		rel = rel[1:]
	}
	return rel
}
