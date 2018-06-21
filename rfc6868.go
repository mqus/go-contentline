package go_contentline

import (
	"strings"
)

func UnescapeParamVal(in string) string {
	//first, find all '^^' and replace them with '^^ ', to separate them from their ^n etc meanings
	s1 := strings.Replace(in, "^^", "^^ ", -1)
	//Then replace all the usual stuff
	s2 := strings.Replace(s1, "^n", "\n", -1)
	s3 := strings.Replace(s2, "^N", "\n", -1)
	s4 := strings.Replace(s3, "^'", "\"", -1)
	//lastly, replace all '^^' of the original string (now '^^ ') with '^'
	return strings.Replace(s4, "^^ ", "^", -1)
}

func EscapeParamVal(in string) string {
	//first, replace all '^' with '^^'
	s1 := strings.Replace(in, "^", "^^", -1)
	//Then replace all the usual stuff
	s2 := strings.Replace(s1, "\"", "^'", -1)
	s3 := strings.Replace(s2, "\r\n", "^n", -1)
	s4 := strings.Replace(s3, "\n", "^n", -1)
	return strings.Replace(s4, "\r", "^n", -1)
}
