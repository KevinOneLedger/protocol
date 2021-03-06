package governance

import (
	"fmt"

	"github.com/Oneledger/protocol/data/keys"
	"github.com/Oneledger/protocol/storage"
	"github.com/pkg/errors"
)

type ProposalVoteStore struct {
	prefix []byte
	store  *storage.State
}

func NewProposalVoteStore(prefix string, state *storage.State) *ProposalVoteStore {
	return &ProposalVoteStore{
		prefix: storage.Prefix(prefix),
		store:  state,
	}
}

func (pvs *ProposalVoteStore) WithState(state *storage.State) *ProposalVoteStore {
	pvs.store = state
	return pvs
}

// Setup an initial voting validator to proposalID
func (pvs *ProposalVoteStore) Setup(proposalID ProposalID, vote *ProposalVote) error {
	info := fmt.Sprintf("Vote Setup: proposalID= %v, %v", proposalID, vote.String())

	if proposalID == "" {
		logger.Errorf("%v, empty proposalID", info)
		return ErrVoteSetupValidatorFailed
	}

	vote.Opinion = OPIN_UNKNOWN // initialize as OPIN_UNKNOWN
	key := GetKey(pvs.prefix, proposalID, vote)
	value := vote.Bytes()
	err := pvs.store.Set(key, value)
	if err != nil {
		logger.Errorf("%v, storage failure", info)
		return ErrVoteSetupValidatorFailed
	}
	logger.Info(info)

	return nil
}

// Update a validator's voting opinion to proposalID
func (pvs *ProposalVoteStore) Update(proposalID ProposalID, vote *ProposalVote) error {
	info := fmt.Sprintf("Vote Update: proposalID= %v, %v", proposalID, vote.String())

	key := GetKey(pvs.prefix, proposalID, vote)
	exist := pvs.store.Exists(key)
	if !exist {
		logger.Errorf("%v, can't participate in voting", info)
		return ErrVoteUpdateVoteFailed
	}

	value := vote.Bytes()
	err := pvs.store.Set(key, value)
	if err != nil {
		logger.Errorf("%v, storage failure", info)
		return ErrVoteUpdateVoteFailed
	}
	logger.Info(info)

	return nil
}

// Delete all voting records under a proposalID
func (pvs *ProposalVoteStore) Delete(proposalID ProposalID) error {
	info := fmt.Sprintf("Vote Delete: proposalID= %v", proposalID)

	succeed := true
	pvs.IterateByID(proposalID, func(key []byte, value []byte) bool {
		_, err := pvs.store.Delete(key)
		if err != nil {
			logger.Errorf("%v, failed to delete vote, key= %v", info, string(key))
			succeed = false
		}
		return false
	})

	if !succeed {
		logger.Errorf("%v, delete failed", info)
		return ErrVoteDeleteVoteRecordsFailed
	}
	logger.Info(info)

	return nil
}

// Check and see if a proposal has passed
func (pvs *ProposalVoteStore) IsPassed(proposalID ProposalID, passPercent int64) (bool, error) {
	info := fmt.Sprintf("Vote IsPassed: proposalID= %v", proposalID)

	_, votes, err := pvs.GetVotesByID(proposalID)
	if err != nil {
		logger.Errorf("%v, getVotesByID failed", info)
		return false, ErrVoteCheckVoteResultFailed
	}

	// Accumulates power of each opinion
	totalPower := int64(0)
	eachPower := make([]int64, 4)
	for _, vote := range votes {
		totalPower += vote.Power
		eachPower[vote.Opinion] += vote.Power
	}

	// Excludes validators that give up voting in percent calculation
	totalPower -= eachPower[OPIN_GIVEUP]

	// Calculate actual percentage
	percent := 0.0
	passed := false
	if totalPower > 0 {
		percent = float64(eachPower[OPIN_POSITIVE]) / float64(totalPower)
	}
	if percent >= float64(passPercent)/100.0 {
		passed = true
	}
	logger.Infof("%v, passed= %v", info, passed)

	return passed, nil
}

// Iterate voting records by proposalID
func (pvs *ProposalVoteStore) IterateByID(proposalID ProposalID, fn func(key []byte, value []byte) bool) (stopped bool) {
	prefix := append(pvs.prefix, (proposalID + storage.DB_PREFIX)...)
	return pvs.store.IterateRange(
		prefix,
		storage.Rangefix(string(prefix)),
		true,
		func(key, value []byte) bool {
			return fn(key, value)
		},
	)
}

// get voting votes by proposalID
func (pvs *ProposalVoteStore) GetVotesByID(proposalID ProposalID) ([]keys.Address, []*ProposalVote, error) {
	info := fmt.Sprintf("Vote getVotesByID: proposalID= %v", proposalID)

	succeed := true
	addrs := make([]keys.Address, 0)
	votes := make([]*ProposalVote, 0)
	pvs.IterateByID(proposalID, func(key []byte, value []byte) bool {
		vote, err := (&ProposalVote{}).FromBytes(value)
		if err != nil {
			logger.Errorf("%v, key= %v, deserialize proposal vote failed", info, key)
			succeed = false
			return true
		}
		votes = append(votes, vote)
		prefix_len := len(append(pvs.prefix, (proposalID + storage.DB_PREFIX)...))
		addr := key[prefix_len:]
		addrs = append(addrs, addr)
		return false
	})

	if !succeed {
		errMsg := fmt.Sprintf("%v, operation failed", info)
		logger.Error(errMsg)
		return nil, nil, errors.New(errMsg)
	}
	// Caused by invalid/deleted proposalID
	if len(votes) == 0 {
		errMsg := fmt.Sprintf("%v, no votes records found", info)
		logger.Error(errMsg)
		return nil, nil, errors.New(errMsg)
	}

	return addrs, votes, nil
}

func GetKey(prefix []byte, proposalID ProposalID, vote *ProposalVote) []byte {
	return storage.StoreKey(string(prefix) + string(proposalID) + storage.DB_PREFIX + string(vote.Validator))
}
