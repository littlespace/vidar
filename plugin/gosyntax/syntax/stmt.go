// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"go/ast"
	"go/token"
	"log"

	"github.com/nelsam/vidar/theme"
)

func (s *Syntax) addStmt(stmt ast.Stmt) {
	if stmt == nil {
		return
	}
	switch src := stmt.(type) {
	case *ast.BadStmt:
		s.addBadStmt(src)
	case *ast.AssignStmt:
		s.addAssignStmt(src)
	case *ast.CommClause:
		s.addCommClause(src)
	case *ast.SwitchStmt:
		s.addSwitchStmt(src)
	case *ast.TypeSwitchStmt:
		s.addTypeSwitchStmt(src)
	case *ast.CaseClause:
		s.addCaseStmt(src)
	case *ast.DeclStmt:
		s.addDecl(src.Decl)
	case *ast.SendStmt:
		s.addSendStmt(src)
	case *ast.SelectStmt:
		s.addSelectStmt(src)
	case *ast.ReturnStmt:
		s.addReturnStmt(src)
	case *ast.RangeStmt:
		s.addRangeStmt(src)
	case *ast.LabeledStmt:
		s.addLabeledStmt(src)
	case *ast.IncDecStmt:
		s.addIncDecStmt(src)
	case *ast.IfStmt:
		s.addIfStmt(src)
	case *ast.GoStmt:
		s.addGoStmt(src)
	case *ast.ForStmt:
		s.addForStmt(src)
	case *ast.ExprStmt:
		s.addExprStmt(src)
	case *ast.EmptyStmt:
		return
	case *ast.DeferStmt:
		s.addDeferStmt(src)
	case *ast.BranchStmt:
		s.addBranchStmt(src)
	case *ast.BlockStmt:
		s.addBlockStmt(src)
	default:
		log.Printf("Error: Unknown stmt type: %T", stmt)
	}
}

func (s *Syntax) addBadStmt(stmt *ast.BadStmt) {
	s.addNode(theme.Bad, stmt)
}

func (s *Syntax) addAssignStmt(stmt *ast.AssignStmt) {
	for _, expr := range stmt.Lhs {
		s.addExpr(expr)
	}
	for _, expr := range stmt.Rhs {
		s.addExpr(expr)
	}
}

func (s *Syntax) addSwitchStmt(stmt *ast.SwitchStmt) {
	s.add(theme.Keyword, stmt.Switch, len("switch"))
	s.addStmt(stmt.Init)
	s.addExpr(stmt.Tag)
	s.addBlockStmt(stmt.Body)
}

func (s *Syntax) addTypeSwitchStmt(stmt *ast.TypeSwitchStmt) {
	s.add(theme.Keyword, stmt.Switch, len("switch"))
	s.addStmt(stmt.Init)
	s.addStmt(stmt.Assign)
	s.addBlockStmt(stmt.Body)
}

func (s *Syntax) addCaseStmt(stmt *ast.CaseClause) {
	length := len("case")
	if stmt.List == nil {
		length = len("default")
	}
	s.add(theme.Keyword, stmt.Case, length)
	for _, expr := range stmt.List {
		s.addExpr(expr)
	}
	for _, stmt := range stmt.Body {
		s.addStmt(stmt)
	}
}

func (s *Syntax) addSendStmt(stmt *ast.SendStmt) {
	s.addExpr(stmt.Chan)
	s.addExpr(stmt.Value)
}

func (s *Syntax) addReturnStmt(stmt *ast.ReturnStmt) {
	s.add(theme.Keyword, stmt.Return, len("return"))
	for _, res := range stmt.Results {
		s.addExpr(res)
	}
}

func (s *Syntax) addRangeStmt(stmt *ast.RangeStmt) {
	s.add(theme.Keyword, stmt.For, len("for"))
	s.addExpr(stmt.Key)
	s.addExpr(stmt.Value)
	var rangeStart token.Pos
	switch stmt.Tok {
	case token.ILLEGAL:
		rangeStart = stmt.For + token.Pos(len("for "))
	case token.ASSIGN:
		rangeStart = stmt.TokPos + token.Pos(len("= "))
	case token.DEFINE:
		rangeStart = stmt.TokPos + token.Pos(len(":= "))
	}
	s.add(theme.Keyword, rangeStart, len("range"))
	s.addExpr(stmt.X)
	s.addBlockStmt(stmt.Body)
}

func (s *Syntax) addIfStmt(stmt *ast.IfStmt) {
	s.add(theme.Keyword, stmt.If, len("if"))
	s.addStmt(stmt.Init)
	s.addExpr(stmt.Cond)
	s.addBlockStmt(stmt.Body)
	if stmt.Else != nil {
		s.add(theme.Keyword, stmt.Body.End()+1, len("else"))
	}
	s.addStmt(stmt.Else)
}

func (s *Syntax) addForStmt(stmt *ast.ForStmt) {
	s.add(theme.Keyword, stmt.For, len("for"))
	s.addStmt(stmt.Init)
	s.addExpr(stmt.Cond)
	s.addStmt(stmt.Post)
	s.addBlockStmt(stmt.Body)
}

func (s *Syntax) addGoStmt(stmt *ast.GoStmt) {
	s.add(theme.Keyword, stmt.Go, len("go"))
	s.addCallExpr(stmt.Call)
}

func (s *Syntax) addLabeledStmt(stmt *ast.LabeledStmt) {
	// TODO: add stmt.Label
	s.addStmt(stmt.Stmt)
}

func (s *Syntax) addDeferStmt(stmt *ast.DeferStmt) {
	s.add(theme.Keyword, stmt.Defer, len("defer"))
	s.addCallExpr(stmt.Call)
}

func (s *Syntax) addIncDecStmt(stmt *ast.IncDecStmt) {
	s.addExpr(stmt.X)
}

func (s *Syntax) addExprStmt(stmt *ast.ExprStmt) {
	s.addExpr(stmt.X)
}

func (s *Syntax) addSelectStmt(stmt *ast.SelectStmt) {
	s.add(theme.Keyword, stmt.Select, len("select"))
	s.addBlockStmt(stmt.Body)
}

func (s *Syntax) addBranchStmt(stmt *ast.BranchStmt) {
	s.add(theme.Keyword, stmt.TokPos, len(stmt.Tok.String()))

	// TODO: add stmt.Label
}

func (s *Syntax) addBlockStmt(stmt *ast.BlockStmt) {
	defer s.rainbowScope(stmt.Lbrace, 1, stmt.Rbrace, 1)()
	for _, stmt := range stmt.List {
		s.addStmt(stmt)
	}
}
