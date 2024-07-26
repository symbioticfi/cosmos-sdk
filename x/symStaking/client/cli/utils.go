package cli

import (
	"encoding/json"
	"errors"
	"os"

	"cosmossdk.io/math"
	"cosmossdk.io/x/symStaking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// validator struct to define the fields of the validator
type validator struct {
	PubKey          cryptotypes.PubKey
	Moniker         string
	Identity        string
	Website         string
	Security        string
	Details         string
	CommissionRates types.CommissionRates
}

func parseAndValidateValidatorJSON(cdc codec.Codec, path string) (validator, error) {
	type internalVal struct {
		PubKey              json.RawMessage `json:"pubkey"`
		Moniker             string          `json:"moniker"`
		Identity            string          `json:"identity,omitempty"`
		Website             string          `json:"website,omitempty"`
		Security            string          `json:"security,omitempty"`
		Details             string          `json:"details,omitempty"`
		CommissionRate      string          `json:"commission-rate"`
		CommissionMaxRate   string          `json:"commission-max-rate"`
		CommissionMaxChange string          `json:"commission-max-change-rate"`
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return validator{}, err
	}

	var v internalVal
	err = json.Unmarshal(contents, &v)
	if err != nil {
		return validator{}, err
	}

	if v.PubKey == nil {
		return validator{}, errors.New("must specify the JSON encoded pubkey")
	}
	var pk cryptotypes.PubKey
	if err := cdc.UnmarshalInterfaceJSON(v.PubKey, &pk); err != nil {
		return validator{}, err
	}

	if v.Moniker == "" {
		return validator{}, errors.New("must specify the moniker name")
	}

	commissionRates, err := buildCommissionRates(v.CommissionRate, v.CommissionMaxRate, v.CommissionMaxChange)
	if err != nil {
		return validator{}, err
	}

	return validator{
		PubKey:          pk,
		Moniker:         v.Moniker,
		Identity:        v.Identity,
		Website:         v.Website,
		Security:        v.Security,
		Details:         v.Details,
		CommissionRates: commissionRates,
	}, nil
}

func buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr string) (commission types.CommissionRates, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.New("must specify all validator commission parameters")
	}

	rate, err := math.LegacyNewDecFromStr(rateStr)
	if err != nil {
		return commission, err
	}

	maxRate, err := math.LegacyNewDecFromStr(maxRateStr)
	if err != nil {
		return commission, err
	}

	maxChangeRate, err := math.LegacyNewDecFromStr(maxChangeRateStr)
	if err != nil {
		return commission, err
	}

	commission = types.NewCommissionRates(rate, maxRate, maxChangeRate)

	return commission, nil
}
