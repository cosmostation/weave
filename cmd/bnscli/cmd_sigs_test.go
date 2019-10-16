package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"io"
	"io/ioutil"
	"testing"

	"github.com/iov-one/weave"
	bnsd "github.com/iov-one/weave/cmd/bnsd/app"
	"github.com/iov-one/weave/crypto"
	"github.com/iov-one/weave/weavetest/assert"
	"github.com/iov-one/weave/x/cash"
)

// TestSignMessage is an internal test code by Cosmostation
func TestSignMessage(t *testing.T) {
	const mnemonic = `guide worth axis butter craft donkey beef carry mechanic road seven food example ensure tip unit various flight antenna shuffle drill slim eyebrow lava`
	priv, err := keygen(mnemonic, "m/44'/234'/0'")
	if err != nil {
		t.Fatalf("cannot generate key: %s", err)
	}

	privKey := crypto.PrivKeyEd25519FromSeed(priv.Seed())

	msg := []byte("foobar")

	sig, err := privKey.Sign(msg)
	assert.Nil(t, err)

	bz, err := sig.Marshal()
	assert.Nil(t, err)

	t.Logf("privKey %s", privKey)
	t.Logf("sigStr %s", sig)
	t.Logf("hexSigStr %s", hex.EncodeToString(bz))
}

func TestCmdSignTransactionHappyPath(t *testing.T) {
	tx := &bnsd.Tx{
		Sum: &bnsd.Tx_CashSendMsg{
			CashSendMsg: &cash.SendMsg{
				Metadata: &weave.Metadata{Schema: 1},
			},
		},
	}
	var input bytes.Buffer
	if _, err := writeTx(&input, tx); err != nil {
		t.Fatalf("cannot marshal transaction: %s", err)
	}

	var output bytes.Buffer
	args := []string{
		"-tm", tmURL,
		"-key", mustCreateFile(t, bytes.NewReader(fromHex(t, privKeyHex))),
	}
	if err := cmdSignTransaction(&input, &output, args); err != nil {
		t.Fatalf("transaction signing failed: %s", err)
	}

	tx, _, err := readTx(&output)
	if err != nil {
		t.Fatalf("cannot read created transaction: %s", err)
	}

	if n := len(tx.Signatures); n != 1 {
		t.Fatalf("want one signature, got %d", n)
	}
}

var logRequestFl = flag.Bool("logrequest", false, "Log all requests send to tendermint mock server. This is useful when writing new test. Use curl to send the same request to a real tendermint node and record the response.")

func mustCreateFile(t testing.TB, r io.Reader) string {
	t.Helper()

	fd, err := ioutil.TempFile("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	if _, err := io.Copy(fd, r); err != nil {
		t.Fatal(err)
	}
	if err := fd.Close(); err != nil {
		t.Fatal(err)
	}
	return fd.Name()
}
