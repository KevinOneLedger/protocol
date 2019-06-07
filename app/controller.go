package app

import (
	"encoding/hex"
	"github.com/Oneledger/protocol/action"
	"github.com/Oneledger/protocol/serialize"
	"github.com/Oneledger/protocol/version"
	"github.com/tendermint/tendermint/libs/common"
	"golang.org/x/crypto/ripemd160"
)

// The following set of functions will be passed to the abciController

// query connection: for querying the application state; only uses query and Info
func (app *App) infoServer() infoServer {
	return func(info RequestInfo) ResponseInfo {
		return ResponseInfo{
			Data:             app.name,
			Version:          version.Fullnode.String(),
			LastBlockHeight:  app.header.Height,
			LastBlockAppHash: app.header.AppHash,
		}
	}
}

func (app *App) queryer() queryer {
	return func(RequestQuery) ResponseQuery {
		// Do stuff
		return ResponseQuery{}
	}
}

func (app *App) optionSetter() optionSetter {
	return func(RequestSetOption) ResponseSetOption {
		// TODO
		return ResponseSetOption{
			Code: CodeOK.uint32(),
		}
	}
}

// consensus methods: for executing transactions that have been committed. Message sequence is -for every block

func (app *App) chainInitializer() chainInitializer {
	return func(req RequestInitChain) ResponseInitChain {
		err := app.setupState(req.AppStateBytes)
		// This should cause consensus to halt
		if err != nil {
			app.logger.Error("Failed to setupState", "err", err)
			return ResponseInitChain{}
		}

		//update the initial validator set to db, this should always comes after setupState as the currency for
		// validator will be registered by setupState
		validators, err := app.setupValidators(req, app.Context.currencies)
		if err != nil {
			app.logger.Error("Failed to setupValidator", "err", err)
			return ResponseInitChain{}
		}
		app.logger.Error("finish chain initialize")
		return ResponseInitChain{Validators: validators}
	}
}

func (app *App) blockBeginner() blockBeginner {
	return func(req RequestBeginBlock) ResponseBeginBlock {

		//update the validator set
		err := app.Context.validators.Set(req)
		if err != nil {
			app.logger.Error("validator set with error", err)
		}
		//update the header to current block
		//todo: store the header in persistent db
		app.header = req.Header

		result := ResponseBeginBlock{
			Tags: []common.KVPair(nil),
		}

		app.logger.Debug("Begin Block:", result, "height=", req.Header.Height, "AppHash=", hex.EncodeToString(req.Header.AppHash))
		return result
	}
}

// mempool connection: for checking if transactions should be relayed before they are committed
func (app *App) txChecker() txChecker {
	return func(msg []byte) ResponseCheckTx {
		tx := &action.BaseTx{}

		err := serialize.GetSerializer(serialize.NETWORK).Deserialize(msg, tx)
		if err != nil {
			app.logger.Errorf("checkTx failed to deserialize msg: %s, error: %s ", msg, err)
		}

		txCtx := app.Context.Action()

		handler := txCtx.Router.Handler(tx.Data)

		ok, err := handler.Validate(txCtx, tx.Data, tx.Fee, tx.Memo, tx.Signatures)
		if err != nil {
			app.logger.Debugf("Check Tx invalid: ", err.Error())
			return ResponseCheckTx{
				Code: getCode(ok).uint32(),
				Log:  err.Error(),
			}
		}
		ok, response := handler.ProcessCheck(txCtx, tx.Data, tx.Fee)

		result := ResponseCheckTx{
			Code:      getCode(ok).uint32(),
			Data:      response.Data,
			Log:       response.Log,
			Info:      response.Info,
			GasWanted: response.GasWanted,
			GasUsed:   response.GasUsed,
			Tags:      response.Tags,
			Codespace: "",
		}
		app.logger.Debug("Check Tx: ", result, "log", response.Log)
		return result

	}
}

func (app *App) txDeliverer() txDeliverer {
	return func(msg []byte) ResponseDeliverTx {
		tx := &action.BaseTx{}

		err := serialize.GetSerializer(serialize.NETWORK).Deserialize(msg, tx)
		if err != nil {
			app.logger.Errorf("deliverTx failed to deserialize msg: %s, error: %s ", msg, err)
		}
		txCtx := app.Context.Action()

		handler := txCtx.Router.Handler(tx.Data)

		ok, response := handler.ProcessDeliver(txCtx, tx.Data, tx.Fee)

		result := ResponseDeliverTx{
			Code:      getCode(ok).uint32(),
			Data:      response.Data,
			Log:       response.Log,
			Info:      response.Info,
			GasWanted: response.GasWanted,
			GasUsed:   response.GasUsed,
			Tags:      response.Tags,
			Codespace: "",
		}
		app.logger.Debug("Deliver Tx: ", result)
		return result
	}
}

func (app *App) blockEnder() blockEnder {
	return func(req RequestEndBlock) ResponseEndBlock {

		updates := app.Context.validators.GetEndBlockUpdate(app.Context.ValidatorCtx(), req)

		result := ResponseEndBlock{
			ValidatorUpdates: updates,
			Tags:             []common.KVPair(nil),
		}
		app.logger.Debug("End Block: ", result, "height=", req.Height)
		return result
	}
}

func (app *App) commitor() commitor {
	return func() ResponseCommit {

		// Commit any pending changes.
		hashb, verb := app.Context.balances.Commit()

		hashv, verv := app.Context.validators.Commit()

		apphash := &appHash{}
		apphash.Hashes = append(apphash.Hashes, hashb, hashv)

		_, _ = verb, verv
		hash := apphash.hash()
		app.logger.Debugf("Committed New Block height[%d], hash[%s]", app.header.Height, hex.EncodeToString(hash))

		result := ResponseCommit{
			Data: hash,
		}

		app.logger.Debug("Commit Result", result)
		return result
	}
}

func getCode(ok bool) (code Code) {
	if ok {
		code = CodeOK
	} else {
		code = CodeNotOK
	}
	return
}

//todo: make appHash to use a commiter function to finish the commit and hashing for all the store that passed
type appHash struct {
	Hashes [][]byte `json:"hashes"`
}

func (ah *appHash) hash() []byte {
	result, _ := serialize.GetSerializer(serialize.JSON).Serialize(ah)
	hasher := ripemd160.New()
	hasher.Write(result)
	return hasher.Sum(nil)
}
