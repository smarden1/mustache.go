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
		context  interface{}
		expected string
	}

	type foo struct {
		Foo string
		Bar int
	}

	expected := [...]expects{
		expects{"hello {{name}}", map[string]string{"name": "steve"}, "hello steve"},
		expects{"hello {{name}} {{ name}}", map[string]string{"name": "steve"}, "hello steve steve"},
		expects{"hello {{first_name}} {{last_name}}", map[string]string{"first_name": "steve"}, "hello steve "},
		expects{"hello {{first_name}} {{last_name}}", map[string]string{"first_name": "steve", "last_name": "m"}, "hello steve m"},
		expects{"{{name}}", map[string]string{"name": "stove", "last_name": "m"}, "stove"},
		expects{"{{#name}}yes{{/name}}", map[string]string{"name": "stove", "last_name": "m"}, "yes"},
		expects{"{{#name}}{{name}}{{/name}}", map[string]string{"name": "stove"}, "stove"},
		expects{"foo{{^name}}{{name}}{{/name}}", map[string]string{"name": "stove"}, "foo"},
		expects{"foo{{^bar}}biz{{/bar}}", map[string]string{"name": "stove"}, "foobiz"},
		expects{"{{#bar}}{{biz}}{{/bar}}", map[string]interface{}{"bar": []map[string]string{{"biz": "bar"}, {"biz": "none"}}}, "barnone"},
		expects{"{{#bar}}{{foo}}{{/bar}}", map[string]interface{}{"foo": "bar", "bar": []map[string]string{{"biz": "bar"}, {"biz": "none"}}}, "barbar"},
		expects{"{{name}}{{> test-assets/partial }}", map[string]string{"name": "stove", "foo": "bar"}, "stovebar"},
		expects{"hello {{name}}", map[string]string{"name": "steve&steve"}, "hello steve&amp;steve"},
		expects{"hello {{{name}}}", map[string]string{"name": "steve&steve"}, "hello steve&steve"},
		expects{"hello {{&name}}!", map[string]string{"name": "steve&steve"}, "hello steve&steve!"},
		expects{"{{num}} / {{dem}}", map[string]int{"num": 1, "dem": 10}, "1 / 10"},
		expects{"{{Foo}}", foo{"bar", 10}, "bar"},
		expects{"{{foo}}", foo{"bar", 10}, ""},
	}

	for _, e := range expected {
		if r, _ := Render(e.template, e.context); r != e.expected {
			t.Errorf("Incorrect rendered template, got %s, expected %s", r, e.expected)
		}
	}
}

func TestContextStackContains(t *testing.T) {
	m := map[string]map[string]string{
		"a":   {"b": "ab"},
		"foo": {"bar": "biz"},
	}

	type expects struct {
		key   string
		valid bool
	}

	expected := [...]expects{
		expects{"a", true},
		expects{"c", false},
		expects{"a.d", false},
		expects{"a.b", true},
	}

	for _, e := range expected {
		_, ok := contextStackContains([]interface{}{m}, e.key)
		if ok != e.valid {
			t.Errorf("Incorrect contextStackContains, got %t, expected %t for key %v", ok, e.valid, e.key)
		}
	}
}

func TestIsFalsey(t *testing.T) {
	a := [...]interface{}{
		false,
		"",
		[]string{},
		map[string]int{},
	}

	for _, e := range a {
		if !isFalsey(e) {
			t.Errorf("Falsey value %s was truthy", e)
		}
	}
}

func TestIsNotFalsey(t *testing.T) {
	a := [...]interface{}{
		true,
		"true",
		"anything",
		[]string{"a"},
		map[string]int{"a": 2},
	}

	for _, e := range a {
		if isFalsey(e) {
			t.Errorf("Truthy value %s was falsey", e)
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
