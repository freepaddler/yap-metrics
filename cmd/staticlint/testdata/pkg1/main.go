package main

import "os"

func main() {
	defer os.Exit(1) // want "direct call of os.Exit is not allowed"
	func() {
		os.Exit(1) // want "direct call of os.Exit is not allowed"
	}()
	if true == false {
		os.Exit(1) // want "direct call of os.Exit is not allowed"
	}
	go os.Exit(1) // want "direct call of os.Exit is not allowed"
	// indirect call
	f := os.Exit
	f(1)
	os.Exit(1) // want "direct call of os.Exit is not allowed"
}

func notmain() {
	os.Exit(1)
}
