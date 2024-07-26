package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the staking Querier
const (
	QueryValidators     = "validators"
	QueryValidator      = "validator"
	QueryParameters     = "parameters"
	QueryHistoricalInfo = "historicalInfo"
)

// defines the params for the following queries:
// - 'custom/staking/validator'
// - 'custom/staking/validatorDelegations'
// - 'custom/staking/validatorUnbondingDelegations'
type QueryValidatorParams struct {
	ValidatorAddr sdk.ValAddress
	Page, Limit   int
}

func NewQueryValidatorParams(validatorAddr sdk.ValAddress, page, limit int) QueryValidatorParams {
	return QueryValidatorParams{
		ValidatorAddr: validatorAddr,
		Page:          page,
		Limit:         limit,
	}
}

// QueryValidatorsParams defines the params for the following queries:
// - 'custom/staking/validators'
type QueryValidatorsParams struct {
	Page, Limit int
	Status      string
}

func NewQueryValidatorsParams(page, limit int, status string) QueryValidatorsParams {
	return QueryValidatorsParams{page, limit, status}
}
