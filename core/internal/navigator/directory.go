// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
)

type Watcher interface {
	Add(path string) error
	Remove(path string) error
}

type directory struct {
	// length is an atomically updated list of child nodes of
	// this directory.  Only access via atomics.
	//
	// Because it's accessed via atomics, it must be the first
	// field in the struct, for the sake of 32-bit systems.
	length int64

	mixins.LinearLayout

	driver  gxui.Driver
	button  *treeButton
	tree    *dirTree
	watcher Watcher
}

func newDirectory(projTree *ProjectTree, path string, watcher Watcher) *directory {
	driver := projTree.driver
	theme := projTree.theme

	button := newTreeButton(driver, theme, filepath.Base(path))
	tree := newDirTree(projTree, path)
	tree.SetMargin(math.Spacing{L: 10})
	d := &directory{
		driver:  driver,
		button:  button,
		tree:    tree,
		watcher: watcher,
	}
	d.Init(d, theme)
	d.AddChild(button)
	button.OnClick(func(gxui.MouseEvent) {
		if projTree.tocCtl != nil {
			projTree.layout.RemoveChild(projTree.tocCtl)
		}
		toc := NewTOC(projTree.cmdr, projTree.driver, projTree.theme, path)
		projTree.SetTOC(toc)
		scrollable := theme.CreateScrollLayout()
		// Disable horiz scrolling until we can figure out an accurate
		// way to calculate our width.
		scrollable.SetScrollAxis(false, true)
		scrollable.SetChild(toc)
		projTree.tocCtl = scrollable
		projTree.layout.AddChild(projTree.tocCtl)
		projTree.layout.SetChildWeight(projTree.tocCtl, 2)
		if d.Length() == 0 {
			return
		}
		if d.tree.Attached() {
			d.tree.Unload(watcher)
			d.button.Collapse()
			d.RemoveChild(d.tree)
			return
		}
		d.tree.Load(watcher)
		d.button.Expand()
		d.AddChild(d.tree)
	})
	d.reload()
	return d
}

func (d *directory) update(path string) {
	if !strings.HasPrefix(path, d.tree.path) {
		return
	}
	if d.tree.path == filepath.Dir(path) {
		d.driver.Call(d.reload)
		return
	}
	for _, dir := range d.tree.Dirs() {
		dir.update(path)
	}
}

func (d *directory) ExpandTo(dir string) {
	if !strings.HasPrefix(dir, d.tree.path) {
		return
	}
	if !d.button.Expanded() {
		d.button.Click(gxui.MouseEvent{})
	}
	for _, child := range d.tree.Dirs() {
		child.ExpandTo(dir)
	}
}

func (d *directory) Length() int64 {
	return atomic.LoadInt64(&d.length)
}

func (d *directory) updateExpandable(children int64) {
	if children == 0 {
		d.button.SetExpandable(false)
		return
	}
	d.button.SetExpandable(true)
}

func (d *directory) reload() {
	finfos, err := ioutil.ReadDir(d.tree.path)
	if err != nil {
		log.Printf("Unexpected error reading directory %s: %s", d.tree.path, err)
		return
	}
	defer d.driver.Call(func() {
		d.Relayout()
		d.Redraw()
	})

	children := int64(0)
	for _, finfo := range finfos {
		if finfo.IsDir() {
			children++
		}
	}

	d.updateExpandable(children)
	atomic.StoreInt64(&d.length, children)
	if d.tree.Attached() {
		d.tree.parse(finfos, d.watcher)
	}
}

type dirTree struct {
	mixins.LinearLayout

	projTree *ProjectTree
	driver   gxui.Driver
	theme    gxui.Theme
	path     string
}

func newDirTree(projTree *ProjectTree, path string) *dirTree {
	t := &dirTree{
		projTree: projTree,
		driver:   projTree.driver,
		theme:    projTree.theme,
		path:     path,
	}
	t.Init(t, projTree.theme)
	t.SetDirection(gxui.TopToBottom)
	return t
}

func (d *dirTree) Dirs() (dirs []*directory) {
	for _, c := range d.Children() {
		dirs = append(dirs, c.Control.(*directory))
	}
	return dirs
}

func (d *dirTree) Unload(w Watcher) error {
	for _, dir := range d.Dirs() {
		if !dir.tree.Attached() {
			continue
		}
		if err := dir.tree.Unload(w); err != nil {
			return err
		}
	}
	return w.Remove(d.path)
}

func (d *dirTree) Load(w Watcher) error {
	w.Add(d.path)
	finfos, err := ioutil.ReadDir(d.path)
	if err != nil {
		return err
	}
	d.parse(finfos, w)
	return nil
}

func (d *dirTree) parse(finfos []os.FileInfo, w Watcher) {
	d.RemoveAll()
	for _, finfo := range finfos {
		if !finfo.IsDir() {
			continue
		}
		if strings.HasPrefix(finfo.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(d.path, finfo.Name())
		dir := newDirectory(d.projTree, fullPath, w)
		d.AddChild(dir)
	}
}
