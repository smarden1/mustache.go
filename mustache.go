package mustache

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"reflect"
	"strings"
)

var defaultOtag = "{{"
var defaultCtag = "}}"

// Commands are the valid commands in mustache.
var commands = map[string]bool{
	"#": true,
	"^": true,
	"/": true,
	"<": true,
	">": true,
	"=": true,
	"!": true,
	"&": true,
}

// Token represents a command or bit of text in the template represented as a tree strucutre.
type token struct {
	cmd        string
	args       string
	within     bool
	notEscaped bool
	children   []*token
}

// AddChild adds a child token to the current token
func (t *token) addChild(child *token) {
	t.children = append(t.children, child)
}

// IsEmpty returns a boolean indicating if token has no value
func (t *token) isEmpty() bool {
	return len(t.children) == 0 && t.cmd == "" && t.args == ""
}

// Render recursively walks the tokens and writes it's output to a buffer.
func (t *token) render(cstack []interface{}, output *bytes.Buffer) {
	if t.within {
		if t.cmd == "#" {
			if val, ok := contextStackContains(cstack, t.args); ok && !isFalsey(val) {
				kind := reflect.TypeOf(val).Kind()
				if kind == reflect.Array || kind == reflect.Slice {
					a := reflect.ValueOf(val)

					for i := 0; i < a.Len(); i++ {
						for _, child := range t.children {
							child.render(append(cstack, a.Index(i).Interface()), output)
						}
					}
				} else if kind == reflect.Map {
					for _, child := range t.children {
						child.render(append(cstack, val), output)
					}
				} else {
					for _, child := range t.children {
						child.render(cstack, output)
					}
				}
			}
		} else if t.cmd == "^" {
			if val, ok := contextStackContains(cstack, t.args); !ok || isFalsey(val) {
				for _, t := range t.children {
					t.render(cstack, output)
				}
			}
		} else if t.cmd == "" {
			if val, ok := contextStackContains(cstack, t.args); ok {
				s := fmt.Sprint(val)
				if !t.notEscaped {
					s = html.EscapeString(s)
				}
				output.WriteString(s)
			}
			for _, child := range t.children {
				child.render(cstack, output)
			}
		}
	} else {
		output.WriteString(t.args)
	}
}

// Compile will take compile a template into a token.
// The entire compiled template is held by a root token.
// Sections are represented as children to current token.
func compile(template string, rootToken *token, buffer *bytes.Buffer, lineTokenPointers []*token) (*token, []*token, error) {
	var err error
	tripleTag, withinTag, notEscaped := false, false, false // booleans that indicate state
	otag, ctag := defaultOtag, defaultCtag                  // opening and closing tags
	sections := []*token{rootToken}                         // section stack
	cmd := ""                                               // current command for this token

	for i := 0; i < len(template); i++ {
		s := string(template[i])
		if withinTag {
			if tripleTag && matchesTag(template, i, "}}}") {
				tripleTag, withinTag = false, false
				notEscaped = true
				i += 2
			} else if matchesTag(template, i, ctag) {
				withinTag = false
				i += len(ctag) - 1
			} else if _, ok := commands[s]; ok && cmd == "" {
				cmd = s
			} else if !isWhiteSpace(s) || cmd == "=" {
				buffer.WriteString(s)
			}
			// we just closed the tag, we should evaluate it
			if !withinTag {
				var currentToken token
				currentToken, err = newToken(cmd, buffer, true, notEscaped)
				lineTokenPointers = append(lineTokenPointers, &currentToken)
				notEscaped = false
				cmd = ""

				if currentToken.cmd == "/" {
					if len(sections) > 0 && sections[len(sections)-1].args == currentToken.args {
						sections = sections[:len(sections)-1]
						lineTokenPointers = addTokenToLastToken(&currentToken, lineTokenPointers, sections)
					} else {
						err = fmt.Errorf("Malformed template: %s was closed but not opened", currentToken.args)
					}
				} else if currentToken.cmd == ">" {
					b, err := ioutil.ReadFile(currentToken.args + ".mustache")

					if err == nil {
						_, lineTokenPointers, err = compile(string(b), sections[len(sections)-1], buffer, lineTokenPointers)
					}
				} else if currentToken.cmd == "=" {
					otag, ctag = parseDelimiters(currentToken.args)
				} else {
					lastToken := sections[len(sections)-1]
					lastToken.children = append(lastToken.children, &currentToken)

					if currentToken.cmd == "#" || currentToken.cmd == "^" {
						sections = append(sections, &currentToken)
					}
				}
			}
		} else {
			if matchesTag(template, i, "{{{") {
				tripleTag, withinTag = true, true
				i += 2
			} else if matchesTag(template, i, otag) {
				withinTag = true
				i += len(otag) - 1
			} else {
				// lines are valid if they contain actual values on them,
				// just a section should not make a newline to the final output
				// hwowever, a line with just whitespace or a single newline is valid
				if isNewLine(s) {
					if !shouldKeepWhiteSpace(lineTokenPointers, buffer) {
						cleanWhiteSpaceOnPastTokens(lineTokenPointers)
						// handle windows carriage returns
						if matchesTag(template, i, "\r\n") {
							i++
						}
						buffer.Reset()
					} else {
						buffer.WriteString(s)
					}
					lineTokenPointers = []*token{}
					currentToken, _ := newToken("", buffer, false, true)
					addTokenToLastToken(&currentToken, lineTokenPointers, sections)
				} else {
					buffer.WriteString(s)
				}
			}
			// we just opened it so set state
			if withinTag {
				var currentToken token
				currentToken, err = newToken(cmd, buffer, false, false)
				lineTokenPointers = addTokenToLastToken(&currentToken, lineTokenPointers, sections)
			}
		}
	}

	if !shouldKeepWhiteSpace(lineTokenPointers, buffer) {
		cleanWhiteSpaceOnPastTokens(lineTokenPointers)
		buffer.Reset()
	}
	var currentToken token
	currentToken, err = newToken(cmd, buffer, false, false)
	addTokenToLastToken(&currentToken, lineTokenPointers, sections)

	if len(sections) > 1 {
		err = fmt.Errorf("Malformed template: %s was not closed", sections[len(sections)-1].args)
	}
	return rootToken, lineTokenPointers, err
}

func addTokenToLastToken(tkn *token, lineTokenPointers []*token, sections []*token) []*token {
	if !tkn.isEmpty() {
		lastToken := sections[len(sections)-1]
		lastToken.addChild(tkn)
	}

	return append(lineTokenPointers, tkn)
}

func cleanWhiteSpaceOnPastTokens(lineTokenPointers []*token) {
	// clear out whitespace
	for _, tkn := range lineTokenPointers {
		if tkn.cmd == "" && !tkn.within {
			tkn.args = ""
		}
	}
}

// ShouldKeepWhiteSpace returns a boolean which indicates whether or not the
// whitespace on a line should be kept or not.
// Whitespace should be removed if the line is only there for making a template legible as a template
// but is not desired for the final output.
// i.e. - in lines that only contain {{/foo}}, then the line should not introduce additional whitespace
func shouldKeepWhiteSpace(lineTokenPointers []*token, buffer *bytes.Buffer) bool {
	if !isStringCompletelyWhiteSpace(buffer.String()) || len(lineTokenPointers) == 0 {
		return true
	}

	for _, tkn := range lineTokenPointers {
		t := *tkn
		if t.cmd == "" && !isStringCompletelyWhiteSpace(t.args) {
			return true
		}
	}

	return false
}

func isStringCompletelyWhiteSpace(s string) bool {
	for _, chr := range s {
		if !isWhiteSpace(string(chr)) {
			return false
		}
	}
	return true
}

// IsWhiteSpace returns a boolean indicating if this character is a whitespace
func isWhiteSpace(chr string) bool {
	return chr == "" || chr == " " || chr == "\t" || isNewLine(chr) || rune(chr[0]) < 32
}

// IsNewLine returns a boolean indicating if this character is a newline character
func isNewLine(chr string) bool {
	return chr == "\n" || chr == "\r"
}

// MatchesTag looks ahead to see if the given tag is found in the template.
func matchesTag(template string, i int, tag string) bool {
	l := len(tag)

	return len(template)-i >= l && template[i:i+l] == tag
}

// ParseDelimiters parses a delimiter command and returns the opening and closing tags
func parseDelimiters(args string) (string, string) {
	var splitArgs []string
	for _, e := range strings.Split(args, " ") {
		e = strings.Replace(e, "=", "", -1)
		if !isStringCompletelyWhiteSpace(e) {
			splitArgs = append(splitArgs, e)
		}
	}

	return splitArgs[0], splitArgs[len(splitArgs)-1]
}

// NewToken is a constructor for token and will return a new token based on several parameters.
func newToken(cmd string, b *bytes.Buffer, within, notEscaped bool) (token, error) {
	var err error
	if !within && cmd != "" {
		err = fmt.Errorf("Parser error: there should be no command for text, received %s", cmd)
	}

	t := token{}
	if cmd == "&" {
		t.cmd = ""
		notEscaped = true
	} else {
		t.cmd = cmd
	}

	t.notEscaped = notEscaped
	t.within = within
	t.args = b.String()
	b.Reset()

	return t, err
}

// IsFalsey returns a boolean indicating whether the value is "falsey"
func isFalsey(val interface{}) bool {
	v := reflect.ValueOf(val)
	switch reflect.TypeOf(val).Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	default:
		return fmt.Sprint(val) == ""
	}
}

// ContextStackContains recursively walks the context stack to see if the given key is available.
// It will return the value and an ok.
func contextStackContains(cstack []interface{}, key string) (interface{}, bool) {
	for i := len(cstack) - 1; i >= 0; i-- {
		c := cstack[i]

		if key == "." {
			return c, true
		}

		k := reflect.TypeOf(c).Kind()
		v := reflect.ValueOf(c)
		if k == reflect.Map {
			if val := v.MapIndex(reflect.ValueOf(key)); val.IsValid() {
				return val.Interface(), true
			}
		} else if k == reflect.Struct {
			if val := v.FieldByName(key); val.IsValid() {
				return val.Interface(), true
			}
		}
	}
	// a solitary "." is the implicit operator
	if strings.Contains(key, ".") {
		s := strings.Split(key, ".")
		searchstack := cstack
		var r interface{}
		for _, prefix := range s {
			var ok bool
			r, ok = contextStackContains(searchstack, prefix)
			searchstack = []interface{}{r}
			if !ok {
				return nil, false
			}
		}
		return r, true
	}

	return nil, false
}

// Template is a compiled template
type Template struct {
	token *token
}

// Compile will compile a template. Compiled templates are faster if you use them more then once,
// otherwise prefer Render.
func Compile(template string) (*Template, error) {
	var b bytes.Buffer
	t, _, err := compile(template, &token{within: true}, &b, []*token{})

	return &Template{t}, err
}

// Render will render a template using the provided data.
func (t *Template) Render(c ...interface{}) string {
	var b bytes.Buffer
	t.token.render(c, &b)

	return b.String()
}

// Render will render a template using the provided data.
func Render(template string, c ...interface{}) (string, error) {
	t, err := Compile(template)

	s := t.Render(c...)

	if err != nil {
		return s, err
	}

	return s, nil
}
