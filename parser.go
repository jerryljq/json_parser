package main

import (
	"fmt"
)

var htmlColorMap = map[string]string{
	"{":       "\"color:rgb(0, 0, 255)\"",
	"}":       "\"color:rgb(0, 0, 255)\"",
	"[":       "\"color:rgb(138, 43, 226)\"",
	"]":       "\"color:rgb(138, 43, 226)\"",
	":":       "\"color:rgb(0, 0, 0)\"",
	",":       "\"color:rgb(46, 139, 87)\"",
	"kw":      "\"color:rgb(255, 127, 80)\"",
	"str":     "\"color:rgb(210, 105, 30)\"",
	"num":     "\"color:rgb(255, 0, 0)\"",
	"escaped": "\"color:rgb(0, 200, 50)\"",
}

type Parser struct {
	tokens     *TokenStream
	html       string
	num_indent int
}

// punc should be the desired punctuation
func (p *Parser) is_punc(punc string) bool {
	tok := p.tokens.peek()
	if tok == (Token{}) || tok.t_type != "punc" || tok.value.(string) != punc {
		return false
	} else {
		return true
	}

}

func (p *Parser) skip_punc(punc string) {
	if p.is_punc(punc) {
		p.html_punc(punc) // the consuming of token is deferred to the html function
	} else {
		tok := p.tokens.peek()
		msg := "expected punctuation " + punc + " but got " + tok.value.(string)
		p.tokens.input.err_msg(msg)
	}
}

// return the string of the html output
func (p *Parser) parse_toplevel() string {
	p.html = "<span style=\"font-family:monospace; white-space:pre\">\n"
	p.parse_value()
	p.html += "\n</span>"
	return p.html
}

func (p *Parser) parse_object() {
	first := true
	p.num_indent += 1
	p.skip_punc("{")
	p.html_new_line()
	for !p.tokens.is_eof() {
		if p.is_punc("}") {
			p.html_new_line()
			p.num_indent -= 1
			p.html_tab()
			p.html_punc("}")
			break
		}
		if first {
			first = false
		} else {
			p.skip_punc(",")
			p.html_new_line()
		}
		if p.is_punc("}") { // allow an object end with a comma
			p.html_new_line()
			p.num_indent -= 1
			p.html_tab()
			p.html_punc("}")
			break
		}
		p.html_tab()
		p.parse_string()
		p.skip_punc(":")
		p.parse_value()
	}
}

// parse a sequence of values, since the length of sequence can be arbitrarily long, parse_array()
// should be recursive by itself
func (p *Parser) parse_array() {
	first := true
	p.skip_punc("[")
	for !p.tokens.is_eof() {
		if p.is_punc("]") {
			p.html_punc("]")
			break
		}
		if first {
			first = false
		} else {
			p.skip_punc(",")
		}
		if p.is_punc("]") { // allow an array to end with a comma
			p.html_punc("]")
			break
		}
		p.parse_value()
	}
}

// this is a quite core part, the logic can enter into different parts
// depending on the type of value
func (p *Parser) parse_value() {
	switch {
	case p.is_punc("{"):
		p.parse_object()
	case p.is_punc("["):
		p.parse_array()
	case p.tokens.peek().t_type == "str":
		p.parse_string()
	case p.tokens.peek().t_type == "num":
		p.html_num()
	case p.tokens.peek().t_type == "kw":
		p.html_kw()
	default:
		p.tokens.input.err_msg("invalid token! Expected a JSON value")
	}
}

func (p *Parser) parse_string() {
	tok := p.tokens.peek()
	if tok.t_type == "str" {
		p.html_string()
	} else {
		p.tokens.input.err_msg("expected string but got token type" + tok.t_type)
	}
}

func (p *Parser) html_string() {
	str_tok := p.tokens.next() // read and consume
	fmt_str := p._html_special_string(str_tok.value.(string))
	p.html += p.html_wrap_color("str", fmt_str)
}

func (p *Parser) _html_special_string(s string) string {
	new_s := ""
	escaped := false
	escaped_count := 0
	for _, r := range s {
		added := ""
		if escaped == false {
			if string(r) == "\\" {
				escaped = true
				escaped_count += 1
				added = "<span style=" + htmlColorMap["escaped"] + ">" + string(r)
			} else {
				added = p._html_s_trans(string(r))
			}
		} else { // under escaped mode
			if string(r) == "u" {
				escaped_count += 4
			}
			added = p._html_s_trans(string(r))
			escaped_count -= 1
			if escaped_count == 0 {
				escaped = false
				added += "</span>"
			}
		}
		new_s += added
	}
	new_s = "&quot;" + new_s + "&quot;"
	return new_s
}

func (p *Parser) _html_s_trans(c string) string {
	trans := ""
	switch {
	case c == "<":
		trans = "&lt;"
	case c == ">":
		trans = "&gt;"
	case c == "&":
		trans = "&amp;"
	case c == "\"":
		trans = "&quot;"
	case c == "'":
		trans = "&apos;"
	default:
		trans = c
	}
	return trans
}

func (p *Parser) html_num() {
	num_tok := p.tokens.next() // read and consume
	p.html += p.html_wrap_color("num", num_tok.value.(string))
}

func (p *Parser) html_kw() {
	kw_tok := p.tokens.next()
	p.html += p.html_wrap_color("kw", kw_tok.value.(string))
}

func (p *Parser) html_punc(punc string) {
	p.tokens.next() // consume the punc token
	p.html += p.html_wrap_color(punc, punc)
}

func (p *Parser) html_new_line() {
	p.html += "\n"
}

func (p *Parser) html_tab() {
	for i := 0; i < p.num_indent; i++ {
		p.html += "\t"
	}
}

func (p *Parser) html_wrap_color(tk_type string, raw_s string) string {
	wrapped_s := "<span style=" + htmlColorMap[tk_type] + ">" + raw_s + "</span>"
	return wrapped_s
}

func main() {
	var ts TokenStream
	ts.set_up("test2.json")
	var p Parser
	p.tokens = &ts
	result := p.parse_toplevel()
	fmt.Println(result)
}
