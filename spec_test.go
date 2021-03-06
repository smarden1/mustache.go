package mustache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type Spec struct {
	Name     string                 `json:"name"`
	Data     map[string]interface{} `json:"data"`
	Expected string                 `json:"expected"`
	Template string                 `json:"template"`
	Partials map[string]string      `json:"partials"`
}

type SpecFile struct {
	Attn     string `json:"__ATTN__"`
	Overview string `json:"overview"`
	Tests    []Spec `json:"tests"`
}

// specs to be ignore in the following format "fileNameWithoutSuffix-TestName"
var ignoreSpecList = map[string]bool{
	"interpolation-HTML Escaping":               true, // go uses different quoting character then the mustache spec
	"partials-Recursion":                        true, // not implemented
	"partials-Standalone Without Previous Line": true, // edge cases with adding additional space
	"partials-Standalone Without Newline":       true, // edge cases with adding additional space
	"partials-Standalone Indentation":           true, // edge cases with adding additional space
}

// TestSpec runs all of the mustache spec files except for the ones in the ignoreSpecList
func TestSpec(t *testing.T) {
	files, _ := ioutil.ReadDir("spec/specs/")

	for _, file := range files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".json") && !strings.HasPrefix(fileName, "~") {
			RunSpecFile(t, fileName)
		}
	}
}

func RunSpecFile(t *testing.T, fileName string) {

	var tests SpecFile

	b, _ := ioutil.ReadFile(fmt.Sprintf("spec/specs/%s", fileName))
	json.Unmarshal(b, &tests)

	for _, test := range tests.Tests {
		if shouldRunSpec(fileName, test.Name) {
			files := makePartials(&test, fileName, test.Name)

			if output, _ := Render(test.Template, test.Data); output != test.Expected {
				t.Errorf("%s:%s, recieved %q and expected %q", fileName, test.Name, output, test.Expected)
			}
			for _, file := range files {
				os.Remove(file)
			}
		}
	}
}

// MakePartials writes the partials to a directory and returns an array of the paths.
// Arguments are the spec and and the index of the test, just to distinguish
func makePartials(test *Spec, fileName string, testName string) []string {
	if len(test.Partials) == 0 {
		return []string{}
	}

	dir, _ := os.Getwd()
	var files []string
	testSuite := strings.Replace(fileName, ".json", "", 1)

	partials := make(map[string]string, len(test.Partials))
	for name, content := range test.Partials {
		dirName := fmt.Sprintf("%s/spec/partials/%s/%s", dir, testSuite, strings.Replace(testName, " ", "_", -1))
		fileName := fmt.Sprintf("%s.mustache", name)

		os.MkdirAll(dirName, 0777)
		ioutil.WriteFile(fileName, []byte(content), 0666)

		files = append(files, fileName)
		partials[content] = fileName
	}
	test.Partials = partials

	return files
}

// shouldRunSpec returns true if we should run this spec
func shouldRunSpec(fileName, testName string) bool {
	specKey := fmt.Sprintf("%s-%s", strings.TrimSuffix(fileName, ".json"), testName)
	_, ok := ignoreSpecList[specKey]

	return !ok
}
