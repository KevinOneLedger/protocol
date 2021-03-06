package governance

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Oneledger/protocol/data/keys"
	"github.com/Oneledger/protocol/serialize"
	"github.com/Oneledger/protocol/storage"
)

type ProposalFundStore struct {
	State  *storage.State
	prefix []byte
}

func (st *ProposalFundStore) set(key storage.StoreKey, amt ProposalAmount) error {
	dat, err := serialize.GetSerializer(serialize.PERSISTENT).Serialize(amt)
	if err != nil {
		return errors.Wrap(err, errorSerialization)
	}
	prefixed := append(st.prefix, key...)
	err = st.State.Set(prefixed, dat)
	return errors.Wrap(err, errorSettingRecord)
}

func (st *ProposalFundStore) get(key storage.StoreKey) (amt *ProposalAmount, err error) {
	prefixed := append(st.prefix, storage.StoreKey(key)...)
	dat, err := st.State.Get(prefixed)
	if err != nil {
		return nil, errors.Wrap(err, errorGettingRecord)
	}
	amt = NewAmount(0)
	if len(dat) == 0 {
		return
	}
	err = serialize.GetSerializer(serialize.PERSISTENT).Deserialize(dat, amt)
	if err != nil {
		err = errors.Wrap(err, errorDeSerialization)
	}
	return
}

func (pf *ProposalFundStore) delete(key storage.StoreKey) (bool, error) {
	prefixed := append(pf.prefix, key...)
	res, err := pf.State.Delete(prefixed)
	if err != nil {
		return false, errors.Wrap(err, errorDeletingRecord)
	}
	return res, err
}

func (pf *ProposalFundStore) iterate(fn func(proposalID ProposalID, addr keys.Address, amt *ProposalAmount) bool) bool {
	return pf.State.IterateRange(
		pf.prefix,
		storage.Rangefix(string(pf.prefix)),
		true,
		func(key, value []byte) bool {

			amt := NewAmount(0)
			err := serialize.GetSerializer(serialize.PERSISTENT).Deserialize(value, amt)
			if err != nil {
				fmt.Println("err", err)
				return true
			}
			arr := strings.Split(string(key), storage.DB_PREFIX)
			proposalID := arr[1]
			fundingAddress := keys.Address(arr[len(arr)-1])
			return fn(ProposalID(proposalID), fundingAddress, amt)
		},
	)
}

// Store Function Called my external Layers
func NewProposalFundStore(prefix string, state *storage.State) *ProposalFundStore {
	return &ProposalFundStore{
		State:  state,
		prefix: storage.Prefix(prefix),
	}
}

func (pf *ProposalFundStore) GetFundersForProposalID(id ProposalID, fn func(proposalID ProposalID, fundingAddr keys.Address, amt *ProposalAmount) ProposalFund) []ProposalFund {
	var foundProposals []ProposalFund
	pf.iterate(func(proposalID ProposalID, fundingAddr keys.Address, amt *ProposalAmount) bool {
		if proposalID == id {
			foundProposals = append(foundProposals, fn(proposalID, fundingAddr, amt))
		}
		return false
	})
	return foundProposals
}

func (pf *ProposalFundStore) GetProposalsForFunder(funderAddress keys.Address, fn func(proposalID ProposalID, fundingAddr keys.Address, amt *ProposalAmount) ProposalFund) []ProposalFund {
	var foundProposals []ProposalFund
	pf.iterate(func(proposalID ProposalID, fundingAddr keys.Address, amt *ProposalAmount) bool {
		if bytes.Equal(keys.Address(funderAddress.String()), fundingAddr) {
			foundProposals = append(foundProposals, fn(proposalID, fundingAddr, amt))
		}
		return false
	})
	return foundProposals
}

func (pf *ProposalFundStore) AddFunds(proposalId ProposalID, fundingAddress keys.Address, amount *ProposalAmount) error {
	key := storage.StoreKey(string(proposalId) + storage.DB_PREFIX + fundingAddress.String())
	amt, err := pf.get(key)
	if err != nil {
		return errors.Wrap(err, errorGettingRecord)
	}
	return pf.set(key, *amt.Plus(amount))
}

func (pf *ProposalFundStore) DeleteFunds(proposalId ProposalID, fundingAddress keys.Address) (bool, error) {
	key := storage.StoreKey(string(proposalId) + storage.DB_PREFIX + fundingAddress.String())
	_, err := pf.get(key)
	if err != nil {
		return false, errors.Wrap(err, errorGettingRecord)
	}
	ok, err := pf.delete(key)
	if err != nil {
		return false, errors.Wrap(err, errorDeletingRecord)
	}
	return ok, nil
}
