## ⚠️ Disclaimer: It is NOT intended for a production use.

**[Symbiotic Protocol](https://symbiotic.fi) is an extremely flexible and permissionless shared security system.**

This repository contains a simple network middleware example implementation.

## Middleware example

SimpleMiddleware contracts serve as an example of how typical middleware can be implemented. Next, we describe the functionality of the contracts and possible improvements.

In the contracts, there is an admin role that can call important functions of the middleware. However, as mentioned, the code is an example, and in real production networks, it can be implemented differently.

The Example network uses epochs to divide time into consecutive blocks of equal size. Each epoch has its own validator set, including keys and stakes, which the network captures at the start of every epoch.

Note that this middleware implements the logic of VALSET and SLASH VERIFIER from the network abstraction section, as well as other functionalities.

### Register operators and vaults

Any operator needs a stake to be eligible to work in the network. The source of the stakes is a vault. Therefore, any network should clarify which operators and vaults it accepts. This process is called opt-in mechanics and is described in the official docs. In Simple middleware contracts, there is a `registerVault` method that performs opt-in to the vault. It also checks and validates the epoch size of the vault and the type of slasher in the vault. Any vault can also be paused and unpaused on the middleware side. However, this is not related to the Symbiotic functionality and is implemented only for example purposes.

Since operators use different nodes with different keys, SimpleMiddleware can register and deregister keys of the operator. There are `registerOperator` and `updateOperatorKey` methods that allow this.

## Get Validator Set

To get the validator set, there is a `getValidatorSet` method that iterates over all registered operators in the middleware and calculates their stakes at the given epoch. This function can be called to get the actual validator set for the current epoch or to identify validators elected in previous epochs. Note that this method can be implemented in different ways, including:

1. Storing different keys for the same operator
2. Checking the minimal stake amount of the operator
3. Splitting the stake of the operator across its keys
4. Introducing a schedule for the operators
5. Updating VALSET when cross-slashing incidents happen

As mentioned, this method is just an example of how a validator set can be implemented.

## Slashing

If an operator misbehaves, it must be slashed. To do this in the middleware, use the `slash(epoch, operator, amount)` method. This method iterates over all the operator’s vaults and slashes the operator proportionally. The epoch argument is related to the network’s epoch because an operator can be slashed not immediately but after some time.

Slashing can be implemented in other ways, including:

1. Custom slashing rules excluding proportional slashing
2. Caching slash incidents and sending them all at once
3. Additional verification of slashing requests, for example, when the network uses fraud proofs
4. Using an aggregated key from the majority of operators instead of a specific role to send slash requests

This method is essentially the SLASH VERIFIER from the network abstraction section but with basic verification.

## Usage

### Env

Create `.env` file using a template:

```
ETH_RPC_URL=
ETHERSCAN_API_KEY=
```

\* ETH_RPC_URL is optional.

\* ETHERSCAN_API_KEY is optional.

### Build

```shell
forge build
```

### Test

```shell
forge test
```

### Format

```shell
forge fmt
```
