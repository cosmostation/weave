package main

import (
	"bytes"
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/cmd/bnsd/app"
	"github.com/iov-one/weave/x/cash"
)

func TestCmdTransactionViewHappyPath(t *testing.T) {
	tx := &app.Tx{
		Sum: &app.Tx_SendMsg{
			SendMsg: &cash.SendMsg{
				Metadata: &weave.Metadata{Schema: 1},
				Memo:     "a memo",
				Ref:      []byte("123"),
			},
		},
	}
	rawTx, err := tx.Marshal()
	if err != nil {
		t.Fatalf("cannot marshal transaction: %s", err)
	}

	var output bytes.Buffer
	if err := cmdTransactionView(bytes.NewReader(rawTx), &output, nil); err != nil {
		t.Fatalf("cannot view a transaction: %s", err)
	}

	const want = `{
	"Sum": {
		"SendMsg": {
			"metadata": {
				"schema": 1
			},
			"memo": "a memo",
			"ref": "MTIz"
		}
	}
}`
	got := output.String()

	if want != got {
		t.Logf("want: %s", want)
		t.Logf(" got: %s", got)
		t.Fatal("unexpected view result")
	}
}
