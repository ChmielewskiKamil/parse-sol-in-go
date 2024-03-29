package main

import (
	"fmt"
	"solparsor/lexer"
	"solparsor/token"
)

func main() {
	input := `
    Contract Vault {
        uint256 x;
        x = 5;
    }
    `
	lexer := lexer.Lex(input)

	for {
		tkn := lexer.NextToken()
		fmt.Printf("Token: %s, at position: %d, with type: %s\n", tkn.String(), tkn.Pos, token.Tokens[tkn.Type])

		if tkn.Type == token.EOF || tkn.Type == token.ILLEGAL {
			break
		}
	}
}
