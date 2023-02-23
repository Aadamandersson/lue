package parser

import (
	"fmt"

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
	var exprs []ast.Expr
	for !p.tok.Is(token.Eof) {
		exprs = append(exprs, p.parseExpr())
	}
	return &ast.BlockExpr{Exprs: exprs}
}

func (p *parser) parseExpr() ast.Expr {
	if ok := p.eat(token.Let); ok {
		return p.parseLetExpr()
	}
	return p.parsePrecExpr(0)
}

// parseLetExpr parses a let binding, `let` token already eaten.
// `let ident = init`
func (p *parser) parseLetExpr() ast.Expr {
	ident := p.parseIdent()
	if ident == nil {
		p.error("expected identifier in let binding, but got `%s`", p.tok.Kind)
	}

	if ok := p.eat(token.Eq); !ok {
		p.error("expected `=`, but got `%s`", p.tok.Kind)
	}

	init := p.parseExpr()
	if init == nil {
		p.error("expected expression, but got `%s`", p.tok.Kind)
	}

	if ident == nil || init == nil {
		return &ast.ErrExpr{}
	}

	return &ast.LetExpr{Ident: ident, Init: init}
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
			expr = &ast.AssignExpr{X: expr, Y: rhs}
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

	return p.parseIdent()
}

func (p *parser) parseIdent() *ast.Ident {
	if ok := p.eat(token.Ident); ok {
		return &ast.Ident{Name: p.prev_tok.Lit, Sp: p.prev_tok.Sp}
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

func (p *parser) error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	diagnostic.NewBuilder(msg, p.tok.Sp).WithLabel("here").Emit(p.diags)
}
