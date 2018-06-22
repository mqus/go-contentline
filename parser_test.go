package go_contentline

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
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

func TestParser_ParseNextObject(t *testing.T) {
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", nil, nil})
	//check Component with inner Component
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"BEGIN:inner\r\n"+
			"END:inner\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", nil, []*Component{{"INNER", nil, nil}}})
	//check Property
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE:Content:'!,;.'\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "Content:'!,;.'", make(Parameters), "FEATURE:Content:'!,;.'"}}, nil})
	//check unfolding
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE:Conten\r\n"+
			" t:'!,;.'\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "Content:'!,;.'", make(Parameters), "FEATURE:Content:'!,;.'"}}, nil})
	//check Parameter
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE;LANG=en:LoremIpsum\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "LoremIpsum", map[string][]string{"LANG": {"en"}}, "FEATURE;LANG=en:LoremIpsum"}}, nil})
	//check quoted Parameter
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE;LAng=\"e;n\":LoremIpsum\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "LoremIpsum", map[string][]string{"LANG": {"e;n"}}, "FEATURE;LAng=\"e;n\":LoremIpsum"}}, nil})
	//check RFC6868-Escaping
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE;LANG=e^^^n:LoremIpsum\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "LoremIpsum", map[string][]string{"LANG": {"e^\n"}}, "FEATURE;LANG=e^^^n:LoremIpsum"}}, nil})
	//check multiple Parameters with multiple values, variably encoded and folded
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE;Par1=e^'^n,\"other^,val\";PAR2=\"\r\n"+
			" display:none;\",not interesting:LoremIpsum\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "LoremIpsum", map[string][]string{"PAR1": {"e\"\n", "other^,val"}, "PAR2": {"display:none;", "not interesting"}}, "FEATURE;Par1=e^'^n,\"other^,val\";PAR2=\"display:none;\",not interesting:LoremIpsum"}}, nil})
	//check property in nested Component
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"BEGIN:iNnErCoMp\r\n"+
			"FEATURE;LAng=\"e;n\":LoremIpsum\r\n"+
			"END:InNeRcOmP\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", nil, []*Component{{"INNERCOMP", []*Property{{"FEATURE", "LoremIpsum", map[string][]string{"LANG": {"e;n"}}, "FEATURE;LAng=\"e;n\":LoremIpsum"}}, nil}}})
	//check property next to nested Component
	parseCompare(t,
		"BEGIN:comp\r\n"+
			"FEATURE;LAng=\"e;n\":LoremIpsum\r\n"+
			"BEGIN:iNnErCoMp\r\n"+
			"END:InNeRcOmP\r\n"+
			"FEATURE;LAng2=\"e;n\":LoremIpsum\r\n"+
			"END:Comp\r\n",
		&Component{"COMP", []*Property{{"FEATURE", "LoremIpsum", map[string][]string{"LANG": {"e;n"}}, "FEATURE;LAng=\"e;n\":LoremIpsum"}, {"FEATURE", "LoremIpsum", map[string][]string{"LANG2": {"e;n"}}, "FEATURE;LAng2=\"e;n\":LoremIpsum"}}, []*Component{{"INNERCOMP", nil, nil}}})

}

func parseCompare(t *testing.T, in string, want *Component) {
	t.Helper()
	p := InitParser(strings.NewReader(in))
	got, e := p.ParseNextObject()
	if e != nil {
		t.Error(e.Error())
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Differences found, Wanted:\n%q\nGot:\n%q\n",
			want,
			got)
	}

}
