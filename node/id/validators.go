package id

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"github.com/Oneledger/protocol/node/log"
	"github.com/Oneledger/protocol/node/serial"
	"github.com/tendermint/tendermint/abci/types"
	"math/big"
	"strings"
)

type Validators struct {
	Signers           []types.SigningValidator
	Byzantines        []types.Evidence
	Approved          []Identity
	SelectedValidator Identity
	NewValidators     []types.Validator
}

func init() {
	serial.Register(Validators{})
}

func NewValidatorList() *Validators {
	return &Validators{}
}

func (list *Validators) Set(app interface{}, validators []types.SigningValidator, badValidators []types.Evidence, hash []byte) {
	if validators == nil {
		return
	}
	list.Signers = validators
	list.Byzantines = badValidators
	list.Approved = list.FindApproved(app)
	if hash != nil {
		list.SelectedValidator = list.FindSelectedValidator(app, hash)
	}
	list.NewValidators = make([]types.Validator, 0)
}

func (list *Validators) FindSelectedValidator(app interface{}, hash []byte) Identity {
	countBigInt := big.NewInt(int64(len(list.Approved)))
	hashBigInt := new(big.Int).SetBytes(hash)
	indexBigInt := new(big.Int)
	indexBigInt = indexBigInt.Mod(hashBigInt, countBigInt)
	var indexInt64, _ = new(big.Int).SetString(indexBigInt.String(), 10)
	index := int(indexInt64.Int64())
	selectedValidator := list.Approved[index]
	return selectedValidator
}

func (list *Validators) FindApproved(app interface{}) []Identity {
	var approvedIdentities []Identity
	for _, entry := range list.Signers {
		entryIsBad := IsByzantine(entry.Validator, list.Byzantines)
		if !entryIsBad {
			formatted := hex.EncodeToString(entry.Validator.Address)
			identities := GetIdentities(app)
			identity := identities.FindTendermint(formatted)
			approvedIdentities = append(approvedIdentities, identity)
		}
	}
	return approvedIdentities
}

func IsByzantine(validator types.Validator, badValidators []types.Evidence) (result bool) {
	for _, entry := range badValidators {
		if bytes.Equal(validator.Address, entry.Validator.Address) {
			return true
		}
	}
	return false
}

func (list Validators) IsValidAccountKey(key AccountKey, index int) bool {
	if index >= len(list.Approved) || index < 0 {
		return false
	}

	id := list.Approved[index]
	if bytes.Equal(id.AccountKey.Bytes(), key.Bytes()) {
		return true
	}

	return false
}

func (list *Validators) AddNewValidator(address string, pubkey string, power int64) bool {
	var validator types.Validator
	validator.Address = []byte(address)
	buffer, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		log.Debug("Failed to decode the pubkey", "pubkey", pubkey)
		return false
	}
	buffer = buffer[4:]
	key, err := ImportBytesKey(buffer, ED25519)
	if err != nil {
		log.Debug("Failed to convert the pubkey", "buffer", buffer)
		return false
	}
	tpubkey := types.PubKey{
		Type: strings.ToLower(ED25519.String()),
		Data: key.Bytes(),
	}
	validator.PubKey = tpubkey
	validator.Power = power

	list.NewValidators = append(list.NewValidators, validator)
	return true
}
