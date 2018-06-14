/*
	Copyright 2017 - 2018 OneLedger
*/

package action

import "github.com/Oneledger/protocol/node/data"

type CommandType int

// Set of possible commands that can be driven from a transaction
const (
	NOOP CommandType = iota
	SUBMIT_TRANSACTION
	INITIATE
	PARTICIPATE
	REDEEM
	REFUND
	EXTRACTSECRET
	AUDITCONTRACT
	WAIT_FOR_CHAIN
)

// A command to execute again a chain, needs to be polymorphic
type Command struct {
	Function CommandType
	Chain    data.ChainType
	Data     map[string]string
}

func (command Command) Execute() bool {
	switch command.Function {
	case NOOP:
		return Noop(command.Chain, command.Data)

	case SUBMIT_TRANSACTION:
		return SubmitTransaction(command.Chain, command.Data)

	case INITIATE:
		return Initiate(command.Chain, command.Data)

	case PARTICIPATE:
		return Participate(command.Chain, command.Data)

	case REDEEM:
		return Redeem(command.Chain, command.Data)

	case REFUND:
		return Refund(command.Chain, command.Data)

	case EXTRACTSECRET:
		return ExtractSecret(command.Chain, command.Data)

	case AUDITCONTRACT:
		return AuditContract(command.Chain, command.Data)

	case WAIT_FOR_CHAIN:
		return WaitForChain(command.Chain, command.Data)
	}

	return true
}

type Commands []Command

func (commands Commands) Count() int {
	return len(commands)
}
