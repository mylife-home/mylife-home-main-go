package internal

import (
	"go/parser"
	"go/token"
	"path"

	annotation "github.com/YReshetko/go-annotation/pkg"
	"github.com/gookit/goutil/errorx/panics"
)

// Note very efficient, but only for errors + go-annotation does not provide access to this

func getNodeLocation(node annotation.Node) token.Position {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, path.Join(node.Meta().Dir(), node.Meta().FileName()), nil, parser.ParseComments)
	panics.IsTrue(err == nil)
	return fset.Position(node.ASTNode().Pos())
}
