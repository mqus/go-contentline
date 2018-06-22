package go_contentline

import (
	"fmt"
	"strings"
)

func ExampleInitParser() {
	r := strings.NewReader("BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"END:VCALENDAR\r\n")
	p := InitParser(r)
	c, _ := p.ParseNextObject()
	fmt.Printf("%q", c)
	//Output:
	// &{"VCALENDAR" [%!q(*go_contentline.Property=&{VERSION 2.0 map[] VERSION:2.0})] []}
}
