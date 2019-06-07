package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ed25519"
)

func cmdKeygen(input io.Reader, output io.Writer, args []string) error {
	fl := flag.NewFlagSet("", flag.ExitOnError)
	fl.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), `
Generate a new private key.

When successful a new file with binary content containing private key is
created. This command fails if the private key file already exists.
`)
		fl.PrintDefaults()
	}
	var (
		keyPathFl = fl.String("key", env("BNSCLI_PRIV_KEY", os.Getenv("HOME")+"/.bnsd.priv.key"),
			"Path to the private key file that transaction should be signed with. You can use BNSCLI_PRIV_KEY environment variable to set it.")
	)
	fl.Parse(args)

	if _, err := os.Stat(*keyPathFl); !os.IsNotExist(err) {
		// Do not allow to overwrite already existing private key. User
		// must manually delete it first to ensure we do not delete
		// such crucial data by an accident (bad command usage).
		return fmt.Errorf("private key file %q already exists, delete this file and try again", *keyPathFl)
	}

	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("cannot generate ed25519 key: %s", err)
	}

	fd, err := os.OpenFile(*keyPathFl, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("cannot create public key file: %s", err)
	}
	defer fd.Close()

	if _, err := fd.Write(priv); err != nil {
		return fmt.Errorf("cannot write private key: %s", err)
	}
	if err := fd.Close(); err != nil {
		return fmt.Errorf("cannot close private key file: %s", err)
	}
	return nil
}
