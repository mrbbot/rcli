package rcli

import (
	"fmt"
	"os"
)

func ExampleApp() {
	// create an empty app
	a := NewApp()

	// create a basic command
	a.Command("hello <name:string>", func(name string) {
		fmt.Printf("hello %s\n", name)
	})

	// create a command with a default value
	a.Command("goodbye <name:string=person>", func(name string) {
		fmt.Printf("goodbye %s\n", name)
	})

	// commands can have no arguments
	a.Command("ping", func() {
		fmt.Println("pong")
	})

	// or they can have many
	a.Command("count <from:int> <to:int> <double:bool=false>", func(from, to int, double bool) {
		multiplier := 1
		if double {
			multiplier = 2
		}
		for i := from; i <= to; i++ {
			fmt.Println(i * multiplier)
		}
	})

	// run the app using the passed command line arguments
	a.Run(os.Args)
}
