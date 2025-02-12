package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/urfave/cli"

	cmn "github.com/tepleton/go-common"
	"github.com/tepleton/go-merkle"
	"github.com/tepleton/go-wire"
	tmtypes "github.com/tepleton/tepleton/types"
)

func cmdQuery(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("query command requires an argument ([key])")
	}
	keyString := c.Args()[0]
	key := []byte(keyString)
	if isHex(keyString) {
		// convert key to bytes
		var err error
		key, err = hex.DecodeString(stripHex(keyString))
		if err != nil {
			return errors.New(cmn.Fmt("Query key (%v) is invalid hex: %v", keyString, err))
		}
	}

	resp, err := query(c.String("node"), key)
	if err != nil {
		return err
	}

	if !resp.Code.IsOK() {
		return errors.New(cmn.Fmt("Query for key (%v) returned non-zero code (%v): %v", keyString, resp.Code, resp.Log))
	}

	val := resp.Value
	proof := resp.Proof
	height := resp.Height

	fmt.Println(string(wire.JSONBytes(struct {
		Value  []byte `json:"value"`
		Proof  []byte `json:"proof"`
		Height uint64 `json:"height"`
	}{val, proof, height})))

	return nil
}

func cmdAccount(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("account command requires an argument ([address])")
	}
	addrHex := stripHex(c.Args()[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.New(cmn.Fmt("Account address (%v) is invalid hex: %v", addrHex, err))
	}

	acc, err := getAcc(c.String("node"), addr)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(acc)))
	return nil
}

func cmdBlock(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("block command requires an argument ([height])")
	}
	heightString := c.Args()[0]
	height, err := strconv.Atoi(heightString)
	if err != nil {
		return errors.New(cmn.Fmt("Height must be an int, got %v: %v", heightString, err))
	}

	block, err := getBlock(c, height)
	if err != nil {
		return err
	}
	nextBlock, err := getBlock(c, height+1)
	if err != nil {
		return err
	}

	fmt.Println(string(wire.JSONBytes(struct {
		Hex  BlockHex  `json:"hex"`
		JSON BlockJSON `json:"json"`
	}{
		BlockHex{
			Header: wire.BinaryBytes(block.Header),
			Commit: wire.BinaryBytes(nextBlock.LastCommit),
		},
		BlockJSON{
			Header: block.Header,
			Commit: nextBlock.LastCommit,
		},
	})))

	return nil
}

type BlockHex struct {
	Header []byte `json:"header"`
	Commit []byte `json:"commit"`
}

type BlockJSON struct {
	Header *tmtypes.Header `json:"header"`
	Commit *tmtypes.Commit `json:"commit"`
}

func cmdVerify(c *cli.Context) error {
	keyString, valueString := c.String("key"), c.String("value")

	var err error
	key := []byte(keyString)
	if isHex(keyString) {
		key, err = hex.DecodeString(stripHex(keyString))
		if err != nil {
			return errors.New(cmn.Fmt("Key (%v) is invalid hex: %v", keyString, err))
		}
	}

	value := []byte(valueString)
	if isHex(valueString) {
		value, err = hex.DecodeString(stripHex(valueString))
		if err != nil {
			return errors.New(cmn.Fmt("Value (%v) is invalid hex: %v", valueString, err))
		}
	}

	root, err := hex.DecodeString(stripHex(c.String("root")))
	if err != nil {
		return errors.New(cmn.Fmt("Root (%v) is invalid hex: %v", c.String("root"), err))
	}

	proofBytes, err := hex.DecodeString(stripHex(c.String("proof")))
	if err != nil {
		return errors.New(cmn.Fmt("Proof (%v) is invalid hex: %v", c.String("proof"), err))
	}

	proof, err := merkle.ReadProof(proofBytes)
	if err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling proof: %v", err))
	}

	if proof.Verify(key, value, root) {
		fmt.Println("OK")
	} else {
		return errors.New("Proof does not verify")
	}
	return nil
}
