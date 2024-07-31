// Code generated by MockGen. DO NOT EDIT.
// Source: x/symSlash/types/expected_keepers.go

// Package testutil is a generated GoMock package.
package testutil

import (
	context "context"
	reflect "reflect"

	address "cosmossdk.io/core/address"
	math "cosmossdk.io/math"
	types "cosmossdk.io/x/symStaking/types"
	types0 "github.com/cosmos/cosmos-sdk/types"
	gomock "github.com/golang/mock/gomock"
)

// MockAccountKeeper is a mock of AccountKeeper interface.
type MockAccountKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockAccountKeeperMockRecorder
}

// MockAccountKeeperMockRecorder is the mock recorder for MockAccountKeeper.
type MockAccountKeeperMockRecorder struct {
	mock *MockAccountKeeper
}

// NewMockAccountKeeper creates a new mock instance.
func NewMockAccountKeeper(ctrl *gomock.Controller) *MockAccountKeeper {
	mock := &MockAccountKeeper{ctrl: ctrl}
	mock.recorder = &MockAccountKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountKeeper) EXPECT() *MockAccountKeeperMockRecorder {
	return m.recorder
}

// AddressCodec mocks base method.
func (m *MockAccountKeeper) AddressCodec() address.Codec {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddressCodec")
	ret0, _ := ret[0].(address.Codec)
	return ret0
}

// AddressCodec indicates an expected call of AddressCodec.
func (mr *MockAccountKeeperMockRecorder) AddressCodec() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddressCodec", reflect.TypeOf((*MockAccountKeeper)(nil).AddressCodec))
}

// GetAccount mocks base method.
func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr types0.AccAddress) types0.AccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", ctx, addr)
	ret0, _ := ret[0].(types0.AccountI)
	return ret0
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockAccountKeeperMockRecorder) GetAccount(ctx, addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockAccountKeeper)(nil).GetAccount), ctx, addr)
}

// MockBankKeeper is a mock of BankKeeper interface.
type MockBankKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockBankKeeperMockRecorder
}

// MockBankKeeperMockRecorder is the mock recorder for MockBankKeeper.
type MockBankKeeperMockRecorder struct {
	mock *MockBankKeeper
}

// NewMockBankKeeper creates a new mock instance.
func NewMockBankKeeper(ctrl *gomock.Controller) *MockBankKeeper {
	mock := &MockBankKeeper{ctrl: ctrl}
	mock.recorder = &MockBankKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBankKeeper) EXPECT() *MockBankKeeperMockRecorder {
	return m.recorder
}

// GetAllBalances mocks base method.
func (m *MockBankKeeper) GetAllBalances(ctx context.Context, addr types0.AccAddress) types0.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllBalances", ctx, addr)
	ret0, _ := ret[0].(types0.Coins)
	return ret0
}

// GetAllBalances indicates an expected call of GetAllBalances.
func (mr *MockBankKeeperMockRecorder) GetAllBalances(ctx, addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllBalances", reflect.TypeOf((*MockBankKeeper)(nil).GetAllBalances), ctx, addr)
}

// GetBalance mocks base method.
func (m *MockBankKeeper) GetBalance(ctx context.Context, addr types0.AccAddress, denom string) types0.Coin {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", ctx, addr, denom)
	ret0, _ := ret[0].(types0.Coin)
	return ret0
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockBankKeeperMockRecorder) GetBalance(ctx, addr, denom interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockBankKeeper)(nil).GetBalance), ctx, addr, denom)
}

// LockedCoins mocks base method.
func (m *MockBankKeeper) LockedCoins(ctx context.Context, addr types0.AccAddress) types0.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LockedCoins", ctx, addr)
	ret0, _ := ret[0].(types0.Coins)
	return ret0
}

// LockedCoins indicates an expected call of LockedCoins.
func (mr *MockBankKeeperMockRecorder) LockedCoins(ctx, addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LockedCoins", reflect.TypeOf((*MockBankKeeper)(nil).LockedCoins), ctx, addr)
}

// SpendableCoins mocks base method.
func (m *MockBankKeeper) SpendableCoins(ctx context.Context, addr types0.AccAddress) types0.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SpendableCoins", ctx, addr)
	ret0, _ := ret[0].(types0.Coins)
	return ret0
}

// SpendableCoins indicates an expected call of SpendableCoins.
func (mr *MockBankKeeperMockRecorder) SpendableCoins(ctx, addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SpendableCoins", reflect.TypeOf((*MockBankKeeper)(nil).SpendableCoins), ctx, addr)
}

// MockStakingKeeper is a mock of StakingKeeper interface.
type MockStakingKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockStakingKeeperMockRecorder
}

// MockStakingKeeperMockRecorder is the mock recorder for MockStakingKeeper.
type MockStakingKeeperMockRecorder struct {
	mock *MockStakingKeeper
}

// NewMockStakingKeeper creates a new mock instance.
func NewMockStakingKeeper(ctrl *gomock.Controller) *MockStakingKeeper {
	mock := &MockStakingKeeper{ctrl: ctrl}
	mock.recorder = &MockStakingKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStakingKeeper) EXPECT() *MockStakingKeeperMockRecorder {
	return m.recorder
}

// ConsensusAddressCodec mocks base method.
func (m *MockStakingKeeper) ConsensusAddressCodec() address.Codec {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConsensusAddressCodec")
	ret0, _ := ret[0].(address.Codec)
	return ret0
}

// ConsensusAddressCodec indicates an expected call of ConsensusAddressCodec.
func (mr *MockStakingKeeperMockRecorder) ConsensusAddressCodec() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConsensusAddressCodec", reflect.TypeOf((*MockStakingKeeper)(nil).ConsensusAddressCodec))
}

// GetAllValidators mocks base method.
func (m *MockStakingKeeper) GetAllValidators(ctx context.Context) ([]types.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllValidators", ctx)
	ret0, _ := ret[0].([]types.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllValidators indicates an expected call of GetAllValidators.
func (mr *MockStakingKeeperMockRecorder) GetAllValidators(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllValidators", reflect.TypeOf((*MockStakingKeeper)(nil).GetAllValidators), ctx)
}

// IsValidatorJailed mocks base method.
func (m *MockStakingKeeper) IsValidatorJailed(ctx context.Context, addr types0.ConsAddress) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsValidatorJailed", ctx, addr)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsValidatorJailed indicates an expected call of IsValidatorJailed.
func (mr *MockStakingKeeperMockRecorder) IsValidatorJailed(ctx, addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsValidatorJailed", reflect.TypeOf((*MockStakingKeeper)(nil).IsValidatorJailed), ctx, addr)
}

// IterateValidators mocks base method.
func (m *MockStakingKeeper) IterateValidators(arg0 context.Context, arg1 func(int64, types.Validator) bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IterateValidators", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// IterateValidators indicates an expected call of IterateValidators.
func (mr *MockStakingKeeperMockRecorder) IterateValidators(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IterateValidators", reflect.TypeOf((*MockStakingKeeper)(nil).IterateValidators), arg0, arg1)
}

// Jail mocks base method.
func (m *MockStakingKeeper) Jail(arg0 context.Context, arg1 types0.ConsAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Jail", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Jail indicates an expected call of Jail.
func (mr *MockStakingKeeperMockRecorder) Jail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Jail", reflect.TypeOf((*MockStakingKeeper)(nil).Jail), arg0, arg1)
}

// MaxValidators mocks base method.
func (m *MockStakingKeeper) MaxValidators(arg0 context.Context) (uint32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MaxValidators", arg0)
	ret0, _ := ret[0].(uint32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MaxValidators indicates an expected call of MaxValidators.
func (mr *MockStakingKeeperMockRecorder) MaxValidators(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxValidators", reflect.TypeOf((*MockStakingKeeper)(nil).MaxValidators), arg0)
}

// Unjail mocks base method.
func (m *MockStakingKeeper) Unjail(arg0 context.Context, arg1 types0.ConsAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unjail", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unjail indicates an expected call of Unjail.
func (mr *MockStakingKeeperMockRecorder) Unjail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unjail", reflect.TypeOf((*MockStakingKeeper)(nil).Unjail), arg0, arg1)
}

// Validator mocks base method.
func (m *MockStakingKeeper) Validator(arg0 context.Context, arg1 types0.ValAddress) (types.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validator", arg0, arg1)
	ret0, _ := ret[0].(types.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Validator indicates an expected call of Validator.
func (mr *MockStakingKeeperMockRecorder) Validator(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validator", reflect.TypeOf((*MockStakingKeeper)(nil).Validator), arg0, arg1)
}

// ValidatorAddressCodec mocks base method.
func (m *MockStakingKeeper) ValidatorAddressCodec() address.Codec {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorAddressCodec")
	ret0, _ := ret[0].(address.Codec)
	return ret0
}

// ValidatorAddressCodec indicates an expected call of ValidatorAddressCodec.
func (mr *MockStakingKeeperMockRecorder) ValidatorAddressCodec() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorAddressCodec", reflect.TypeOf((*MockStakingKeeper)(nil).ValidatorAddressCodec))
}

// ValidatorByConsAddr mocks base method.
func (m *MockStakingKeeper) ValidatorByConsAddr(arg0 context.Context, arg1 types0.ConsAddress) (types.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorByConsAddr", arg0, arg1)
	ret0, _ := ret[0].(types.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidatorByConsAddr indicates an expected call of ValidatorByConsAddr.
func (mr *MockStakingKeeperMockRecorder) ValidatorByConsAddr(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorByConsAddr", reflect.TypeOf((*MockStakingKeeper)(nil).ValidatorByConsAddr), arg0, arg1)
}

// MockStakingHooks is a mock of StakingHooks interface.
type MockStakingHooks struct {
	ctrl     *gomock.Controller
	recorder *MockStakingHooksMockRecorder
}

// MockStakingHooksMockRecorder is the mock recorder for MockStakingHooks.
type MockStakingHooksMockRecorder struct {
	mock *MockStakingHooks
}

// NewMockStakingHooks creates a new mock instance.
func NewMockStakingHooks(ctrl *gomock.Controller) *MockStakingHooks {
	mock := &MockStakingHooks{ctrl: ctrl}
	mock.recorder = &MockStakingHooksMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStakingHooks) EXPECT() *MockStakingHooksMockRecorder {
	return m.recorder
}

// AfterValidatorBeginUnbonding mocks base method.
func (m *MockStakingHooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr types0.ConsAddress, valAddr types0.ValAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterValidatorBeginUnbonding", ctx, consAddr, valAddr)
	ret0, _ := ret[0].(error)
	return ret0
}

// AfterValidatorBeginUnbonding indicates an expected call of AfterValidatorBeginUnbonding.
func (mr *MockStakingHooksMockRecorder) AfterValidatorBeginUnbonding(ctx, consAddr, valAddr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterValidatorBeginUnbonding", reflect.TypeOf((*MockStakingHooks)(nil).AfterValidatorBeginUnbonding), ctx, consAddr, valAddr)
}

// AfterValidatorBonded mocks base method.
func (m *MockStakingHooks) AfterValidatorBonded(ctx context.Context, consAddr types0.ConsAddress, valAddr types0.ValAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterValidatorBonded", ctx, consAddr, valAddr)
	ret0, _ := ret[0].(error)
	return ret0
}

// AfterValidatorBonded indicates an expected call of AfterValidatorBonded.
func (mr *MockStakingHooksMockRecorder) AfterValidatorBonded(ctx, consAddr, valAddr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterValidatorBonded", reflect.TypeOf((*MockStakingHooks)(nil).AfterValidatorBonded), ctx, consAddr, valAddr)
}

// AfterValidatorCreated mocks base method.
func (m *MockStakingHooks) AfterValidatorCreated(ctx context.Context, valAddr types0.ValAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterValidatorCreated", ctx, valAddr)
	ret0, _ := ret[0].(error)
	return ret0
}

// AfterValidatorCreated indicates an expected call of AfterValidatorCreated.
func (mr *MockStakingHooksMockRecorder) AfterValidatorCreated(ctx, valAddr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterValidatorCreated", reflect.TypeOf((*MockStakingHooks)(nil).AfterValidatorCreated), ctx, valAddr)
}

// AfterValidatorRemoved mocks base method.
func (m *MockStakingHooks) AfterValidatorRemoved(ctx context.Context, consAddr types0.ConsAddress, valAddr types0.ValAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterValidatorRemoved", ctx, consAddr, valAddr)
	ret0, _ := ret[0].(error)
	return ret0
}

// AfterValidatorRemoved indicates an expected call of AfterValidatorRemoved.
func (mr *MockStakingHooksMockRecorder) AfterValidatorRemoved(ctx, consAddr, valAddr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterValidatorRemoved", reflect.TypeOf((*MockStakingHooks)(nil).AfterValidatorRemoved), ctx, consAddr, valAddr)
}

// BeforeValidatorModified mocks base method.
func (m *MockStakingHooks) BeforeValidatorModified(ctx context.Context, valAddr types0.ValAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeforeValidatorModified", ctx, valAddr)
	ret0, _ := ret[0].(error)
	return ret0
}

// BeforeValidatorModified indicates an expected call of BeforeValidatorModified.
func (mr *MockStakingHooksMockRecorder) BeforeValidatorModified(ctx, valAddr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeforeValidatorModified", reflect.TypeOf((*MockStakingHooks)(nil).BeforeValidatorModified), ctx, valAddr)
}

// BeforeValidatorSlashed mocks base method.
func (m *MockStakingHooks) BeforeValidatorSlashed(ctx context.Context, valAddr types0.ValAddress, fraction math.LegacyDec) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeforeValidatorSlashed", ctx, valAddr, fraction)
	ret0, _ := ret[0].(error)
	return ret0
}

// BeforeValidatorSlashed indicates an expected call of BeforeValidatorSlashed.
func (mr *MockStakingHooksMockRecorder) BeforeValidatorSlashed(ctx, valAddr, fraction interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeforeValidatorSlashed", reflect.TypeOf((*MockStakingHooks)(nil).BeforeValidatorSlashed), ctx, valAddr, fraction)
}
