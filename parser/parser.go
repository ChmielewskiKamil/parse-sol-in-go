package parser

import (
	"fmt"
	"solparsor/ast"
	"solparsor/lexer"
	"solparsor/token"
)

type Parser struct {
	l      lexer.Lexer
	errors ErrorList

	currTkn token.Token
	peekTkn token.Token
}

func (p *Parser) init(src string) {
	p.l = *lexer.Lex(src)
	p.errors = ErrorList{}

	// Read two tokens, so currTkn and peekTkn are both set
	p.nextToken()
	p.nextToken()
}

func (p *Parser) nextToken() {
	p.currTkn = p.peekTkn
	p.peekTkn = p.l.NextToken()
}

func (p *Parser) ParseFile() *ast.File {
	file := &ast.File{}
	file.Declarations = []ast.Declaration{}

	for p.currTkn.Type != token.EOF {
		decl := p.parseDeclaration()
		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
		p.nextToken()
	}

	return file
}

func (p *Parser) parseDeclaration() ast.Declaration {
	switch p.currTkn.Type {
	case token.ADDRESS, token.UINT_256, token.BOOL:
		return p.parseVariableDeclaration()
	case token.FUNCTION:
		return p.parseFunctionDeclaration()
	default:
		return nil
	}
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	decl := &ast.FunctionDeclaration{}

	// 1. Function keyword
	fnType := &ast.FunctionType{}
	fnType.Func = p.currTkn.Pos

	// 2. Function identifier
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	decl.Name = &ast.Identifier{
		NamePos: p.currTkn.Pos,
		Name:    p.currTkn.Literal,
	}

	// 3. ( Param List )
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	params := &ast.ParamList{}
	params.Opening = p.currTkn.Pos
	fnType.Params = params

	p.nextToken()

	for !p.currTknIs(token.RPAREN) {
		if !p.expectPeek(token.IDENTIFIER) {
			return nil
		}

		// @TODO: We skip the type for now since it is an expression.
		param := &ast.Param{
			Name: &ast.Identifier{
				NamePos: p.currTkn.Pos,
				Name:    p.currTkn.Literal,
			},
		}

		fnType.Params.List = append(fnType.Params.List, param)
		p.nextToken()
	}

	fnType.Params.Closing = p.currTkn.Pos

	// 4. Visibility, State Mutability, Modifier Invocation, Override, Virtual

	// 5. Returns ( Param List )

	// 6. Body block

	// 7. Semicolon
	for !p.currTknIs(token.SEMICOLON) {
		p.nextToken()
	}

	decl.Type = fnType
	return decl
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	decl := &ast.VariableDeclaration{}

	// We are sitting on the variable type e.g. address or uint256
	decl.Type = &ast.ElementaryType{
		ValuePos: p.currTkn.Pos,
		Kind:     p.currTkn,
		Value:    p.currTkn.Literal,
	}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	decl.Name = &ast.Identifier{
		NamePos: p.currTkn.Pos,
		Name:    p.currTkn.Literal,
	}

	// @TODO: We skip the Value for now since it is an expression.

	// The variable declaration ends with a semicolon.
	for !p.currTknIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

// expectPeek checks if the next token is of the expected type.
// If it is it advances the tokens.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTkn.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be: %s, got: %s instead (at offset: %d)",
		t.String(), p.peekTkn.Type.String(), p.peekTkn.Pos)
	p.errors.Add(p.peekTkn.Pos, msg)
}

// currTknIs checks if the current token is of the expected type.
func (p *Parser) currTknIs(t token.TokenType) bool {
	return p.currTkn.Type == t
}
