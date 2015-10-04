package mustache

import "testing"

func TestCompile(t *testing.T) {
	p, _ := compile("hello {{name}} ")
	expected := []string{"hello ", "name"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}

	p, _ = compile("hello {{name}}")
	expected = []string{"hello ", "name"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}

	p, _ = compile("hello {{name}}, goodbye {{name}}")
	expected = []string{"hello ", "name", ", goodbye ", "name"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}

	p, _ = compile(" hello {{  name  }}")
	expected = []string{" hello ", "name"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}
}

func TestCompileSetTag(t *testing.T) {
	p, _ := compile("{{=<% %>=}}<% erb_style_tags %><%={{ }}=%>{{test}}")
	expected := []string{"erb_style_tags", "test"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}
}

func TestCompileComments(t *testing.T) {
	p, _ := compile("{{  name  }}{{! blah}}{{gnome   }}{{ !# blah}}")
	expected := []string{"name", "gnome"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}
}

/*func TestCompilePartial(t *testing.T) {
	p, _ := compile("{{name}}{{> test-assets/partial }}")
	expected := []string{"name", "foo"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}
}*/

func TestCompileSection(t *testing.T) {
	p, _ := compile("hello, {{#name}}again, {{first_name}} {{last_name}}{{/name}}")
	expected := []string{"hello, ", "name"}

	for i, e := range expected {
		if p.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p.children[i].args)
		}
	}

	expected = []string{"again, ", "first_name", " ", "last_name"}
	p2 := *p.children[1]
	for i, e := range expected {
		if p2.children[i].args != e {
			t.Errorf("Invalid arguments while parsing, expected %s but got %s", e, p2.children[i].args)
		}
	}
}

func TestRender(t *testing.T) {
	type expects struct {
		template string
		context  Context
		expected string
	}

	expected := [...]expects{
		expects{"hello {{name}}", Context{"name": "steve"}, "hello steve"},
		expects{"hello {{name}} {{ name}}", Context{"name": "steve"}, "hello steve steve"},
		expects{"hello {{first_name}} {{last_name}}", Context{"first_name": "steve"}, "hello steve "},
		expects{"hello {{first_name}} {{last_name}}", Context{"first_name": "steve", "last_name": "m"}, "hello steve m"},
		expects{"{{name}}", Context{"name": "stove", "last_name": "m"}, "stove"},
		expects{"{{#name}}yes{{/name}}", Context{"name": "stove", "last_name": "m"}, "yes"},
		expects{"{{#name}}{{name}}{{/name}}", Context{"name": "stove"}, "stove"},
		expects{"foo{{^name}}{{name}}{{/name}}", Context{"name": "stove"}, "foo"},
		expects{"foo{{^bar}}biz{{/bar}}", Context{"name": "stove"}, "foobiz"},
		expects{"{{#bar}}{{biz}}{{/bar}}", Context{"bar": []Context{{"biz": "bar"}, {"biz": "none"}}}, "barnone"},
		expects{"{{#bar}}{{foo}}{{/bar}}", Context{"foo": "bar", "bar": []Context{{"biz": "bar"}, {"biz": "none"}}}, "barbar"},
		expects{"{{name}}{{> test-assets/partial }}", Context{"name": "stove", "foo": "bar"}, "stovebar"},
		expects{"hello {{name}}", Context{"name": "steve&steve"}, "hello steve&amp;steve"},
		expects{"hello {{{name}}}", Context{"name": "steve&steve"}, "hello steve&steve"},
		expects{"hello {{&name}}!", Context{"name": "steve&steve"}, "hello steve&steve!"},
	}

	for _, e := range expected {
		if r, _ := Render(e.template, e.context); r != e.expected {
			t.Errorf("Incorrect rendered template, got %s, expected %s", r, e.expected)
		}
	}
}

func TestMatchesTag(t *testing.T) {
	template := "abcde {{ efg }} fg }}} < }>"

	type expects struct {
		i   int
		tag string
	}

	e := [...]expects{
		expects{0, "a"},
		expects{0, "abc"},
		expects{6, "{{"},
		expects{6, "{{ "},
		expects{13, "}}"},
		expects{19, "}}"},
		expects{19, "}}}"},
		expects{25, "}>"},
		expects{26, ">"},
	}

	for _, ex := range e {
		if !matchesTag(template, ex.i, ex.tag) {
			t.Errorf("Unable to find expectedTag of %s with index %d", ex.tag, ex.i)
		}
	}

	e2 := [...]expects{
		expects{0, "b"},
		expects{0, "bc"},
		expects{6, " {{"},
		expects{7, "{{ "},
		expects{13, "} "},
		expects{21, "}}"},
		expects{26, "}>"},
	}

	for _, ex := range e2 {
		if matchesTag(template, ex.i, ex.tag) {
			t.Errorf("Found erroneous expectedTag of %s with index %d", ex.tag, ex.i)
		}
	}
}
