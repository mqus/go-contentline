package go_contentline

import (
	"fmt"
	"testing"
)

func ExampleUnescapeParamVal() {
	in := "A ^'String^' with more than^n2^^0^Ncharacters!"
	fmt.Println(UnescapeParamVal(in))
	// Output:
	// A "String" with more than
	// 2^0
	// characters!
}

func ExampleEscapeParamVal() {
	in := "A \"String\" with more than \n2^0\n characters!"
	fmt.Println(EscapeParamVal(in))
	// Output: A ^'String^' with more than ^n2^^0^n characters!
}

func TestUnescapeParamVal(t *testing.T) {
	checks := map[string]string{
		"":               "",
		"^^":             "^",
		"^'":             "\"",
		"^n":             "\n",
		"^N":             "\n",
		"^m":             "^m",
		"^^n":            "^n",
		"^^^'":           "^\"",
		"^^^'^^^n^'^^N^": "^\"^\n\"^N^",
		"^^^^":           "^^",
		"^^^^n":          "^^n",
		"^^^^^n":         "^^\n",
	}
	for in, want := range checks {
		strFun(UnescapeParamVal).assert(t, in, want)
	}
}

func TestEscapeParamVal(t *testing.T) {
	checks := map[string]string{
		"":            "",
		"^":           "^^",
		"\"":          "^'",
		"\n":          "^n",
		"\r\n":        "^n",
		"^m":          "^^m",
		"^n":          "^^n",
		"^\"":         "^^^'",
		"^\"^\n\"^N^": "^^^'^^^n^'^^N^^",
		"^^":          "^^^^",
		"^^n":         "^^^^n",
		"^^\n":        "^^^^^n",
	}
	for in, want := range checks {
		strFun(EscapeParamVal).assert(t, in, want)
	}
}

type strFun func(string) string

func (f strFun) assert(t *testing.T, in, want string) {
	if got := f(in); got != want {
		t.Errorf("Wanted: '%s'\n Got:%s", want, got)
	}
}
