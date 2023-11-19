package main

import osx "os"

func main() {
	defer osx.Exit(1) // want "direct call of os.Exit is not allowed"
	func() {
		osx.Exit(1) // want "direct call of os.Exit is not allowed"
	}()
	if true == false {
		osx.Exit(1) // want "direct call of os.Exit is not allowed"
	}
	go osx.Exit(1) // want "direct call of os.Exit is not allowed"
	osx.Exit(1)    // want "direct call of os.Exit is not allowed"
}

func notmain() {
	osx.Exit(1)
}
