package symGov_test

import (
	"cosmossdk.io/math"
	"cosmossdk.io/x/symGov/types/v1beta1"
	stakingtypes "cosmossdk.io/x/symStaking/types"
)

var (
	TestProposal        = v1beta1.NewTextProposal("Test", "description")
	TestDescription     = stakingtypes.NewDescription("T", "E", "S", "T", "Z")
	TestCommissionRates = stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
)
