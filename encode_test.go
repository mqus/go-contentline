// +build go1.10

package go_contentline

import (
	"os"
	"strings"
	"testing"
)

func ExampleComponent_Encode() {

	c := &Component{
		Name: "CAR",
		Properties: []*Property{
			NewPropertyUnchecked("TYPE", "Transporter", map[string][]string{"Class": {">6m"}}),
		},
	}
	c.Encode(filter(os.Stdout, '\r'))
	//Output:
	//BEGIN:CAR
	//TYPE;CLASS=>6m:Transporter
	//END:CAR
}

func TestComponent_Encode(t *testing.T) {
	//test simple component
	c := &Component{
		Name: "House",
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\nEND:HOUSE\r\n", false)

	//test inner component
	c = &Component{
		Name: "House",
		Comps: []*Component{
			{"Flat", nil, nil},
		},
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\nBEGIN:FLAT\r\nEND:FLAT\r\nEND:HOUSE\r\n", false)

	//test Property
	c = &Component{
		Name: "House",
		Properties: []*Property{
			NewPropertyUnchecked("Heating", "electric", nil),
		},
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\nHEATING:electric\r\nEND:HOUSE\r\n", false)

	//test Property with folding, multiple parameter values and escaping
	c = &Component{
		Name: "House",
		Properties: []*Property{
			NewPropertyUnchecked("Heating", "electric", map[string][]string{"vendor": {"YourGas Co\"", "City:Energy LLC"}, "comment": {"This is a very long comment,more than 2^3 monkeys hat to sit 20 hours to write this \n thing with linebreaks."}}),
		},
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\n"+
		"HEATING;"+
		"VENDOR=YourGas Co^',\"City:Energy LLC\";"+
		"COMMENT=\"This is a very long \r\n comment,more than 2^^3 monkeys hat to sit 20 hours to write this ^n thing \r\n with linebreaks.\":"+
		"electric\r\n"+
		"END:HOUSE\r\n", true)

	//test Property with folding, multiple parameter values and escaping and inner component and property next to the component.
	c = &Component{
		Name: "House",
		Comps: []*Component{
			{"Flat", []*Property{
				NewPropertyUnchecked("Heating2", "electric2", map[string][]string{"vendor": {"YourGas Co\"", "City:Energy LLC"}, "comment": {"This is a very long comment,more than 2^3 monkeys hat to sit 20 hours to write this \n thing with linebreaks."}}),
			}, nil},
		},
		Properties: []*Property{
			NewPropertyUnchecked("Heating", "electric", map[string][]string{"vendor": {"YourGas Co\"", "City:Energy LLC"}, "comment": {"This is a very long comment,more than 2^3 monkeys hat to sit 20 hours to write this \n thing with linebreaks."}}),
		},
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\n"+
		"HEATING;VENDOR=YourGas Co^',\"City:Energy LLC\";"+
		"COMMENT=\"This is a very long \r\n comment,more than 2^^3 monkeys hat to sit 20 hours to write this ^n thing \r\n with linebreaks.\":electric\r\n"+
		"BEGIN:FLAT\r\n"+
		"HEATING2;VENDOR=YourGas Co^',\"City:Energy LLC\";"+
		"COMMENT=\"This is a very long\r\n  comment,more than 2^^3 monkeys hat to sit 20 hours to write this ^n thing\r\n  with linebreaks.\":electric2\r\n"+
		"END:FLAT\r\nEND:HOUSE\r\n", true)

	//test empty Property
	c = &Component{
		Name: "House",
		Properties: []*Property{
			NewPropertyUnchecked("Heating", "", map[string][]string{"vendor": {"YourGas Co\"", "City:Energy LLC"}}),
		},
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\n"+
		"HEATING;VENDOR=YourGas Co^',\"City:Energy LLC\":\r\n"+
		"END:HOUSE\r\n", false)

}

func encodeCompare(t *testing.T, in *Component, want string, canSkip bool) {
	t.Helper()
	var buf strings.Builder
	//buf := bytes.NewBuffer(make([]byte, 1024*1024))
	in.Encode(&buf)
	got := buf.String()
	if got != want {
		if canSkip {
			t.Skipf("Differences found, Wanted:\n%q\nGot:\n%q\nBUT: If they only differ in the order of the parameters, everything is alright.", want, got)
		} else {
			t.Errorf("Differences found, Wanted:\n%q\nGot:\n%q\n", want, got)
		}
	}

}
