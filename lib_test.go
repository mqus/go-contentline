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
	c := &Component{
		Name: "House",
	}
	encodeCompare(t, c, "BEGIN:HOUSE\r\nEND:HOUSE\r\n")
}

func encodeCompare(t *testing.T, in *Component, want string) {
	t.Helper()
	var buf strings.Builder
	//buf := bytes.NewBuffer(make([]byte, 1024*1024))
	in.Encode(&buf)
	got := buf.String()
	if got != want {
		t.Errorf("Differences found, Wanted:\n%q\nGot:\n%q\n", want, got)
	}

}
