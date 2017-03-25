package main

import "flag"
import "fmt"
import "os"

func autoforgeUsage() {
	fmt.Printf("Usage: %s [options] [package_range]\n\n", os.Args[0])

	flag.PrintDefaults()
}

func main() {
	flag.Usage = autoforgeUsage

	init := flag.Bool("init", false, "Initialize the working copy")

	flag.Parse()

	if *init {
		fmt.Println("Initializing the working copy...")
	}
}
