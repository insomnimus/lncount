package engine

import (
	"regexp"
	"strings"
)

type compiler struct {
	input        string
	ch           byte
	pos, readpos int
}

type Regexp struct {
	*regexp.Regexp
	CompiledFrom string
	Literal      string
}

func MustCompile(s string, caseInsensitive bool) *Regexp {
	r, err := Compile(s, caseInsensitive)
	if err != nil {
		panic(err)
	}
	return r
}

func Compile(s string, caseInsensitive bool) (*Regexp, error) {
	c := &compiler{input: s}
	c.read()
	text := c.compile()
	if caseInsensitive {
		text = "(?i)" + text
	}
	reg, err := regexp.Compile(text)
	if err != nil {
		return nil, err
	}
	return &Regexp{reg, s, text}, nil
}

func (c *compiler) read() {
	if c.readpos >= len(c.input) {
		c.ch = 0
	} else {
		c.ch = c.input[c.readpos]
	}
	c.pos = c.readpos
	c.readpos++
}

func (c *compiler) peek() byte {
	if c.readpos >= len(c.input) {
		return 0
	}
	return c.input[c.readpos]
}

func (c *compiler) compile() string {
	var out strings.Builder
LOOP:
	for {
		switch c.ch {
		case 0:
			break LOOP
		case '\\':
			c.read()
			switch c.ch {
			case '*':
				out.WriteString("\\*")
			case '?':
				out.WriteString("\\?")
			default:
				out.WriteByte('\\')
				out.WriteByte(c.ch)
			}
		case '*':
			if c.peek() == '*' {
				c.read()
				out.WriteString(".*")
			} else {
				out.WriteString(".*?")
			}
		case '[':
			out.WriteString(c.readRange())
		case '?':
			out.WriteString(".?")
		case '-', ',', '.', '{', '(', ')', '}', ']', ':':
			out.WriteByte('\\')
			out.WriteByte(c.ch)
		default:
			out.WriteByte(c.ch)
		}
		c.read()
	}
	return out.String()
}

func (c *compiler) readRange() string {
	var out strings.Builder
LOOP:
	for {
		switch c.ch {
		case 0:
			break LOOP
		case ']':
			break LOOP
		case '\\':
			c.read()
			out.WriteByte('\\')
			out.WriteByte(c.ch)
		default:
			out.WriteByte(c.ch)
		}
		c.read()
	}
	return out.String()
}
