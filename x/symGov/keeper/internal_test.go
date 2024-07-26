package keeper

import "cosmossdk.io/x/symGov/types"

// UnsafeSetHooks updates the symGov keeper's hooks, overriding any potential
// pre-existing hooks.
// WARNING: this function should only be used in tests.
func UnsafeSetHooks(k *Keeper, h types.GovHooks) {
	k.hooks = h
}
