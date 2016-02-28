// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package navigator

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commands"
	"github.com/nelsam/vidar/controller"
	"github.com/nelsam/vidar/settings"
)

type Projects struct {
	theme gxui.Theme

	button          gxui.Button
	projects        gxui.List
	projectsAdapter *gxui.DefaultAdapter

	projectFrame gxui.Control
}

func NewProjectsPane(driver gxui.Driver, theme gxui.Theme, projFrame gxui.Control) *Projects {
	pane := &Projects{
		theme:           theme,
		projectFrame:    projFrame,
		button:          createIconButton(driver, theme, "projects.png"),
		projects:        theme.CreateList(),
		projectsAdapter: gxui.CreateDefaultAdapter(),
	}
	pane.projectsAdapter.SetItems(settings.Projects())
	pane.projects.SetAdapter(pane.projectsAdapter)
	return pane
}

func (p *Projects) Add(project settings.Project) {
	projects := append(p.projectsAdapter.Items().([]settings.Project), project)
	p.projectsAdapter.SetItems(projects)
}

func (p *Projects) Button() gxui.Button {
	return p.button
}

func (p *Projects) Frame() gxui.Control {
	return p.projects
}

func (p *Projects) Projects() []settings.Project {
	return p.projectsAdapter.Items().([]settings.Project)
}

func (p *Projects) OnComplete(onComplete func(controller.Executor)) {
	opener := commands.NewProjectOpener(p.theme, p.projectFrame)
	p.projects.OnSelectionChanged(func(selected gxui.AdapterItem) {
		opener.SetProject(selected.(settings.Project))
		onComplete(opener)
	})
}