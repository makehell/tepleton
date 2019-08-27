# Tepleton
<img src="docs/tepleton_logo.png" width="250" height="250">

The Tepleton is a framework for building blockchain applications in Golang.


**Note**: Requires [Go 1.12+](https://golang.org/dl/)


# Basecoin

Basecoin is a sample [WRSP application](https://github.com/tepleton/wrsp) designed to be used with the [tepleton consensus engine](https://tepleton.com/) to form a Proof-of-Stake cryptocurrency. This project has two main purposes:

  1. As an example for anyone wishing to build a custom application using tepleton.
  2. As a framework for anyone wishing to build a tepleton-based currency, extensible using the plugin system.

## Contents

  1. [Installation](#installation)
  1. [Using the plugin system](#using-the-plugin-system)
  1. [Using the cli](#using-the-cli)
  1. [Tutorials and other reading](#tutorials-and-other-reading)
  1. [Contributing](#contributing)

## Installation

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get github.com/tepleton/basecoin
cd $GOPATH/src/github.com/tepleton/basecoin
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.

## Using the Plugin System

Basecoin is designed to serve as a common base layer for developers building cryptocurrency applications.
It handles public-key authentication of transactions, maintaining the balance of arbitrary types of currency (BTC, ATOM, ETH, MYCOIN, ...), 
sending currency (one-to-one or n-to-m multisig), and providing merkle-proofs of the state. 
These are common factors that many people wish to have in a crypto-currency system, 
so instead of trying to start from scratch, developers can extend the functionality of Basecoin using the plugin system!

The Plugin interface is defined in `types/plugin.go`:

```
type Plugin interface {
  Name() string
  SetOption(store KVStore, key string, value string) (log string)
  RunTx(store KVStore, ctx CallContext, txBytes []byte) (res wrsp.Result)
  InitChain(store KVStore, vals []*wrsp.Validator)
  BeginBlock(store KVStore, height uint64)
  EndBlock(store KVStore, height uint64) []*wrsp.Validator
}
```

`RunTx` is where you can handle any special transactions directed to your application. 
To see a very simple implementation, look at the demo [counter plugin](./plugins/counter/counter.go). 
If you want to create your own currency using a plugin, you don't have to fork basecoin at all.  
Just make your own repo, add the implementation of your custom plugin, and then build your own main script that instatiates Basecoin and registers your plugin.

An example is worth a 1000 words, so please take a look [at this example](https://github.com/tepleton/basecoin/blob/develop/cmd/paytovote/main.go#L25-L31). 
Note for now it is in a dev branch.
You can use the same technique in your own repo.

## Using the CLI

The basecoin cli can be used to start a stand-alone basecoin instance (`basecoin start`),
or to start basecoin with tepleton in the same process (`basecoin start --in-proc`).
It can also be used to send transactions, eg. `basecoin sendtx --to 0x4793A333846E5104C46DD9AB9A00E31821B2F301 --amount 100`
See `basecoin --help` and `basecoin [cmd] --help` for more details`.

## Tutorials and Other Reading

See our [introductory blog post](https://cosmos.network/blog/cosmos-creating-interoperable-blockchains-part-1), which explains the motivation behind Basecoin.

