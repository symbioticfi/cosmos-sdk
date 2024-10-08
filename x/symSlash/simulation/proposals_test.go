package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/symSlash/simulation"
	"cosmossdk.io/x/symSlash/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)
	ac := codectestutil.CodecOptions{}.GetAddressCodec()

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	assert.Assert(t, len(weightedProposalMsgs) == 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	assert.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	assert.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg, err := w0.MsgSimulatorFn()(r, accounts, ac)
	assert.NilError(t, err)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	moduleAddr, err := ac.BytesToString(address.Module("symGov"))
	assert.NilError(t, err)

	assert.Equal(t, moduleAddr, msgUpdateParams.Authority)
	assert.Equal(t, int64(905), msgUpdateParams.Params.SignedBlocksWindow)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(7, 2), msgUpdateParams.Params.MinSignedPerWindow)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(60, 2), msgUpdateParams.Params.SlashFractionDoubleSign)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(89, 2), msgUpdateParams.Params.SlashFractionDowntime)
	assert.Equal(t, 3313479009*time.Second, msgUpdateParams.Params.DowntimeJailDuration)
}
