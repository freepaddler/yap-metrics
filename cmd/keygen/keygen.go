package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/freepaddler/yap-metrics/internal/pkg/crypt"
)

func exitOnErr(str string, a ...any) {
	fmt.Printf(str+"\n", a)
	os.Exit(1)
}

func main() {
	keySize := flag.Int("n", 1024, "keys size bytes")
	outPath := flag.String("o", ".", "output directory")
	flag.Parse()

	dir, err := os.Stat(*outPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(*outPath, os.ModePerm); err != nil {
				exitOnErr("Unable to create output dir '%s': %s", outPath, err)
			}
		} else {
			exitOnErr(fmt.Sprintf("%s", err))
		}
	} else {
		if !dir.IsDir() {
			exitOnErr("Output destination '%s' is not a dir", *outPath)
		}
	}

	pubFilePath := *outPath + "/public.key"
	pubFile, err := os.Create(pubFilePath)
	if err != nil {
		exitOnErr("Unable to write public key file %s", pubFilePath)
	}
	defer pubFile.Close()
	privFilePath := *outPath + "/private.key"
	privFile, err := os.Create(privFilePath)
	if err != nil {
		exitOnErr("Unable to write private key file %s", privFilePath)
	}
	defer privFile.Close()

	err = crypt.WritePair(pubFile, privFile, *keySize)
	if err != nil {
		exitOnErr("Error %w", err)
	}

	fmt.Println("Private key is written to: " + privFilePath)
	fmt.Println("Public key is written to: " + pubFilePath)
}
