package action

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/Oneledger/protocol/data/accounts"
	"github.com/Oneledger/protocol/data/balance"
	"github.com/Oneledger/protocol/data/bitcoin"
	"github.com/Oneledger/protocol/data/ethereum"
	"github.com/Oneledger/protocol/data/fees"
	"github.com/Oneledger/protocol/data/jobs"
	"github.com/Oneledger/protocol/data/ons"
	"github.com/Oneledger/protocol/identity"
	"github.com/Oneledger/protocol/log"
	"github.com/Oneledger/protocol/storage"
)

type Context struct {
	Router          Router
	State           *storage.State
	Header          *abci.Header
	Accounts        accounts.Wallet
	Balances        *balance.Store
	Domains         *ons.DomainStore
	FeePool         *fees.Store
	Currencies      *balance.CurrencySet
	FeeOpt          *fees.FeeOption
	Validators      *identity.ValidatorStore
	Witnesses       *identity.WitnessStore
	BTCTrackers     *bitcoin.TrackerStore
	ETHTrackers     *ethereum.TrackerStore
	Logger          *log.Logger
	JobStore        *jobs.JobStore
	LockScriptStore *bitcoin.LockScriptStore
}

func NewContext(r Router, header *abci.Header, state *storage.State,
	wallet accounts.Wallet, balances *balance.Store,
	currencies *balance.CurrencySet, feePool *fees.Store,
	validators *identity.ValidatorStore, witnesses *identity.WitnessStore,
	domains *ons.DomainStore, btcTrackers *bitcoin.TrackerStore,
	ethTrackers *ethereum.TrackerStore, jobStore *jobs.JobStore,
	lockScriptStore *bitcoin.LockScriptStore, logger *log.Logger) *Context {

	return &Context{
		Router:          r,
		State:           state,
		Header:          header,
		Accounts:        wallet,
		Balances:        balances,
		Domains:         domains,
		FeePool:         feePool,
		Currencies:      currencies,
		Validators:      validators,
		Witnesses:       witnesses,
		BTCTrackers:     btcTrackers,
		ETHTrackers:     ethTrackers,
		Logger:          logger,
		JobStore:        jobStore,
		LockScriptStore: lockScriptStore,
	}
}
