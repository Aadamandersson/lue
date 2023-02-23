package parser

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/ast"
	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/lexer"
	"github.com/aadamandersson/lue/internal/span"
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

	sp := span.NewEmpty(0)
	if len(exprs) > 1 {
		sp = exprs[0].Span().To(exprs[len(exprs)-1].Span())
	}
	return &ast.BlockExpr{Exprs: exprs, Sp: sp}
}

func (p *parser) parseExpr() ast.Expr {
	if sp, ok := p.eat(token.Let); ok {
		return p.parseLetExpr(sp)
	}
	return p.parsePrecExpr(0)
}

// parseLetExpr parses a let binding, `let` token already eaten.
// `let ident = init`
func (p *parser) parseLetExpr(let_sp span.Span) ast.Expr {
	ident := p.parseIdent()
	if ident == nil {
		p.error("expected identifier in let binding, but got `%s`", p.tok.Kind)
	}

	if _, ok := p.eat(token.Eq); !ok {
		p.error("expected `=`, but got `%s`", p.tok.Kind)
	}

	init := p.parseExpr()
	if init == nil {
		p.error("expected expression, but got `%s`", p.tok.Kind)
	}

	if ident == nil || init == nil {
		return &ast.ErrExpr{}
	}

	sp := let_sp.To(init.Span())
	return &ast.LetExpr{Ident: ident, Init: init, Sp: sp}
}

func (p *parser) parsePrecExpr(min_prec int) ast.Expr {
	expr := p.parseBotExpr()

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
		sp := expr.Span().To(rhs.Span())
		switch op.Kind {
		case ast.Assign:
			expr = &ast.AssignExpr{X: expr, Y: rhs, Sp: sp}
		default:
			expr = &ast.BinaryExpr{X: expr, Op: op, Y: rhs, Sp: sp}
		}
	}

	return expr
}

func (p *parser) parseBotExpr() ast.Expr {
	if sp, ok := p.eat(token.If); ok {
		return p.parseIfExpr(sp)
	}

	if sp, ok := p.eat(token.Ident); ok {
		return &ast.Ident{Name: p.prev_tok.Lit, Sp: sp}
	}

	if sp, ok := p.eat(token.Number); ok {
		return &ast.IntegerLiteral{V: p.prev_tok.Lit, Sp: sp}
	}

	if sp, ok := p.eat(token.False); ok {
		return &ast.BooleanLiteral{V: false, Sp: sp}
	}

	if sp, ok := p.eat(token.True); ok {
		return &ast.BooleanLiteral{V: true, Sp: sp}
	}

	return p.parseIdent()
}

// parseIfExpr parses `if cond { exprs } [else [if cond] { exprs ]`
// `if` token already eaten.
func (p *parser) parseIfExpr(if_sp span.Span) ast.Expr {
	cond := p.parseExpr()
	if cond == nil {
		p.error("expected condition")
	}

	then := p.parseBlockExpr()
	sp := if_sp.To(then.Span())
	var els ast.Expr
	if _, ok := p.eat(token.Else); ok {
		els = p.parseElseExpr()
		sp = if_sp.To(els.Span())
	}
	return &ast.IfExpr{Cond: cond, Then: then, Else: els, Sp: sp}
}

// parseElseExpr parses `else [if cond] { exprs }`.
// `else` token already eaten.
func (p *parser) parseElseExpr() ast.Expr {
	if if_sp, ok := p.eat(token.If); ok {
		return p.parseIfExpr(if_sp)
	}
	return p.parseBlockExpr()
}

// parseBlockExpr parses `{ exprs }`
func (p *parser) parseBlockExpr() ast.Expr {
	open_sp, ok := p.eat(token.LBrace)
	if !ok {
		p.error("expected opening delimiter `%s`", token.LBrace)
		return &ast.ErrExpr{}
	}

	var exprs []ast.Expr
	for !p.tok.IsOneOf(token.RBrace, token.Eof) {
		exprs = append(exprs, p.parseExpr())
	}

	close_sp, ok := p.eat(token.RBrace)
	if !ok {
		p.error("expected closing delimiter `%s`", token.LBrace)
		return &ast.ErrExpr{}
	}

	sp := open_sp.To(close_sp)
	return &ast.BlockExpr{Exprs: exprs, Sp: sp}
}

func (p *parser) parseIdent() *ast.Ident {
	if sp, ok := p.eat(token.Ident); ok {
		return &ast.Ident{Name: p.prev_tok.Lit, Sp: sp}
	}
	return nil
}

// eat advances the parser to the next token and returns true, if the current token kind
// is kind k, otherwise returns false.
func (p *parser) eat(k token.Kind) (span.Span, bool) {
	sp := p.tok.Sp
	if p.tok.Is(k) {
		p.next()
		return sp, true
	}
	return sp, false
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
