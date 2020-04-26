package app

import (
	"fmt"
)

func ExampleSyncEventBus() {
	b := NewSyncEventBus()
	b.Sub("test", func(e Event) {
		fmt.Println("start sub")
		fmt.Println("receive", e.Data)
		fmt.Println("end sub")
	})

	fmt.Println("start pub")
	b.Pub("test", "hello")
	fmt.Println("end pub")
	// Output:
	// start pub
	// start sub
	// receive hello
	// end sub
	// end pub
}
