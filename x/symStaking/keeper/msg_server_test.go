package keeper_test

import (
	"testing"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/symStaking/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PKS     = simtestutil.CreateTestPubKeys(3)
	Addr    = sdk.AccAddress(PKS[0].Address())
	ValAddr = sdk.ValAddress(Addr)
)

func (s *KeeperTestSuite) execExpectCalls() {
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
}

func (s *KeeperTestSuite) TestMsgCreateValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk1 := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk1)

	pubkey, err := codectypes.NewAnyWithValue(pk1)
	require.NoError(err)

	var ed25519pk cryptotypes.PubKey = &ed25519.PubKey{Key: []byte{1, 2, 3, 4, 5, 6}}
	pubkeyInvalidLen, err := codectypes.NewAnyWithValue(ed25519pk)
	require.NoError(err)

	invalidPk, _ := secp256r1.GenPrivKey()
	invalidPubkey, err := codectypes.NewAnyWithValue(invalidPk.PubKey())
	require.NoError(err)

	badKey := secp256k1.GenPrivKey()
	badPubKey, err := codectypes.NewAnyWithValue(&secp256k1.PubKey{Key: badKey.PubKey().Bytes()[:len(badKey.PubKey().Bytes())-1]})
	require.NoError(err)

	testCases := []struct {
		name        string
		input       *types.MsgCreateValidator
		expErr      bool
		expErrMsg   string
		expPanic    bool
		expPanicMsg string
	}{
		{
			name: "empty description",
			input: &types.MsgCreateValidator{
				Description: types.Description{},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
				Pubkey:           pubkey,
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "invalid validator address",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				ValidatorAddress: s.addressToString([]byte("invalid")),
				Pubkey:           pubkey,
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty validator pubkey",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
				Pubkey:           nil,
			},
			expErr:    true,
			expErrMsg: "empty validator public key",
		},
		{
			name: "validator pubkey len is invalid",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker: "NewValidator",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
				Pubkey:           pubkeyInvalidLen,
			},
			expErr:    true,
			expErrMsg: "consensus pubkey len is invalid",
		},
		{
			name: "invalid pubkey type",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker:         "NewValidator",
					Identity:        "xyz",
					Website:         "xyz.com",
					SecurityContact: "xyz@gmail.com",
					Details:         "details",
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
				Pubkey:           invalidPubkey,
			},
			expErr:    true,
			expErrMsg: "got: secp256r1, expected: [ed25519 secp256k1]: validator pubkey type is not supported",
		},
		{
			name: "invalid pubkey length",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker:  "NewValidator",
					Identity: "xyz",
					Website:  "xyz.com",
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
				Pubkey:           badPubKey,
			},
			expPanic:    true,
			expPanicMsg: "length of pubkey is incorrect",
		},
		{
			name: "valid msg",
			input: &types.MsgCreateValidator{
				Description: types.Description{
					Moniker:         "NewValidator",
					Identity:        "xyz",
					Website:         "xyz.com",
					SecurityContact: "xyz@gmail.com",
					Details:         "details",
				},
				Commission: types.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
				Pubkey:           pubkey,
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.expPanic {
				require.PanicsWithValue(tc.expPanicMsg, func() {
					_, _ = msgServer.CreateValidator(ctx, tc.input)
				})
				return
			}

			_, err := msgServer.CreateValidator(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgEditValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	// create new context with updated block time
	newCtx := ctx.WithHeaderInfo(header.Info{Time: ctx.HeaderInfo().Time.AddDate(0, 0, 1)})
	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	msg, err := types.NewMsgCreateValidator(s.valAddressToString(ValAddr), pk, types.Description{Moniker: "NewVal"}, comm)
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	testCases := []struct {
		name      string
		ctx       sdk.Context
		input     *types.MsgEditValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress: s.addressToString([]byte("invalid")),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty description",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description:      types.Description{},
				ValidatorAddress: s.valAddressToString(ValAddr),
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "validator does not exist",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress: s.valAddressToString([]byte("val")),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "valid msg",
			ctx:  newCtx,
			input: &types.MsgEditValidator{
				Description: types.Description{
					Moniker:         "TestValidator",
					Identity:        "abc",
					Website:         "abc.com",
					SecurityContact: "abc@gmail.com",
					Details:         "newDetails",
				},
				ValidatorAddress: s.valAddressToString(ValAddr),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.EditValidator(tc.ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}
