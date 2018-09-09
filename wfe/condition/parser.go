package condition

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
	"regexp"
	"strings"
	"text/scanner"
)

var namePattern = regexp.MustCompile(`\A[a-z][a-zA-Z0-9_]*\z`)

type parser struct {
	str string
	scn scanner.Scanner
}

func Parse(str string) api.Condition {
	p := &parser{}
	p.str = str
	p.scn.Init(strings.NewReader(str))
	c, r := p.parseOr()
	if r != scanner.EOF {
		panic(eval.Error(WF_CONDITION_SYNTAX_ERROR, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
	}
	return c
}

func (p *parser) parseOr() (api.Condition, rune) {
	es := make([]api.Condition, 0)
	for {
		lh, r := p.parseAnd()
		es = append(es, lh)
		if p.scn.TokenText() != `or` {
			if len(es) == 1 {
				return es[0], r
			}
			return Or(es), r
		}
	}
}

func (p *parser) parseAnd() (api.Condition, rune) {
	es := make([]api.Condition, 0)
	for {
		lh, r := p.parseUnary()
		es = append(es, lh)
		if p.scn.TokenText() != `and` {
			if len(es) == 1 {
				return es[0], r
			}
			return And(es), r
		}
	}
}

func (p *parser) parseUnary() (c api.Condition, r rune) {
	r = p.scn.Scan()
	if r == '!' {
		c, r = p.parseAtom(p.scn.Scan())
		return Not(c), r
	}
	return p.parseAtom(r)
}

func (p *parser) parseAtom(r rune) (api.Condition, rune) {
	if r == scanner.EOF {
		panic(eval.Error(WF_CONDITION_UNEXPECTED_END, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
	}

	if r == '(' {
		var c api.Condition
		c, r = p.parseOr()
		if r != ')' {
			panic(eval.Error(WF_CONDITION_MISSING_RP, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
		}
		return c, p.scn.Scan()
	}
	w := p.scn.TokenText()
	if namePattern.MatchString(w) {
		r = p.scn.Scan()
		switch w {
		case `true`:
			return Always, r
		case `false`:
			return Never, r
		default:
			return Truthy(w), r
		}
	}
	panic(eval.Error(WF_CONDITION_INVALID_NAME, issue.H{`name`: w, `text`: p.str, `pos`: p.scn.Offset}))
}
