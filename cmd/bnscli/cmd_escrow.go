package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/cmd/bnsd/app"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/x/escrow"
)

func cmdReleaseEscrow(input io.Reader, output io.Writer, args []string) error {
	fl := flag.NewFlagSet("", flag.ExitOnError)
	fl.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), `
Create a transaction for releasing funds from given escrow.
		`)
		fl.PrintDefaults()
	}
	var (
		escrowFl = flHex(fl, "escrow", "", "A hex encoded ID of an escrow that is to be released.")
		amountFl = flCoin(fl, "amount", "", "Optional amount that is to be transferred from the escrow. The whole escrow hold amount is used if no value is provided.")
	)
	fl.Parse(args)

	var amount []*coin.Coin
	if !coin.IsEmpty(amountFl) {
		amount = append(amount, amountFl)
	}
	tx := &app.Tx{
		Sum: &app.Tx_ReleaseEscrowMsg{
			ReleaseEscrowMsg: &escrow.ReleaseEscrowMsg{
				Metadata: &weave.Metadata{Schema: 1},
				EscrowId: *escrowFl,
				Amount:   amount,
			},
		},
	}
	raw, err := tx.Marshal()
	if err != nil {
		return fmt.Errorf("cannot serialize transaction: %s", err)
	}
	_, err = output.Write(raw)
	return err
}