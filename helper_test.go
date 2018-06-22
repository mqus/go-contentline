package go_contentline

import "fmt"

func ExampleValidID() {
	// Is valid:
	fmt.Printf("%v\n", ValidID("X-DRESSCODE"))

	// Is invalid:
	fmt.Print(string(*ValidID("Not-So-'Valid'")))

	//Output:
	//<nil>
	//'
}
