package ast

import "solparsor/token"

type Node interface {
	Start() token.Position // First character of the node
	End() token.Position   // First character immediately after the node
}

type Expression interface {
	Node
	expressionNode()
}

type Statement interface {
	Node
	statementNode()
}

type Declaration interface {
	Node
	declarationNode()
}

type Comment struct {
	Slash token.Position // Position of the leading '/'
	Text  string
}

// In Solidity grammar it's called "SourceUnit" and represents the entire source file.
type File struct {
	Declarations []Declaration
}

func (f *File) Start() token.Position {
	if len(f.Declarations) > 0 {
		return f.Declarations[0].Start()
	}
	return 0
}

func (f *File) End() token.Position {
	if len(f.Declarations) > 0 {
		return f.Declarations[len(f.Declarations)-1].End()
	}
	return 0
}

/*~*~*~*~*~*~*~*~*~*~*~*~ Expressions *~*~*~*~*~*~*~*~*~*~*~*~*~*/

type Identifier struct {
	NamePos token.Position
	Name    string
}

// In Solidity grammar called "ElementaryTypeName".
// address, address payable, bool, string, uint, int, bytes, fixed, fixed-bytes and ufixed
type ElementaryType struct {
	ValuePos token.Position
	Kind     token.Token
	Value    string
}

// Start and End implementations for Expression type Nodes

func (x *Identifier) Start() token.Position     { return x.NamePos }
func (x *ElementaryType) Start() token.Position { return x.ValuePos }

func (x *Identifier) End() token.Position     { return token.Position(int(x.NamePos) + len(x.Name)) }
func (x *ElementaryType) End() token.Position { return token.Position(int(x.ValuePos) + len(x.Value)) }

// expressionNode() implementations to ensure that only expressions and types
// can be assigned to an Expression. This is useful if by mistake we try to use
// a Statement in a place where an Expression should be used instead.

func (*Identifier) expressionNode()     {}
func (*ElementaryType) expressionNode() {}

/*~*~*~*~*~*~*~*~*~*~*~*~* Statements *~*~*~*~*~*~*~*~*~*~*~*~*~*/

/*~*~*~*~*~*~*~*~*~*~*~*~ Declarations ~*~*~*~*~*~*~*~*~*~*~*~*~*/
