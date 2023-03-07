package parser

import (
	"fmt"

	"github.com/aadamandersson/lue/internal/diagnostic"
	"github.com/aadamandersson/lue/internal/ir/ast"
	"github.com/aadamandersson/lue/internal/lexer"
	"github.com/aadamandersson/lue/internal/session"
	"github.com/aadamandersson/lue/internal/span"
	"github.com/aadamandersson/lue/internal/token"
)

func Parse(sess *session.Session) []ast.Item {
	tokens := lexer.Lex(sess)
	p := new(sess, tokens)
	return p.parse()
}

type parser struct {
	sess    *session.Session
	tokens  []token.Token
	tok     token.Token
	prevTok token.Token
	pos     int
}

func new(sess *session.Session, tokens []token.Token) parser {
	p := parser{sess: sess, tokens: tokens}
	p.next()
	return p
}

func (p *parser) parse() []ast.Item {
	var items []ast.Item
	for !p.tok.Is(token.Eof) {
		if item := p.parseItem(); item != nil {
			items = append(items, item)
		} else {
			p.error("expected item")
			p.next()
		}
	}

	return items
}

func (p *parser) parseItem() ast.Item {
	if fnSp, ok := p.eat(token.Fn); ok {
		return p.parseFnDecl(fnSp)
	}
	if classSp, ok := p.eat(token.Class); ok {
		return p.parseClassDecl(classSp)
	}
	return nil
}

// parseFnDecl parses a function declaration, `fn` token already eaten.
// `fn ident([params]) [:ty] { exprs }`
func (p *parser) parseFnDecl(fnSp span.Span) ast.Item {
	ident := p.parseIdent()
	if ident == nil {
		p.error("expected function name, but got `%s`", p.tok.Kind)
	}

	params := p.parseParams()
	if params == nil {
		return nil
	}

	var ty *ast.Ty
	if sp, ok := p.eat(token.Colon); ok {
		ty = p.parseTy()
	} else {
		ty = &ast.Ty{Kind: ast.TyUnit, Sp: sp}
	}

	if ty == nil {
		return &ast.ErrItem{}
	}

	body := p.parseBlockExpr()
	sp := fnSp.To(body.Span())
	return &ast.FnDecl{Ident: ident, In: params, Out: ty, Body: body, Sp: sp}
}

// parseClassDecl parses `class ident { fields }`.
// `class` token already eaten.
func (p *parser) parseClassDecl(classSp span.Span) ast.Item {
	ident := p.parseIdent()
	if ident == nil {
		p.error("expected class name, but got `%s`", p.tok.Kind)
		return &ast.ErrItem{}
	}

	if _, ok := p.eat(token.LBrace); !ok {
		p.error("expected opening delimiter `%s`", token.LBrace)
		return &ast.ErrItem{}
	}

	var fields []*ast.VarDecl
	for !p.tok.IsOneOf(token.RBrace, token.Eof) {
		ident := p.parseIdent()
		if ident == nil {
			p.error("expected field name, but got `%s`", p.tok.Kind)
			continue
		}

		if _, ok := p.eat(token.Colon); !ok {
			p.error("expected `:`")
		}

		ty := p.parseTy()
		if ty == nil {
			continue
		}

		field := &ast.VarDecl{Ident: ident, Ty: ty}
		fields = append(fields, field)

		if _, ok := p.eat(token.Comma); !ok {
			break
		}
	}

	closeSp, ok := p.eat(token.RBrace)
	if !ok {
		p.error("expected closing delimiter `%s`", token.LBrace)
		return &ast.ErrItem{}
	}

	sp := classSp.To(closeSp)
	return &ast.ClassDecl{Ident: ident, Fields: fields, Sp: sp}
}

func (p *parser) parseParams() []*ast.VarDecl {
	params := make([]*ast.VarDecl, 0)

	if _, ok := p.eat(token.LParen); !ok {
		p.error("expected opening delimiter `%s`", token.LParen)
		return nil
	}

	for !p.tok.IsOneOf(token.RParen, token.Eof) {
		ident := p.parseIdent()
		if ident == nil {
			p.error("expected parameter name, but got `%s`", p.tok.Kind)
			continue
		}

		if _, ok := p.eat(token.Colon); !ok {
			p.error("expected `:`")
		}

		ty := p.parseTy()
		if ty == nil {
			continue
		}

		param := &ast.VarDecl{Ident: ident, Ty: ty}
		params = append(params, param)

		if _, ok := p.eat(token.Comma); !ok {
			break
		}
	}

	if _, ok := p.eat(token.RParen); !ok {
		p.error("expected closing delimiter `%s`", token.RParen)
		return nil
	}

	return params
}

func (p *parser) parseExpr() ast.Expr {
	if sp, ok := p.eat(token.Let); ok {
		return p.parseLetExpr(sp)
	}

	if sp, ok := p.eat(token.Return); ok {
		return p.parseReturnExpr(sp)
	}

	if sp, ok := p.eat(token.Break); ok {
		return p.parseBreakExpr(sp)
	}

	return p.parsePrecExpr(0)
}

// parseLetExpr parses a let binding, `let` token already eaten.
// `let ident [: ty] = init`
func (p *parser) parseLetExpr(let_sp span.Span) ast.Expr {
	ident := p.parseIdent()
	if ident == nil {
		p.error("expected identifier in let binding, but got `%s`", p.tok.Kind)
	}

	var ty *ast.Ty
	if colonSp, ok := p.eat(token.Colon); ok {
		ty = p.parseTy()
	} else {
		ty = &ast.Ty{Kind: ast.TyInfer, Sp: colonSp}
	}

	if _, ok := p.eat(token.Eq); !ok {
		p.error("expected `=`, but got `%s`", p.tok.Kind)
	}

	init := p.parseExpr()
	if init == nil {
		p.error("expected expression, but got `%s`", p.tok.Kind)
	}

	if ident == nil || init == nil || ty == nil {
		return &ast.ErrExpr{}
	}

	sp := let_sp.To(init.Span())
	return &ast.LetExpr{Decl: &ast.VarDecl{Ident: ident, Ty: ty}, Init: init, Sp: sp}
}

func (p *parser) parseReturnExpr(retSp span.Span) ast.Expr {
	retLine := p.sess.File.Line(retSp.Start)
	currLine := p.sess.File.Line(p.tok.Sp.Start)
	if retLine == currLine {
		if expr := p.parseExpr(); expr != nil {
			return &ast.ReturnExpr{X: expr, Sp: retSp.To(expr.Span())}
		}
	}

	return &ast.ReturnExpr{Sp: retSp}
}

// parseBreakExpr parses `break [expr]`
// `break` token already eaten.
func (p *parser) parseBreakExpr(breakSp span.Span) ast.Expr {
	retLine := p.sess.File.Line(breakSp.Start)
	currLine := p.sess.File.Line(p.tok.Sp.Start)
	if retLine == currLine {
		if expr := p.parseExpr(); expr != nil {
			return &ast.BreakExpr{X: expr, Sp: breakSp.To(expr.Span())}
		}
	}

	return &ast.BreakExpr{Sp: breakSp}
}

func (p *parser) parsePrecExpr(min_prec int) ast.Expr {
	expr := p.parseCallOrIndexExpr()

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

func (p *parser) parseCallOrIndexExpr() ast.Expr {
	expr := p.parseBotExpr()

	if _, ok := p.eat(token.LParen); ok {
		var args []ast.Expr
		for !p.tok.IsOneOf(token.RParen, token.Eof) {
			args = append(args, p.parseExpr())
			if _, ok := p.eat(token.Comma); !ok {
				break
			}
		}

		rpSp, ok := p.eat(token.RParen)
		if !ok {
			p.error("expected closing delimiter `%s`", token.RParen)
			return &ast.ErrExpr{}
		}

		sp := expr.Span().To(rpSp)
		return &ast.CallExpr{Fn: expr, Args: args, Sp: sp}
	}

	if _, ok := p.eat(token.LBrack); ok {
		i := p.parseExpr()

		closeSp, ok := p.eat(token.RBrack)
		if !ok {
			p.error("expected closing delimiter `%s`", token.RBrack)
			return &ast.ErrExpr{}
		}
		sp := expr.Span().To(closeSp)
		return &ast.IndexExpr{Arr: expr, I: i, Sp: sp}
	}

	return expr
}

func (p *parser) parseBotExpr() ast.Expr {
	if sp, ok := p.eat(token.If); ok {
		return p.parseIfExpr(sp)
	}

	if sp, ok := p.eat(token.For); ok {
		return p.parseForExpr(sp)
	}

	if sp, ok := p.eat(token.Number); ok {
		return &ast.IntegerLiteral{V: p.prevTok.Lit, Sp: sp}
	}

	if sp, ok := p.eat(token.String); ok {
		return &ast.StringLiteral{V: p.prevTok.Lit, Sp: sp}
	}

	if sp, ok := p.eat(token.False); ok {
		return &ast.BooleanLiteral{V: false, Sp: sp}
	}

	if sp, ok := p.eat(token.True); ok {
		return &ast.BooleanLiteral{V: true, Sp: sp}
	}

	if sp, ok := p.eat(token.LBrack); ok {
		return p.parseArrayExpr(sp)
	}

	ident := p.parseIdent()
	if ident != nil {
		return p.parseClassExpr(ident)
	}

	return ident
}

func (p *parser) parseClassExpr(ident *ast.Ident) ast.Expr {
	if _, ok := p.eat(token.LBrace); !ok {
		return ident
	}

	var fields []*ast.ExprField
	for !p.tok.IsOneOf(token.RBrace, token.Eof) {
		ident := p.parseIdent()
		if ident == nil {
			p.error("expected field name, but got `%s`", p.tok.Kind)
			continue
		}

		if _, ok := p.eat(token.Colon); !ok {
			p.error("expected `:`")
		}

		expr := p.parseExpr()
		if expr == nil {
			p.error("expected expression after `:`")
			continue
		}

		field := &ast.ExprField{Ident: ident, Expr: expr}
		fields = append(fields, field)

		if _, ok := p.eat(token.Comma); !ok {
			break
		}
	}

	closeSp, ok := p.eat(token.RBrace)
	if !ok {
		p.error("expected closing delimiter `%s`", token.LBrace)
		return &ast.ErrExpr{}
	}

	sp := ident.Sp.To(closeSp)
	return &ast.ClassExpr{Ident: ident, Fields: fields, Sp: sp}

}

// parseArrayExpr parses `[expr, expr...]`
// `[` token already eaten.
func (p *parser) parseArrayExpr(openSp span.Span) ast.Expr {
	var exprs []ast.Expr
	for !p.tok.IsOneOf(token.RBrack, token.Eof) {
		exprs = append(exprs, p.parseExpr())
		if _, ok := p.eat(token.Comma); !ok {
			break
		}
	}

	closeSp, ok := p.eat(token.RBrack)
	if !ok {
		p.error("expected closing delimiter `%s`", token.RBrack)
		return &ast.ErrExpr{}
	}

	sp := openSp.To(closeSp)
	return &ast.ArrayExpr{Exprs: exprs, Sp: sp}
}

// parseIfExpr parses `if cond { exprs } [else [if cond] { exprs ]`
// `if` token already eaten.
func (p *parser) parseIfExpr(ifSp span.Span) ast.Expr {
	cond := p.parseExpr()
	if cond == nil {
		p.error("expected condition")
	}

	then := p.parseBlockExpr()
	sp := ifSp.To(then.Span())
	var els ast.Expr
	if _, ok := p.eat(token.Else); ok {
		els = p.parseElseExpr()
		sp = ifSp.To(els.Span())
	}
	return &ast.IfExpr{Cond: cond, Then: then, Else: els, Sp: sp}
}

// parseElseExpr parses `else [if cond] { exprs }`.
// `else` token already eaten.
func (p *parser) parseElseExpr() ast.Expr {
	if ifSp, ok := p.eat(token.If); ok {
		return p.parseIfExpr(ifSp)
	}
	return p.parseBlockExpr()
}

// parseForExpr parses `for { exprs }`
// `for` token already eaten.
func (p *parser) parseForExpr(forSp span.Span) ast.Expr {
	body := p.parseBlockExpr()
	sp := forSp.To(body.Span())
	return &ast.ForExpr{Body: body, Sp: sp}
}

// parseBlockExpr parses `{ exprs }`
func (p *parser) parseBlockExpr() ast.Expr {
	openSp, ok := p.eat(token.LBrace)
	if !ok {
		p.error("expected opening delimiter `%s`", token.LBrace)
		return &ast.ErrExpr{}
	}

	var exprs []ast.Expr
	for !p.tok.IsOneOf(token.RBrace, token.Eof) {
		exprs = append(exprs, p.parseExpr())
	}

	closeSp, ok := p.eat(token.RBrace)
	if !ok {
		p.error("expected closing delimiter `%s`", token.LBrace)
		return &ast.ErrExpr{}
	}

	sp := openSp.To(closeSp)
	return &ast.BlockExpr{Exprs: exprs, Sp: sp}
}

// `:` token already eaten.
func (p *parser) parseTy() *ast.Ty {
	if openSp, ok := p.eat(token.LBrack); ok {
		ident := p.parseIdent()
		if ident == nil {
			p.error("expected a type after `[`")
			return nil
		}
		closeSp, ok := p.eat(token.RBrack)
		if !ok {
			p.error("expected closing delimiter `%s`", token.LBrack)
			return nil
		}
		sp := openSp.To(closeSp)
		return &ast.Ty{Kind: ast.TyArray, Ident: ident, Sp: sp}
	}

	ident := p.parseIdent()
	if ident == nil {
		p.error("expected type after `:`")
		return nil
	}
	return &ast.Ty{Kind: ast.TyIdent, Ident: ident, Sp: ident.Sp}
}

func (p *parser) parseIdent() *ast.Ident {
	if sp, ok := p.eat(token.Ident); ok {
		return &ast.Ident{Name: p.prevTok.Lit, Sp: sp}
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
		p.prevTok = p.tok
		p.tok = p.tokens[p.pos]
		p.pos += 1
	}
}

func (p *parser) error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	diagnostic.NewBuilder(msg, p.tok.Sp).WithLabel("here").Emit(p.sess.Diags)
}
