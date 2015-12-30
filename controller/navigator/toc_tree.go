package navigator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
)

var (
	genericColor = gxui.Color{
		R: 0.6,
		G: 0.8,
		B: 1,
		A: 1,
	}
	nameColor = gxui.Color{
		R: 0.6,
		G: 1,
		B: 0.5,
		A: 1,
	}
	skippableColor = gxui.Color{
		R: 0.9,
		G: 0.6,
		B: 0.8,
		A: 1,
	}

	// Since the const values aren't exported by go/build, I've just copied them
	// from https://github.com/golang/go/blob/master/src/go/build/syslist.go
	gooses   = []string{"android", "darwin", "dragonfly", "freebsd", "linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "windows"}
	goarches = []string{"386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "s390", "s390x", "sparc", "sparc64"}
)

type genericTreeNode struct {
	gxui.AdapterBase

	name     string
	path     string
	color    gxui.Color
	children []gxui.TreeNode
}

func (t genericTreeNode) Count() int {
	return len(t.children)
}

func (t genericTreeNode) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	for i, child := range t.children {
		childPath := child.Item().(string)
		if path == childPath || strings.HasPrefix(path, childPath+".") {
			return i
		}
	}
	return -1
}

func (t genericTreeNode) Size(theme gxui.Theme) math.Size {
	return math.Size{
		W: 20 * theme.DefaultMonospaceFont().GlyphMaxSize().W,
		H: theme.DefaultMonospaceFont().GlyphMaxSize().H,
	}
}

func (t genericTreeNode) Item() gxui.AdapterItem {
	return t.path
}

func (t genericTreeNode) NodeAt(index int) gxui.TreeNode {
	return t.children[index]
}

func (t genericTreeNode) Create(theme gxui.Theme) gxui.Control {
	label := theme.CreateLabel()
	label.SetText(t.name)
	label.SetColor(t.color)
	return label
}

// skippingTreeNode is a gxui.TreeNode that will skip to the child
// node if there is exactly one child node on implementations of all
// gxui.TreeNode methods.  This means that any number of nested
// skippingTreeNodes will display as a single node, so long as no
// child contains more than one child.
type skippingTreeNode struct {
	genericTreeNode
}

func (t skippingTreeNode) Count() int {
	if len(t.children) == 1 {
		return t.children[0].Count()
	}
	return t.genericTreeNode.Count()
}

func (t skippingTreeNode) ItemIndex(item gxui.AdapterItem) int {
	if len(t.children) == 1 {
		return t.children[0].ItemIndex(item)
	}
	return t.genericTreeNode.ItemIndex(item)
}

func (t skippingTreeNode) Item() gxui.AdapterItem {
	if len(t.children) == 1 {
		return t.children[0].Item()
	}
	return t.genericTreeNode.Item()
}

func (t skippingTreeNode) NodeAt(index int) gxui.TreeNode {
	if len(t.children) == 1 {
		return t.children[0].NodeAt(index)
	}
	return t.genericTreeNode.NodeAt(index)
}

func (t skippingTreeNode) Create(theme gxui.Theme) gxui.Control {
	if len(t.children) == 1 {
		return t.children[0].Create(theme)
	}
	return t.genericTreeNode.Create(theme)
}

type Location struct {
	filename string
	pos      int
}

func (l Location) File() string {
	return l.filename
}

func (l Location) Pos() int {
	return l.pos
}

type Name struct {
	genericTreeNode
	Location
}

type TOC struct {
	skippingTreeNode

	path string
}

func NewTOC(path string) *TOC {
	toc := &TOC{}
	toc.Init(path)
	return toc
}

func (t *TOC) Init(path string) {
	t.path = path
	t.Reload()
}

func (t *TOC) Reload() {
	pkgs, err := parser.ParseDir(token.NewFileSet(), t.path, nil, 0)
	if err != nil {
		log.Printf("Error parsing dir: %s", err)
	}
	for _, pkg := range pkgs {
		t.children = append(t.children, t.parsePkg(pkg))
	}
}

func (t *TOC) parsePkg(pkg *ast.Package) genericTreeNode {
	var (
		pkgNode = genericTreeNode{name: pkg.Name, path: pkg.Name, color: skippableColor}

		consts  []gxui.TreeNode
		vars    []gxui.TreeNode
		typeMap = make(map[string]*Name)
		types   []gxui.TreeNode
		funcs   []gxui.TreeNode
	)
	for filename, f := range pkg.Files {
		for _, decl := range f.Decls {
			switch src := decl.(type) {
			case *ast.GenDecl:
				switch src.Tok.String() {
				case "const":
					consts = append(consts, valueNamesFrom(filename, pkg.Name+".constants", src.Specs)...)
				case "var":
					vars = append(vars, valueNamesFrom(filename, pkg.Name+".global vars", src.Specs)...)
				case "type":
					// I have yet to see a case where a type declaration has more than one Specs.
					typeSpec := src.Specs[0].(*ast.TypeSpec)
					typeName := typeSpec.Name.String()

					// We can't guarantee that the type declaration was found before method
					// declarations, so the value may already exist in the map.
					typ, ok := typeMap[typeName]
					if !ok {
						typ = &Name{}
						typeMap[typeName] = typ
					}
					typ.name = typeName
					typ.path = pkg.Name + ".types." + typ.name
					typ.color = nameColor
					typ.filename = filename
					typ.pos = int(typeSpec.Pos())
					types = append(types, typ)
				}
			case *ast.FuncDecl:
				var name Name
				name.name = src.Name.String()
				name.path = pkg.Name + ".funcs." + name.name
				name.color = nameColor
				name.filename = filename
				name.pos = int(src.Pos())
				if src.Recv == nil {
					funcs = append(funcs, name)
					continue
				}
				recvTyp := src.Recv.List[0].Type
				if starExpr, ok := recvTyp.(*ast.StarExpr); ok {
					recvTyp = starExpr.X
				}
				recvTypeName := recvTyp.(*ast.Ident).String()
				typ, ok := typeMap[recvTypeName]
				if !ok {
					typ = &Name{}
					typeMap[recvTypeName] = typ
				}
				name.path = pkg.Name + ".types." + recvTypeName + "." + name.name
				typ.children = append(typ.children, name)
			}
		}
	}
	pkgNode.children = []gxui.TreeNode{
		genericTreeNode{name: "constants", path: pkg.Name + ".constants", color: genericColor, children: consts},
		genericTreeNode{name: "global vars", path: pkg.Name + ".global vars", color: genericColor, children: vars},
		genericTreeNode{name: "types", path: pkg.Name + ".types", color: genericColor, children: types},
		genericTreeNode{name: "funcs", path: pkg.Name + ".funcs", color: genericColor, children: funcs},
	}
	return pkgNode
}

func valueNamesFrom(filename, parentName string, specs []ast.Spec) (names []gxui.TreeNode) {
	for _, spec := range specs {
		valSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valSpec.Names {
			var newName Name
			newName.name = name.String()
			newName.path = parentName + "." + newName.name
			newName.color = nameColor
			newName.filename = filename
			newName.pos = int(name.Pos())
			names = append(names, newName)
		}
	}
	return
}
