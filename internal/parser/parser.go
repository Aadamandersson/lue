package parser

import (
	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/lexer"
	"github.com/aadamandersson/lue/internal/token"
)

func Parse(src []byte, diags *diagnostic.Bag) ast.Expr {
	tokens := lexer.Lex(src)
	p := new(diags, tokens)
	return p.parse()
}

type parser struct {
	diags    *diagnostic.Bag
	tokens   []token.Token
	tok      token.Token
	prev_tok token.Token
	pos      int
}

func new(diags *diagnostic.Bag, tokens []token.Token) parser {
	p := parser{diags: diags, tokens: tokens}
	p.next()
	return p
}

func (p *parser) parse() ast.Expr {
	return p.parseExpr()
}

func (p *parser) parseExpr() ast.Expr {
	return p.parsePrecExpr(0)
}

func (p *parser) parsePrecExpr(min_prec int) ast.Expr {
	expr := p.parsePrimaryExpr()

	for {
		op, ok := ast.BinOpFromToken(p.tok)
		if !ok || op.Prec() < min_prec {
			break
		}
		p.next()

		prec := op.Prec()
		if op.Assoc() == ast.AssocLeft {
			prec = prec + 1
		}

		rhs := p.parsePrecExpr(prec)
		switch op.Kind {
		case ast.Assign:
			ident := expr.(*ast.Ident)
			expr = &ast.AssignExpr{Ident: ident, Init: rhs}
		default:
			expr = &ast.BinaryExpr{X: expr, Op: op, Y: rhs}
		}
	}

	return expr
}

func (p *parser) parsePrimaryExpr() ast.Expr {
	if ok := p.eat(token.Ident); ok {
		return &ast.Ident{Name: p.prev_tok.Lit, Sp: p.prev_tok.Sp}
	}

	if ok := p.eat(token.Number); ok {
		return &ast.IntegerLiteral{V: p.prev_tok.Lit, Sp: p.prev_tok.Sp}
	}

	if ok := p.eat(token.False); ok {
		return &ast.BooleanLiteral{V: false, Sp: p.prev_tok.Sp}
	}

	if ok := p.eat(token.True); ok {
		return &ast.BooleanLiteral{V: true, Sp: p.prev_tok.Sp}
	}

	return nil
}

// eat advances the parser to the next token and returns true, if the current token kind
// is kind k, otherwise returns false.
func (p *parser) eat(k token.Kind) bool {
	if p.tok.Is(k) {
		p.next()
		return true
	}
	return false
}

// next advances the parser to the next token in tokens.
func (p *parser) next() {
	if p.pos < len(p.tokens) {
		p.prev_tok = p.tok
		p.tok = p.tokens[p.pos]
		p.pos += 1
	}
}
