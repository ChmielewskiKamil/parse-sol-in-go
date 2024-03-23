package lexer

import (
	"solparsor/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `
    Contract Vault {
        uint256 x;
        x = 5;
        address owner = 0xDEADBEEF;
        mapping(address => uint256) balances;
        function deposit(uint256 amount) public {
            balances[msg.sender] += amount;
        }
    }

    Library SafeMath {
        i != 0;
        i++;
        i--;
    }

    uint256 y;
    address attacker1337;

    address constant UniswapV3Factory = 0x1F98431c8aD98523631AE4a59f267346ea31F984;
    `
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		// Vault contract start
		{token.CONTRACT, "Contract"},
		{token.IDENTIFIER, "Vault"},
		{token.LBRACE, "{"},
		{token.UINT_256, "uint256"},
		{token.IDENTIFIER, "x"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "x"},
		{token.ASSIGN, "="},
		{token.DECIMAL_NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.ADDRESS, "address"},
		{token.IDENTIFIER, "owner"},
		{token.ASSIGN, "="},
		{token.HEX_NUMBER, "0xDEADBEEF"},
		{token.SEMICOLON, ";"},
		{token.MAPPING, "mapping"},
		{token.LPAREN, "("},
		{token.ADDRESS, "address"},
		{token.DOUBLE_ARROW, "=>"},
		{token.UINT_256, "uint256"},
		{token.RPAREN, ")"},
		{token.IDENTIFIER, "balances"},
		{token.SEMICOLON, ";"},
		{token.FUNCTION, "function"},
		{token.IDENTIFIER, "deposit"},
		{token.LPAREN, "("},
		{token.UINT_256, "uint256"},
		{token.IDENTIFIER, "amount"},
		{token.RPAREN, ")"},
		{token.PUBLIC, "public"},
		{token.LBRACE, "{"},
		{token.IDENTIFIER, "balances"},
		{token.LBRACKET, "["},
		{token.IDENTIFIER, "msg"},
		{token.PERIOD, "."},
		{token.IDENTIFIER, "sender"},
		{token.RBRACKET, "]"},
		{token.ASSIGN_ADD, "+="},
		{token.IDENTIFIER, "amount"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.RBRACE, "}"},
		// Vault contract end

		// SafeMath library start
		{token.LIBRARY, "Library"},
		{token.IDENTIFIER, "SafeMath"},
		{token.LBRACE, "{"},
		{token.IDENTIFIER, "i"},
		{token.NOT_EQUAL, "!="},
		{token.DECIMAL_NUMBER, "0"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "i"},
		{token.INC, "++"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "i"},
		{token.DEC, "--"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		// SafeMath library end

		// Variables outside of contract
		{token.UINT_256, "uint256"},
		{token.IDENTIFIER, "y"},
		{token.SEMICOLON, ";"},
		{token.ADDRESS, "address"},
		{token.IDENTIFIER, "attacker1337"},
		{token.SEMICOLON, ";"},
		{token.ADDRESS, "address"},
		{token.CONSTANT, "constant"},
		{token.IDENTIFIER, "UniswapV3Factory"},
		{token.ASSIGN, "="},
		{token.HEX_NUMBER, "0x1F98431c8aD98523631AE4a59f267346ea31F984"},
		{token.SEMICOLON, ";"},

		{token.EOF, ""},
	}

	lexer := Lex(input)

	for i, tt := range tests {
		tkn := lexer.NextToken()

		if tkn.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected: %s (%d), got: %s",
				i, token.Tokens[tt.expectedType], tt.expectedType, token.Tokens[tkn.Type])
		}

		if tkn.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected: %s, got: %s",
				i, tt.expectedLiteral, tkn.Literal)
		}
	}
}
