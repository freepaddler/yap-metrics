package main

type fake struct{}

func (f fake) Exit() {
}

func main() {
	os := new(fake)
	os.Exit()
}
