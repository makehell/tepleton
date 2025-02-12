package app

import (
	"fmt"
	"strings"

	wrsp "github.com/tepleton/wrsp/types"
	sm "github.com/tepleton/basecoin/state"
	"github.com/tepleton/basecoin/types"
	. "github.com/tepleton/go-common"
	"github.com/tepleton/go-wire"
	eyes "github.com/tepleton/merkleeyes/client"
)

const (
	version   = "0.1"
	maxTxSize = 10240

	PluginNameBase = "base"
)

type Basecoin struct {
	eyesCli    *eyes.Client
	state      *sm.State
	cacheState *sm.State
	plugins    *types.Plugins
}

func NewBasecoin(eyesCli *eyes.Client) *Basecoin {
	state := sm.NewState(eyesCli)
	plugins := types.NewPlugins()
	return &Basecoin{
		eyesCli:    eyesCli,
		state:      state,
		cacheState: nil,
		plugins:    plugins,
	}
}

// TMSP::Info
func (app *Basecoin) Info() wrsp.ResponseInfo {
	return wrsp.ResponseInfo{Data: Fmt("Basecoin v%v", version)}
}

func (app *Basecoin) RegisterPlugin(plugin types.Plugin) {
	app.plugins.RegisterPlugin(plugin)
}

// TMSP::SetOption
func (app *Basecoin) SetOption(key string, value string) (log string) {
	PluginName, key := splitKey(key)
	if PluginName != PluginNameBase {
		// Set option on plugin
		plugin := app.plugins.GetByName(PluginName)
		if plugin == nil {
			return "Invalid plugin name: " + PluginName
		}
		return plugin.SetOption(app.state, key, value)
	} else {
		// Set option on basecoin
		switch key {
		case "chainID":
			app.state.SetChainID(value)
			return "Success"
		case "account":
			var err error
			var acc *types.Account
			wire.ReadJSONPtr(&acc, []byte(value), &err)
			if err != nil {
				return "Error decoding acc message: " + err.Error()
			}
			app.state.SetAccount(acc.PubKey.Address(), acc)
			return "Success"
		}
		return "Unrecognized option key " + key
	}
}

// TMSP::DeliverTx
func (app *Basecoin) DeliverTx(txBytes []byte) (res wrsp.Result) {
	if len(txBytes) > maxTxSize {
		return wrsp.CodeType_BaseEncodingError, nil, "Tx size exceeds maximum"
	}

	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return wrsp.CodeType_BaseEncodingError, nil, "Error decoding tx: " + err.Error()
	}

	// Validate and exec tx
	res = sm.ExecTx(app.state, app.plugins, tx, false, nil)
	if res.IsErr() {
		return res.PrependLog("Error in DeliverTx")
	}
	// Store accounts
	storeAccounts(app.eyesCli, accs)
	return wrsp.CodeType_OK, nil, "Success"
}

// TMSP::CheckTx
func (app *Basecoin) CheckTx(txBytes []byte) (code wrsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return wrsp.CodeType_BaseEncodingError, nil, "Tx size exceeds maximum"
	}

	fmt.Printf("%X\n", txBytes)

	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return wrsp.CodeType_BaseEncodingError, nil, "Error decoding tx: " + err.Error()
	}

	// Validate tx
	res = sm.ExecTx(app.cacheState, app.plugins, tx, true, nil)
	if res.IsErr() {
		return res.PrependLog("Error in CheckTx")
	}
	return wrsp.CodeType_OK, nil, "Success"
}

// TMSP::Query
func (app *Basecoin) Query(reqQuery wrsp.RequestQuery) (resQuery wrsp.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = wrsp.CodeType_EncodingError
		return
	}

	resQuery, err := app.eyesCli.QuerySync(reqQuery)
	if err != nil {
		resQuery.Log = "Failed to query MerkleEyes: " + err.Error()
		resQuery.Code = wrsp.CodeType_InternalError
		return
	}
	return
}

// TMSP::Commit
func (app *Basecoin) Commit() (res wrsp.Result) {

	// Commit state
	res = app.state.Commit()

	// Wrap the committed state in cache for CheckTx
	app.cacheState = app.state.CacheWrap()

	if res.IsErr() {
		PanicSanity("Error getting hash: " + res.Error())
	}
	return hash, "Success"
}

// TMSP::InitChain
func (app *Basecoin) InitChain(validators []*wrsp.Validator) {
	for _, plugin := range app.plugins.GetList() {
		plugin.InitChain(app.state, validators)
	}
}

// TMSP::BeginBlock
func (app *Basecoin) BeginBlock(height uint64) {
	for _, plugin := range app.plugins.GetList() {
		plugin.BeginBlock(app.state, height)
	}
}

// TMSP::EndBlock
func (app *Basecoin) EndBlock(height uint64) (diffs []*wrsp.Validator) {
	for _, plugin := range app.plugins.GetList() {
		moreDiffs := plugin.EndBlock(app.state, height)
		diffs = append(diffs, moreDiffs...)
	}
	return
}

//----------------------------------------

// Splits the string at the first '/'.
// if there are none, the second string is nil.
func splitKey(key string) (prefix string, suffix string) {
	if strings.Contains(key, "/") {
		keyParts := strings.SplitN(key, "/", 2)
		return keyParts[0], keyParts[1]
	}
	return key, ""
}
