package main

import "flag"
import "fmt"

func main() {
	init := flag.Bool("init", false, "Initialize the working copy")

	flag.Parse()

	if *init {
		fmt.Println("Initializing the working copy...")
	}
}
