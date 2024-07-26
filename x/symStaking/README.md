---
sidebar_position: 1
---

# `x/symStaking`

## Abstract

This paper specifies the Staking module of the Cosmos SDK that was first
described in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper)
in June 2016.

The module enables Cosmos SDK-based blockchain to support an advanced
Proof-of-Stake (PoS) system. In this system, holders of the native staking token of
the chain can become validators and can delegate tokens to validators,
ultimately determining the effective validator set for the system.

This module is used in the Cosmos Hub, the first Hub in the Cosmos
network.

## Symbiotic stake

Set **$BEACON_API_URL** and **$ETH_API_URL** env variables or use default rpc **ONLY FOR TESTING**.

## Contents

* [State](#state)
    * [LastTotalPower](#lasttotalpower)
    * [UnbondingID](#unbondingid)
    * [Params](#params)
    * [Validator](#validator)
    * [Queues](#queues)
    * [HistoricalInfo](#historicalinfo)
* [State Transitions](#state-transitions)
    * [Validators](#validators)
    <!-- * [Slashing](#slashing) -->
* [Messages](#messages)
    * [MsgCreateValidator](#msgcreatevalidator)
    * [MsgEditValidator](#msgeditvalidator)
    * [MsgUpdateParams](#msgupdateparams)
* [Begin-Block](#begin-block)
    * [Historical Info Tracking](#historical-info-tracking)
* [End-Block](#end-block)
    * [Validator Set Changes](#validator-set-changes)
    * [Queues](#queues-1)
* [Hooks](#hooks)
* [Events](#events)
    * [EndBlocker](#endblocker)
    * [Msg's](#msgs)
* [Parameters](#parameters)
* [Client](#client)
    * [CLI](#cli)
    * [gRPC](#grpc)
    * [REST](#rest)

## State

### LastTotalPower

LastTotalPower tracks the total amounts of bonded tokens recorded during the previous end block.
Store entries prefixed with "Last" must remain unchanged until EndBlock.

* LastTotalPower: `0x12 -> ProtocolBuffer(math.Int)`

### UnbondingID

UnbondingID stores the ID of the latest unbonding operation. It enables creating unique IDs for unbonding operations, i.e., UnbondingID is incremented every time a new unbonding operation (validator unbonding, unbonding delegation, redelegation) is initiated.

* UnbondingID: `0x37 -> uint64`

### Params

The staking module stores its params in state with the prefix of `0x51`,
it can be updated with governance or the address with authority.

* Params: `0x51 | ProtocolBuffer(Params)`

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/staking.proto#L310-L333
```

### Validator

Validators can have one of three statuses

* `Unbonded`: The validator is not in the active set. They cannot sign blocks and do not earn
  rewards. They can receive delegations.
* `Bonded`: Once the validator receives sufficient bonded tokens they automatically join the
  active set during [`EndBlock`](#validator-set-changes) and their status is updated to `Bonded`.
  They are signing blocks and receiving rewards. They can receive further delegations.
  They can be slashed for misbehavior. Delegators to this validator who unbond their delegation
  must wait the duration of the UnbondingTime, a chain-specific param, during which time
  they are still slashable for offences of the source validator if those offences were committed
  during the period of time that the tokens were bonded.
* `Unbonding`: When a validator leaves the active set, either by choice or due to slashing, jailing or
  tombstoning, an unbonding of all their delegations begins. All delegations must then wait the UnbondingTime
  before their tokens are moved to their accounts from the `BondedPool`.

:::warning
Tombstoning is permanent, once tombstoned a validator's consensus key can not be reused within the chain where the tombstoning happened.
:::

Validators objects should be primarily stored and accessed by the
`OperatorAddr`, an SDK validator address for the operator of the validator. Two
additional indices are maintained per validator object in order to fulfill
required lookups for slashing and validator-set updates. A third special index
(`LastValidatorPower`) is also maintained which however remains constant
throughout each block, unlike the first two indices which mirror the validator
records within a block.

* Validators: `0x21 | OperatorAddrLen (1 byte) | OperatorAddr -> ProtocolBuffer(validator)`
* ValidatorsByConsAddr: `0x22 | ConsAddrLen (1 byte) | ConsAddr -> OperatorAddr`
* ValidatorsByPower: `0x23 | BigEndian(ConsensusPower) | OperatorAddrLen (1 byte) | OperatorAddr -> OperatorAddr`
* LastValidatorsPower: `0x11 | OperatorAddrLen (1 byte) | OperatorAddr -> ProtocolBuffer(ConsensusPower)`
* ValidatorsByUnbondingID: `0x38 | UnbondingID ->  0x21 | OperatorAddrLen (1 byte) | OperatorAddr`

`Validators` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future. Delegators can refer to the immutable operator of the validator, without
concern for the changing public key.

`ValidatorsByUnbondingID` is an additional index that enables lookups for
 validators by the unbonding IDs corresponding to their current unbonding.

`ValidatorByConsAddr` is an additional index that enables lookups for slashing.
When CometBFT reports evidence, it provides the validator address, so this
map is needed to find the operator. Note that the `ConsAddr` corresponds to the
address which can be derived from the validator's `ConsPubKey`.

`ValidatorsByPower` is an additional index that provides a sorted list of
potential validators to quickly determine the current active set. Here
ConsensusPower is validator.Tokens/10^6 by default. Note that all validators
where `Jailed` is true are not stored within this index.

`LastValidatorsPower` is a special index that provides a historical list of the
last-block's bonded validators. This index remains constant during a block but
is updated during the validator set update process which takes place in [`EndBlock`](#end-block).

Each validator's state is stored in a `Validator` struct:

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/staking.proto#L82-L138
```

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/staking.proto#L26-L80
```

### Queues

All queue objects are sorted by timestamp. The time used within any queue is
firstly converted to UTC, rounded to the nearest nanosecond then sorted. The sortable time format
used is a slight modification of the RFC3339Nano and uses the format string
`"2006-01-02T15:04:05.000000000"`. Notably this format:

* right pads all zeros
* drops the time zone info (we already use UTC)

In all cases, the stored timestamp represents the maturation time of the queue
element.

#### ValidatorQueue

For the purpose of tracking progress of unbonding validators the validator
queue is kept.

* ValidatorQueueTime: `0x43 | format(time) -> []sdk.ValAddress`

The stored object by each key is an array of validator operator addresses from
which the validator object can be accessed. Typically it is expected that only
a single validator record will be associated with a given timestamp however it is possible
that multiple validators exist in the queue at the same location.

### HistoricalInfo

HistoricalInfo objects are stored and pruned at each block such that the staking keeper persists
the `n` most recent historical info defined by staking module parameter: `HistoricalEntries`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/staking.proto#L17-L24
```

At each BeginBlock, the staking keeper will persist the current Header and the Validators that committed
the current block in a `HistoricalInfo` object. The Validators are sorted on their address to ensure that
they are in a deterministic order.
The oldest HistoricalEntries will be pruned to ensure that there only exist the parameter-defined number of
historical entries.

## State Transitions

### Validators

State transitions in validators are performed on every [`EndBlock`](#validator-set-changes)
in order to check for changes in the active `ValidatorSet`.

A validator can be `Unbonded`, `Unbonding` or `Bonded`. `Unbonded`
and `Unbonding` are collectively called `Not Bonded`. A validator can move
directly between all the states, except for from `Bonded` to `Unbonded`.

#### Not bonded to Bonded

The following transition occurs when a validator's ranking in the `ValidatorPowerIndex` surpasses
that of the `LastValidator`.

* set `validator.Status` to `Bonded`
* delete the existing record from `ValidatorByPowerIndex`
* add a new updated record to the `ValidatorByPowerIndex`
* update the `Validator` object for this validator
* if it exists, delete any `ValidatorQueue` record for this validator

#### Bonded to Unbonding

When a validator begins the unbonding process the following operations occur:

* set `validator.Status` to `Unbonding`
* delete the existing record from `ValidatorByPowerIndex`
* add a new updated record to the `ValidatorByPowerIndex`
* update the `Validator` object for this validator
* insert a new record into the `ValidatorQueue` for this validator

#### Unbonding to Unbonded

A validator moves from unbonding to unbonded when the `ValidatorQueue` object
moves from bonded to unbonded

* update the `Validator` object for this validator
* set `validator.Status` to `Unbonded`

#### Jail/Unjail

when a validator is jailed it is effectively removed from the CometBFT set.
this process may be also be reversed. the following operations occur:

* set `Validator.Jailed` and update object
* if jailed delete record from `ValidatorByPowerIndex`
* if unjailed add record to `ValidatorByPowerIndex`

Jailed validators are not present in any of the following stores:

* the power store (from consensus power to address)

<!-- ### Slashing

#### Slash Validator

When a Validator is slashed, the following occurs:

* The total `slashAmount` is calculated as the `slashFactor` (a chain parameter) \* `TokensFromConsensusPower`,
  the total number of tokens bonded to the validator at the time of the infraction.
* Every unbonding delegation and pseudo-unbonding redelegation such that the infraction occurred before the unbonding or
  redelegation began from the validator are slashed by the `slashFactor` percentage of the initialBalance.
* Each amount slashed from redelegations and unbonding delegations is subtracted from the
  total slash amount.
* The `remaingSlashAmount` is then slashed from the validator's tokens in the `BondedPool` or
  `NonBondedPool` depending on the validator's status. This reduces the total supply of tokens.

In the case of a slash due to any infraction that requires evidence to submitted (for example double-sign), the slash
occurs at the block where the evidence is included, not at the block where the infraction occurred.
Put otherwise, validators are not slashed retroactively, only when they are caught. -->

## Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](#state) section.

### MsgCreateValidator

A validator is created using the `MsgCreateValidator` message.
The validator must be created with an initial delegation from the operator.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L20-L21
```

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L50-L73
```

This message is expected to fail if:

* another validator with this operator address is already registered
* another validator with this pubkey is already registered
* the initial self-delegation tokens are of a denom not specified as the bonding denom
* the commission parameters are faulty, namely:
    * `MaxRate` is either > 1 or < 0
    * the initial `Rate` is either negative or > `MaxRate`
    * the initial `MaxChangeRate` is either negative or > `MaxRate`
* the description fields are too large

This message creates and stores the `Validator` object at appropriate indexes.

### MsgEditValidator

The `Description`, `CommissionRate` of a validator can be updated using the
`MsgEditValidator` message.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L23-L24
```

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L78-L97
```

This message is expected to fail if:

* the initial `CommissionRate` is either negative or > `MaxRate`
* the `CommissionRate` has already been updated within the previous 24 hours
* the `CommissionRate` is > `MaxChangeRate`
* the description fields are too large

This message stores the updated `Validator` object.

### MsgUpdateParams

The `MsgUpdateParams` update the staking module parameters.
The params are updated through a governance proposal where the signer is the gov module account address.
When the `MinCommissionRate` is updated, all validators with a lower (max) commission rate than `MinCommissionRate` will be updated to `MinCommissionRate`.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/proto/cosmos/staking/v1beta1/tx.proto#L182-L195
```

The message handling can fail if:

* signer is not the authority defined in the staking keeper (usually the gov module account).

## Begin-Block

Each abci begin block call, the historical info will get stored and pruned
according to the `HistoricalEntries` parameter.

### Historical Info Tracking

If the `HistoricalEntries` parameter is 0, then the `BeginBlock` performs a no-op.

Otherwise, the latest historical info is stored under the key `historicalInfoKey|height`, while any entries older than `height - HistoricalEntries` is deleted.
In most cases, this results in a single entry being pruned per block.
However, if the parameter `HistoricalEntries` has changed to a lower value there will be multiple entries in the store that must be pruned.

## End-Block

Each abci end block call, the operations to update queues and validator set
changes are specified to execute.

### Validator Set Changes

The staking validator set is updated during this process by state transitions
that run at the end of every block. As a part of this process any updated
validators are also returned back to CometBFT for inclusion in the CometBFT
validator set which is responsible for validating CometBFT messages at the
consensus layer. Operations are as following:

* the new validator set is taken as the top `params.MaxValidators` number of
  validators retrieved from the `ValidatorsByPower` index
* the previous validator set is compared with the new validator set:
    * missing validators begin unbonding
    * new validators are instantly bonded

In all cases, any validators leaving or entering the bonded validator set or
changing balances and staying within the bonded validator set incur an update
message reporting their new consensus power which is passed back to CometBFT.

The `LastTotalPower` and `LastValidatorsPower` hold the state of the total power
and validator power from the end of the last block, and are used to check for
changes that have occurred in `ValidatorsByPower` and the total new power, which
is calculated during `EndBlock`.

### Queues

Within staking, certain state-transitions are not instantaneous but take place
over a duration of time (typically the unbonding period). When these
transitions are mature certain operations must take place in order to complete
the state operation. This is achieved through the use of queues which are
checked/processed at the end of each block.

#### Unbonding Validators

When a validator is kicked out of the bonded validator set (either through
being jailed, or not having sufficient bonded tokens) it begins the unbonding
process along with all its delegations begin unbonding (while still being
delegated to this validator). At this point the validator is said to be an
"unbonding validator", whereby it will mature to become an "unbonded validator"
after the unbonding period has passed.

Each block the validator queue is to be checked for mature unbonding validators
(namely with a completion time <= current time and completion height <= current
block height). At this point any mature validators which do not have any
delegations remaining are deleted from state. For all other mature unbonding
validators that still have remaining delegations, the `validator.Status` is
switched from `types.Unbonding` to
`types.Unbonded`.

Unbonding operations can be put on hold by external modules via the `PutUnbondingOnHold(unbondingId)` method.
 As a result, an unbonding operation (e.g., an unbonding delegation) that is on hold, cannot complete
 even if it reaches maturity. For an unbonding operation with `unbondingId` to eventually complete
 (after it reaches maturity), every call to `PutUnbondingOnHold(unbondingId)` must be matched
 by a call to `UnbondingCanComplete(unbondingId)`.

## Hooks

Other modules may register operations to execute when a certain event has
occurred within staking.  These events can be registered to execute either
right `Before` or `After` the staking event (as per the hook name). The
following hooks can registered with staking:

* `AfterValidatorCreated(Context, ValAddress) error`
    * called when a validator is created
* `BeforeValidatorModified(Context, ValAddress) error`
    * called when a validator's state is changed
* `AfterValidatorRemoved(Context, ConsAddress, ValAddress) error`
    * called when a validator is deleted
* `AfterValidatorBonded(Context, ConsAddress, ValAddress) error`
    * called when a validator is bonded
* `AfterValidatorBeginUnbonding(Context, ConsAddress, ValAddress) error`
    * called when a validator begins unbonding


## Events

The staking module emits the following events:

## Msg's

### MsgCreateValidator

| Type             | Attribute Key | Attribute Value    |
| ---------------- | ------------- | ------------------ |
| create_validator | validator     | {validatorAddress} |
| create_validator | amount        | {delegationAmount} |
| message          | module        | staking            |
| message          | action        | create_validator   |
| message          | sender        | {senderAddress}    |

### MsgEditValidator

| Type           | Attribute Key       | Attribute Value     |
| -------------- | ------------------- | ------------------- |
| edit_validator | commission_rate     | {commissionRate}    |
| message        | module              | staking             |
| message        | action              | edit_validator      |
| message        | sender              | {senderAddress}     |

## Parameters

The staking module contains the following parameters:

| Key                    | Type             | Example                |
|-------------------     |------------------|------------------------|
| UnbondingTime          | string (time ns) | "259200000000000"      |
| MaxValidators          | uint16           | 100                    |
| KeyMaxEntries          | uint16           | 7                      |
| HistoricalEntries      | uint16           | 3                      |
| BondDenom              | string           | "stake"                |
| MinCommissionRate      | string           | "0.000000000000000000" |

:::warning
Manually updating the `MinCommissionRate` parameter will not affect the commission rate of the existing validators. It will only affect the commission rate of the new validators. Update the parameter with `MsgUpdateParams` to affect the commission rate of the existing validators as well.
:::

## Client

### CLI

A user can query and interact with the `staking` module using the CLI.

#### Query

The `query` commands allows users to query `staking` state.

```bash
symd query staking --help
```

##### historical-info

The `historical-info` command allows users to query historical information at given height.

Usage:

```bash
symd query staking historical-info [height] [flags]
```

Example:

```bash
symd query staking historical-info 10
```

Example Output:

```bash
header:
  app_hash: Lbx8cXpI868wz8sgp4qPYVrlaKjevR5WP/IjUxwp3oo=
  chain_id: testnet
  consensus_hash: BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=
  data_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  evidence_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  height: "10"
  last_block_id:
    hash: RFbkpu6pWfSThXxKKl6EZVDnBSm16+U0l0xVjTX08Fk=
    part_set_header:
      hash: vpIvXD4rxD5GM4MXGz0Sad9I7//iVYLzZsEU4BVgWIU=
      total: 1
  last_commit_hash: Ne4uXyx4QtNp4Zx89kf9UK7oG9QVbdB6e7ZwZkhy8K0=
  last_results_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  next_validators_hash: nGBgKeWBjoxeKFti00CxHsnULORgKY4LiuQwBuUrhCs=
  proposer_address: mMEP2c2IRPLr99LedSRtBg9eONM=
  time: "2021-10-01T06:00:49.785790894Z"
  validators_hash: nGBgKeWBjoxeKFti00CxHsnULORgKY4LiuQwBuUrhCs=
  version:
    app: "0"
    block: "11"
valset:
- commission:
    commission_rates:
      max_change_rate: "0.010000000000000000"
      max_rate: "0.200000000000000000"
      rate: "0.100000000000000000"
    update_time: "2021-10-01T05:52:50.380144238Z"
  consensus_pubkey:
    '@type': /cosmos.crypto.ed25519.PubKey
    key: Auxs3865HpB/EfssYOzfqNhEJjzys2Fo6jD5B8tPgC8=
  description:
    details: ""
    identity: ""
    moniker: myvalidator
    security_contact: ""
    website: ""
  jailed: false
  operator_address: cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc
  status: BOND_STATUS_BONDED
  tokens: "10000000"
  unbonding_height: "0"
  unbonding_time: "1970-01-01T00:00:00Z"
```

##### params

The `params` command allows users to query values set as staking parameters.

Usage:

```bash
symd query staking params [flags]
```

Example:

```bash
symd query staking params
```

Example Output:

```bash
bond_denom: stake
historical_entries: 10000
max_entries: 7
max_validators: 50
unbonding_time: 1814400s
```

##### validator

The `validator` command allows users to query details about an individual validator.

Usage:

```bash
symd query staking validator [validator-addr] [flags]
```

Example:

```bash
symd query staking validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
commission:
  commission_rates:
    max_change_rate: "0.020000000000000000"
    max_rate: "0.200000000000000000"
    rate: "0.050000000000000000"
  update_time: "2021-10-01T19:24:52.663191049Z"
consensus_pubkey:
  '@type': /cosmos.crypto.ed25519.PubKey
  key: sIiexdJdYWn27+7iUHQJDnkp63gq/rzUq1Y+fxoGjXc=
description:
  details: Witval is the validator arm from Vitwit. Vitwit is into software consulting
    and services business since 2015. We are working closely with Cosmos ecosystem
    since 2018. We are also building tools for the ecosystem, Aneka is our explorer
    for the cosmos ecosystem.
  identity: 51468B615127273A
  moniker: Witval
  security_contact: ""
  website: ""
jailed: false
operator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
status: BOND_STATUS_BONDED
tokens: "32948270000"
unbonding_height: "0"
unbonding_time: "1970-01-01T00:00:00Z"
```

##### validators

The `validators` command allows users to query details about all validators on a network.

Usage:

```bash
symd query staking validators [flags]
```

Example:

```bash
symd query staking validators
```

Example Output:

```bash
pagination:
  next_key: FPTi7TKAjN63QqZh+BaXn6gBmD5/
  total: "0"
validators:
commission:
  commission_rates:
    max_change_rate: "0.020000000000000000"
    max_rate: "0.200000000000000000"
    rate: "0.050000000000000000"
  update_time: "2021-10-01T19:24:52.663191049Z"
consensus_pubkey:
  '@type': /cosmos.crypto.ed25519.PubKey
  key: sIiexdJdYWn27+7iUHQJDnkp63gq/rzUq1Y+fxoGjXc=
description:
    details: Witval is the validator arm from Vitwit. Vitwit is into software consulting
      and services business since 2015. We are working closely with Cosmos ecosystem
      since 2018. We are also building tools for the ecosystem, Aneka is our explorer
      for the cosmos ecosystem.
    identity: 51468B615127273A
    moniker: Witval
    security_contact: ""
    website: ""
  jailed: false
  operator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
  status: BOND_STATUS_BONDED
  tokens: "32948270000"
  unbonding_height: "0"
  unbonding_time: "1970-01-01T00:00:00Z"
- commission:
    commission_rates:
      max_change_rate: "0.100000000000000000"
      max_rate: "0.200000000000000000"
      rate: "0.050000000000000000"
    update_time: "2021-10-04T18:02:21.446645619Z"
  consensus_pubkey:
    '@type': /cosmos.crypto.ed25519.PubKey
    key: GDNpuKDmCg9GnhnsiU4fCWktuGUemjNfvpCZiqoRIYA=
  description:
    details: Noderunners is a professional validator in POS networks. We have a huge
      node running experience, reliable soft and hardware. Our commissions are always
      low, our support to delegators is always full. Stake with us and start receiving
      your Cosmos rewards now!
    identity: 812E82D12FEA3493
    moniker: Noderunners
    security_contact: info@noderunners.biz
    website: http://noderunners.biz
  jailed: false
  operator_address: cosmosvaloper1q5ku90atkhktze83j9xjaks2p7uruag5zp6wt7
  status: BOND_STATUS_BONDED
  tokens: "559343421"
  unbonding_height: "0"
  unbonding_time: "1970-01-01T00:00:00Z"
```

#### Transactions

The `tx` commands allows users to interact with the `staking` module.

```bash
symd tx staking --help
```

##### create-validator

The command `create-validator` allows users to create new validator initialized with a self-delegation to it.

Usage:

```bash
symd tx staking create-validator [path/to/validator.json] [flags]
```

Example:

```bash
symd tx staking create-validator /path/to/validator.json \
  --chain-id="name_of_chain_id" \
  --gas="auto" \
  --gas-adjustment="1.2" \
  --gas-prices="0.025stake" \
  --from=mykey
```

where `validator.json` contains:

```json
{
  "pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"BnbwFpeONLqvWqJb3qaUbL5aoIcW3fSuAp9nT3z5f20="},
  "moniker": "my-moniker",
  "website": "https://myweb.site",
  "security": "security-contact@gmail.com",
  "details": "description of your validator",
  "commission-rate": "0.10",
  "commission-max-rate": "0.20",
  "commission-max-change-rate": "0.01",
}
```

and pubkey can be obtained by using `simd tendermint show-validator` command.

##### edit-validator

The command `edit-validator` allows users to edit an existing validator account.

Usage:

```bash
symd tx staking edit-validator [flags]
```

Example:

```bash
symd tx staking edit-validator --moniker "new_moniker_name" --website "new_website_url" --from mykey
```

### gRPC

A user can query the `staking` module using gRPC endpoints.

#### Validators

The `Validators` endpoint queries all validators that match the given status.

```bash
cosmos.staking.v1beta1.Query/Validators
```

Example:

```bash
grpcurl -plaintext localhost:9090 cosmos.staking.v1beta1.Query/Validators
```

Example Output:

```bash
{
  "validators": [
    {
      "operatorAddress": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
      "consensusPubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"Auxs3865HpB/EfssYOzfqNhEJjzys2Fo6jD5B8tPgC8="},
      "status": "BOND_STATUS_BONDED",
      "tokens": "10000000",
      "delegatorShares": "10000000000000000000000000",
      "description": {
        "moniker": "myvalidator"
      },
      "unbondingTime": "1970-01-01T00:00:00Z",
      "commission": {
        "commissionRates": {
          "rate": "100000000000000000",
          "maxRate": "200000000000000000",
          "maxChangeRate": "10000000000000000"
        },
        "updateTime": "2021-10-01T05:52:50.380144238Z"
      },
      "minSelfDelegation": "1"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

#### Validator

The `Validator` endpoint queries validator information for given validator address.

```bash
cosmos.staking.v1beta1.Query/Validator
```

Example:

```bash
grpcurl -plaintext -d '{"validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/Validator
```

Example Output:

```bash
{
  "validator": {
    "operatorAddress": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
    "consensusPubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"Auxs3865HpB/EfssYOzfqNhEJjzys2Fo6jD5B8tPgC8="},
    "status": "BOND_STATUS_BONDED",
    "tokens": "10000000",
    "delegatorShares": "10000000000000000000000000",
    "description": {
      "moniker": "myvalidator"
    },
    "unbondingTime": "1970-01-01T00:00:00Z",
    "commission": {
      "commissionRates": {
        "rate": "100000000000000000",
        "maxRate": "200000000000000000",
        "maxChangeRate": "10000000000000000"
      },
      "updateTime": "2021-10-01T05:52:50.380144238Z"
    },
    "minSelfDelegation": "1"
  }
}
```

#### ValidatorDelegations

The `ValidatorDelegations` endpoint queries delegate information for given validator.

```bash
cosmos.staking.v1beta1.Query/ValidatorDelegations
```

Example:

```bash
grpcurl -plaintext -d '{"validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/ValidatorDelegations
```

Example Output:

```bash
{
  "delegationResponses": [
    {
      "delegation": {
        "delegatorAddress": "cosmos1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgy3ua5t",
        "validatorAddress": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
        "shares": "10000000000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "10000000"
      }
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

#### ValidatorUnbondingDelegations

The `ValidatorUnbondingDelegations` endpoint queries delegate information for given validator.

```bash
cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations
```

Example:

```bash
grpcurl -plaintext -d '{"validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1z3pzzw84d6xn00pw9dy3yapqypfde7vg6965fy",
      "validator_address": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
      "entries": [
        {
          "creation_height": "25325",
          "completion_time": "2021-10-31T09:24:36.797320636Z",
          "initial_balance": "20000000",
          "balance": "20000000"
        }
      ]
    },
    {
      "delegator_address": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
      "validator_address": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
      "entries": [
        {
          "creation_height": "13100",
          "completion_time": "2021-10-30T12:53:02.272266791Z",
          "initial_balance": "1000000",
          "balance": "1000000"
        }
      ]
    },
  ],
  "pagination": {
    "next_key": null,
    "total": "8"
  }
}
```

#### Delegation

The `Delegation` endpoint queries delegate information for given validator delegator pair.

```bash
cosmos.staking.v1beta1.Query/Delegation
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77", validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/Delegation
```

Example Output:

```bash
{
  "delegation_response":
  {
    "delegation":
      {
        "delegator_address":"cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
        "validator_address":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
        "shares":"25083119936.000000000000000000"
      },
    "balance":
      {
        "denom":"stake",
        "amount":"25083119936"
      }
  }
}
```

#### UnbondingDelegation

The `UnbondingDelegation` endpoint queries unbonding information for given validator delegator.

```bash
cosmos.staking.v1beta1.Query/UnbondingDelegation
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77", validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/UnbondingDelegation
```

Example Output:

```bash
{
  "unbond": {
    "delegator_address": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
    "validator_address": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
    "entries": [
      {
        "creation_height": "136984",
        "completion_time": "2021-11-08T05:38:47.505593891Z",
        "initial_balance": "400000000",
        "balance": "400000000"
      },
      {
        "creation_height": "137005",
        "completion_time": "2021-11-08T05:40:53.526196312Z",
        "initial_balance": "385000000",
        "balance": "385000000"
      }
    ]
  }
}
```

#### DelegatorDelegations

The `DelegatorDelegations` endpoint queries all delegations of a given delegator address.

```bash
cosmos.staking.v1beta1.Query/DelegatorDelegations
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorDelegations
```

Example Output:

```bash
{
  "delegation_responses": [
    {"delegation":{"delegator_address":"cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77","validator_address":"cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8","shares":"25083339023.000000000000000000"},"balance":{"denom":"stake","amount":"25083339023"}}
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

#### DelegatorUnbondingDelegations

The `DelegatorUnbondingDelegations` endpoint queries all unbonding delegations of a given delegator address.

```bash
cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
      "validator_address": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9uxyejze",
      "entries": [
        {
          "creation_height": "136984",
          "completion_time": "2021-11-08T05:38:47.505593891Z",
          "initial_balance": "400000000",
          "balance": "400000000"
        },
        {
          "creation_height": "137005",
          "completion_time": "2021-11-08T05:40:53.526196312Z",
          "initial_balance": "385000000",
          "balance": "385000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

#### Redelegations

The `Redelegations` endpoint queries redelegations of given address.

```bash
cosmos.staking.v1beta1.Query/Redelegations
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1ld5p7hn43yuh8ht28gm9pfjgj2fctujp2tgwvf", "src_validator_addr" : "cosmosvaloper1j7euyj85fv2jugejrktj540emh9353ltgppc3g", "dst_validator_addr" : "cosmosvaloper1yy3tnegzmkdcm7czzcy3flw5z0zyr9vkkxrfse"}' \
localhost:9090 cosmos.staking.v1beta1.Query/Redelegations
```

Example Output:

```bash
{
  "redelegation_responses": [
    {
      "redelegation": {
        "delegator_address": "cosmos1ld5p7hn43yuh8ht28gm9pfjgj2fctujp2tgwvf",
        "validator_src_address": "cosmosvaloper1j7euyj85fv2jugejrktj540emh9353ltgppc3g",
        "validator_dst_address": "cosmosvaloper1yy3tnegzmkdcm7czzcy3flw5z0zyr9vkkxrfse",
        "entries": null
      },
      "entries": [
        {
          "redelegation_entry": {
            "creation_height": 135932,
            "completion_time": "2021-11-08T03:52:55.299147901Z",
            "initial_balance": "2900000",
            "shares_dst": "2900000.000000000000000000"
          },
          "balance": "2900000"
        }
      ]
    }
  ],
  "pagination": null
}
```

#### DelegatorValidators

The `DelegatorValidators` endpoint queries all validators information for given delegator.

```bash
cosmos.staking.v1beta1.Query/DelegatorValidators
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1ld5p7hn43yuh8ht28gm9pfjgj2fctujp2tgwvf"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorValidators
```

Example Output:

```bash
{
  "validators": [
    {
      "operator_address": "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "UPwHWxH1zHJWGOa/m6JB3f5YjHMvPQPkVbDqqi+U7Uw="
      },
      "jailed": false,
      "status": "BOND_STATUS_BONDED",
      "tokens": "347260647559",
      "delegator_shares": "347260647559.000000000000000000",
      "description": {
        "moniker": "BouBouNode",
        "identity": "",
        "website": "https://boubounode.com",
        "security_contact": "",
        "details": "AI-based Validator. #1 AI Validator on Game of Stakes. Fairly priced. Don't trust (humans), verify. Made with BouBou love."
      },
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z",
      "commission": {
        "commission_rates": {
          "rate": "0.061000000000000000",
          "max_rate": "0.300000000000000000",
          "max_change_rate": "0.150000000000000000"
        },
        "update_time": "2021-10-01T15:00:00Z"
      },
      "min_self_delegation": "1"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

#### DelegatorValidator

The `DelegatorValidator` endpoint queries validator information for given delegator validator

```bash
cosmos.staking.v1beta1.Query/DelegatorValidator
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1eh5mwu044gd5ntkkc2xgfg8247mgc56f3n8rr7", "validator_addr": "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorValidator
```

Example Output:

```bash
{
  "validator": {
    "operator_address": "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8",
    "consensus_pubkey": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "UPwHWxH1zHJWGOa/m6JB3f5YjHMvPQPkVbDqqi+U7Uw="
    },
    "jailed": false,
    "status": "BOND_STATUS_BONDED",
    "tokens": "347262754841",
    "delegator_shares": "347262754841.000000000000000000",
    "description": {
      "moniker": "BouBouNode",
      "identity": "",
      "website": "https://boubounode.com",
      "security_contact": "",
      "details": "AI-based Validator. #1 AI Validator on Game of Stakes. Fairly priced. Don't trust (humans), verify. Made with BouBou love."
    },
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z",
    "commission": {
      "commission_rates": {
        "rate": "0.061000000000000000",
        "max_rate": "0.300000000000000000",
        "max_change_rate": "0.150000000000000000"
      },
      "update_time": "2021-10-01T15:00:00Z"
    },
    "min_self_delegation": "1"
  }
}
```

#### HistoricalInfo

```bash
cosmos.staking.v1beta1.Query/HistoricalInfo
```

Example:

```bash
grpcurl -plaintext -d '{"height" : 1}' localhost:9090 cosmos.staking.v1beta1.Query/HistoricalInfo
```

Example Output:

```bash
{
  "hist": {
    "header": {
      "version": {
        "block": "11",
        "app": "0"
      },
      "chain_id": "simd-1",
      "height": "140142",
      "time": "2021-10-11T10:56:29.720079569Z",
      "last_block_id": {
        "hash": "9gri/4LLJUBFqioQ3NzZIP9/7YHR9QqaM6B2aJNQA7o=",
        "part_set_header": {
          "total": 1,
          "hash": "Hk1+C864uQkl9+I6Zn7IurBZBKUevqlVtU7VqaZl1tc="
        }
      },
      "last_commit_hash": "VxrcS27GtvGruS3I9+AlpT7udxIT1F0OrRklrVFSSKc=",
      "data_hash": "80BjOrqNYUOkTnmgWyz9AQ8n7SoEmPVi4QmAe8RbQBY=",
      "validators_hash": "95W49n2hw8RWpr1GPTAO5MSPi6w6Wjr3JjjS7AjpBho=",
      "next_validators_hash": "95W49n2hw8RWpr1GPTAO5MSPi6w6Wjr3JjjS7AjpBho=",
      "consensus_hash": "BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=",
      "app_hash": "ZZaxnSY3E6Ex5Bvkm+RigYCK82g8SSUL53NymPITeOE=",
      "last_results_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
      "evidence_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
      "proposer_address": "aH6dO428B+ItuoqPq70efFHrSMY="
    },
  "valset": [
      {
        "operator_address": "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcqcnylw",
        "consensus_pubkey": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "/O7BtNW0pafwfvomgR4ZnfldwPXiFfJs9mHg3gwfv5Q="
        },
        "jailed": false,
        "status": "BOND_STATUS_BONDED",
        "tokens": "1426045203613",
        "delegator_shares": "1426045203613.000000000000000000",
        "description": {
          "moniker": "SG-1",
          "identity": "48608633F99D1B60",
          "website": "https://sg-1.online",
          "security_contact": "",
          "details": "SG-1 - your favorite validator on Witval. We offer 100% Soft Slash protection."
        },
        "unbonding_height": "0",
        "unbonding_time": "1970-01-01T00:00:00Z",
        "commission": {
          "commission_rates": {
            "rate": "0.037500000000000000",
            "max_rate": "0.200000000000000000",
            "max_change_rate": "0.030000000000000000"
          },
          "update_time": "2021-10-01T15:00:00Z"
        },
        "min_self_delegation": "1"
      }
    ]
  }
}

```

#### Pool

The `Pool` endpoint queries the pool information.

```bash
cosmos.staking.v1beta1.Query/Pool
```

Example:

```bash
grpcurl -plaintext -d localhost:9090 cosmos.staking.v1beta1.Query/Pool
```

Example Output:

```bash
{
  "pool": {
    "not_bonded_tokens": "369054400189",
    "bonded_tokens": "15657192425623"
  }
}
```

#### Params

The `Params` endpoint queries the pool information.

```bash
cosmos.staking.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext localhost:9090 cosmos.staking.v1beta1.Query/Params
```

Example Output:

```bash
{
  "params": {
    "unbondingTime": "1814400s",
    "maxValidators": 100,
    "maxEntries": 7,
    "historicalEntries": 10000,
    "bondDenom": "stake"
  }
}
```

### REST

A user can query the `staking` module using REST endpoints.

#### DelegatorDelegations

The `DelegtaorDelegations` REST endpoint queries all delegations of a given delegator address.

```bash
/cosmos/staking/v1beta1/delegations/{delegatorAddr}
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/delegations/cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5" -H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_responses": [
    {
      "delegation": {
        "delegator_address": "cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5",
        "validator_address": "cosmosvaloper1quqxfrxkycr0uzt4yk0d57tcq3zk7srm7sm6r8",
        "shares": "256250000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "256250000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5",
        "validator_address": "cosmosvaloper194v8uwee2fvs2s8fa5k7j03ktwc87h5ym39jfv",
        "shares": "255150000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "255150000"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

#### Redelegations

The `Redelegations` REST endpoint queries redelegations of given address.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/redelegations
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1thfntksw0d35n2tkr0k8v54fr8wxtxwxl2c56e/redelegations?srcValidatorAddr=cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qx4cuznf&dstValidatorAddr=cosmosvaloper1vq8tw77kp8lvxq9u3c8eeln9zymn68rng8pgt4" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "redelegation_responses": [
    {
      "redelegation": {
        "delegator_address": "cosmos1thfntksw0d35n2tkr0k8v54fr8wxtxwxl2c56e",
        "validator_src_address": "cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qx4cuznf",
        "validator_dst_address": "cosmosvaloper1vq8tw77kp8lvxq9u3c8eeln9zymn68rng8pgt4",
        "entries": null
      },
      "entries": [
        {
          "redelegation_entry": {
            "creation_height": 151523,
            "completion_time": "2021-11-09T06:03:25.640682116Z",
            "initial_balance": "200000000",
            "shares_dst": "200000000.000000000000000000"
          },
          "balance": "200000000"
        }
      ]
    }
  ],
  "pagination": null
}
```

#### DelegatorUnbondingDelegations

The `DelegatorUnbondingDelegations` REST endpoint queries all unbonding delegations of a given delegator address.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/unbonding_delegations
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1nxv42u3lv642q0fuzu2qmrku27zgut3n3z7lll/unbonding_delegations" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1nxv42u3lv642q0fuzu2qmrku27zgut3n3z7lll",
      "validator_address": "cosmosvaloper1e7mvqlz50ch6gw4yjfemsc069wfre4qwmw53kq",
      "entries": [
        {
          "creation_height": "2442278",
          "completion_time": "2021-10-12T10:59:03.797335857Z",
          "initial_balance": "50000000000",
          "balance": "50000000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

#### DelegatorValidators

The `DelegatorValidators` REST endpoint queries all validators information for given delegator address.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1xwazl8ftks4gn00y5x3c47auquc62ssune9ppv/validators" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "validators": [
    {
      "operator_address": "cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuvynw64",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "5v4n3px3PkfNnKflSgepDnsMQR1hiNXnqOC11Y72/PQ="
      },
      "jailed": false,
      "status": "BOND_STATUS_BONDED",
      "tokens": "21592843799",
      "delegator_shares": "21592843799.000000000000000000",
      "description": {
        "moniker": "jabbey",
        "identity": "",
        "website": "https://twitter.com/JoeAbbey",
        "security_contact": "",
        "details": "just another dad in the cosmos"
      },
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z",
      "commission": {
        "commission_rates": {
          "rate": "0.100000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.100000000000000000"
        },
        "update_time": "2021-10-09T19:03:54.984821705Z"
      },
      "min_self_delegation": "1"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

#### DelegatorValidator

The `DelegatorValidator` REST endpoint queries validator information for given delegator validator pair.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators/{validatorAddr}
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1xwazl8ftks4gn00y5x3c47auquc62ssune9ppv/validators/cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuvynw64" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "validator": {
    "operator_address": "cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuvynw64",
    "consensus_pubkey": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "5v4n3px3PkfNnKflSgepDnsMQR1hiNXnqOC11Y72/PQ="
    },
    "jailed": false,
    "status": "BOND_STATUS_BONDED",
    "tokens": "21592843799",
    "delegator_shares": "21592843799.000000000000000000",
    "description": {
      "moniker": "jabbey",
      "identity": "",
      "website": "https://twitter.com/JoeAbbey",
      "security_contact": "",
      "details": "just another dad in the cosmos"
    },
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z",
    "commission": {
      "commission_rates": {
        "rate": "0.100000000000000000",
        "max_rate": "0.200000000000000000",
        "max_change_rate": "0.100000000000000000"
      },
      "update_time": "2021-10-09T19:03:54.984821705Z"
    },
    "min_self_delegation": "1"
  }
}
```

#### HistoricalInfo

The `HistoricalInfo` REST endpoint queries the historical information for given height.

```bash
/cosmos/staking/v1beta1/historical_info/{height}
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/historical_info/153332" -H  "accept: application/json"
```

Example Output:

```bash
{
  "hist": {
    "header": {
      "version": {
        "block": "11",
        "app": "0"
      },
      "chain_id": "cosmos-1",
      "height": "153332",
      "time": "2021-10-12T09:05:35.062230221Z",
      "last_block_id": {
        "hash": "NX8HevR5khb7H6NGKva+jVz7cyf0skF1CrcY9A0s+d8=",
        "part_set_header": {
          "total": 1,
          "hash": "zLQ2FiKM5tooL3BInt+VVfgzjlBXfq0Hc8Iux/xrhdg="
        }
      },
      "last_commit_hash": "P6IJrK8vSqU3dGEyRHnAFocoDGja0bn9euLuy09s350=",
      "data_hash": "eUd+6acHWrNXYju8Js449RJ99lOYOs16KpqQl4SMrEM=",
      "validators_hash": "mB4pravvMsJKgi+g8aYdSeNlt0kPjnRFyvtAQtaxcfw=",
      "next_validators_hash": "mB4pravvMsJKgi+g8aYdSeNlt0kPjnRFyvtAQtaxcfw=",
      "consensus_hash": "BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=",
      "app_hash": "fuELArKRK+CptnZ8tu54h6xEleSWenHNmqC84W866fU=",
      "last_results_hash": "p/BPexV4LxAzlVcPRvW+lomgXb6Yze8YLIQUo/4Kdgc=",
      "evidence_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
      "proposer_address": "G0MeY8xQx7ooOsni8KE/3R/Ib3Q="
    },
    "valset": [
      {
        "operator_address": "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcqcnylw",
        "consensus_pubkey": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "/O7BtNW0pafwfvomgR4ZnfldwPXiFfJs9mHg3gwfv5Q="
        },
        "jailed": false,
        "status": "BOND_STATUS_BONDED",
        "tokens": "1416521659632",
        "delegator_shares": "1416521659632.000000000000000000",
        "description": {
          "moniker": "SG-1",
          "identity": "48608633F99D1B60",
          "website": "https://sg-1.online",
          "security_contact": "",
          "details": "SG-1 - your favorite validator on cosmos. We offer 100% Soft Slash protection."
        },
        "unbonding_height": "0",
        "unbonding_time": "1970-01-01T00:00:00Z",
        "commission": {
          "commission_rates": {
            "rate": "0.037500000000000000",
            "max_rate": "0.200000000000000000",
            "max_change_rate": "0.030000000000000000"
          },
          "update_time": "2021-10-01T15:00:00Z"
        },
        "min_self_delegation": "1"
      },
      {
        "operator_address": "cosmosvaloper1t8ehvswxjfn3ejzkjtntcyrqwvmvuknzmvtaaa",
        "consensus_pubkey": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "uExZyjNLtr2+FFIhNDAMcQ8+yTrqE7ygYTsI7khkA5Y="
        },
        "jailed": false,
        "status": "BOND_STATUS_BONDED",
        "tokens": "1348298958808",
        "delegator_shares": "1348298958808.000000000000000000",
        "description": {
          "moniker": "Cosmostation",
          "identity": "AE4C403A6E7AA1AC",
          "website": "https://www.cosmostation.io",
          "security_contact": "admin@stamper.network",
          "details": "Cosmostation validator node. Delegate your tokens and Start Earning Staking Rewards"
        },
        "unbonding_height": "0",
        "unbonding_time": "1970-01-01T00:00:00Z",
        "commission": {
          "commission_rates": {
            "rate": "0.050000000000000000",
            "max_rate": "1.000000000000000000",
            "max_change_rate": "0.200000000000000000"
          },
          "update_time": "2021-10-01T15:06:38.821314287Z"
        },
        "min_self_delegation": "1"
      }
    ]
  }
}
```

#### Parameters

The `Parameters` REST endpoint queries the staking parameters.

```bash
/cosmos/staking/v1beta1/params
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/params" -H  "accept: application/json"
```

Example Output:

```bash
{
  "params": {
    "unbonding_time": "2419200s",
    "max_validators": 100,
    "max_entries": 7,
    "historical_entries": 10000,
    "bond_denom": "stake"
  }
}
```

#### Pool

The `Pool` REST endpoint queries the pool information.

```bash
/cosmos/staking/v1beta1/pool
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/pool" -H  "accept: application/json"
```

Example Output:

```bash
{
  "pool": {
    "not_bonded_tokens": "432805737458",
    "bonded_tokens": "15783637712645"
  }
}
```

#### Validators

The `Validators` REST endpoint queries all validators that match the given status.

```bash
/cosmos/staking/v1beta1/validators
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/validators" -H  "accept: application/json"
```

Example Output:

```bash
{
  "validators": [
    {
      "operator_address": "cosmosvaloper1q3jsx9dpfhtyqqgetwpe5tmk8f0ms5qywje8tw",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "N7BPyek2aKuNZ0N/8YsrqSDhGZmgVaYUBuddY8pwKaE="
      },
      "jailed": false,
      "status": "BOND_STATUS_BONDED",
      "tokens": "383301887799",
      "delegator_shares": "383301887799.000000000000000000",
      "description": {
        "moniker": "SmartNodes",
        "identity": "D372724899D1EDC8",
        "website": "https://smartnodes.co",
        "security_contact": "",
        "details": "Earn Rewards with Crypto Staking & Node Deployment"
      },
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z",
      "commission": {
        "commission_rates": {
          "rate": "0.050000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.100000000000000000"
        },
        "update_time": "2021-10-01T15:51:31.596618510Z"
      },
      "min_self_delegation": "1"
    },
    {
      "operator_address": "cosmosvaloper1q5ku90atkhktze83j9xjaks2p7uruag5zp6wt7",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "GDNpuKDmCg9GnhnsiU4fCWktuGUemjNfvpCZiqoRIYA="
      },
      "jailed": false,
      "status": "BOND_STATUS_UNBONDING",
      "tokens": "1017819654",
      "delegator_shares": "1017819654.000000000000000000",
      "description": {
        "moniker": "Noderunners",
        "identity": "812E82D12FEA3493",
        "website": "http://noderunners.biz",
        "security_contact": "info@noderunners.biz",
        "details": "Noderunners is a professional validator in POS networks. We have a huge node running experience, reliable soft and hardware. Our commissions are always low, our support to delegators is always full. Stake with us and start receiving your cosmos rewards now!"
      },
      "unbonding_height": "147302",
      "unbonding_time": "2021-11-08T22:58:53.718662452Z",
      "commission": {
        "commission_rates": {
          "rate": "0.050000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.100000000000000000"
        },
        "update_time": "2021-10-04T18:02:21.446645619Z"
      },
      "min_self_delegation": "1"
    }
  ],
  "pagination": {
    "next_key": "FONDBFkE4tEEf7yxWWKOD49jC2NK",
    "total": "2"
  }
}
```

#### Validator

The `Validator` REST endpoint queries validator information for given validator address.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "validator": {
    "operator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
    "consensus_pubkey": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "sIiexdJdYWn27+7iUHQJDnkp63gq/rzUq1Y+fxoGjXc="
    },
    "jailed": false,
    "status": "BOND_STATUS_BONDED",
    "tokens": "33027900000",
    "delegator_shares": "33027900000.000000000000000000",
    "description": {
      "moniker": "Witval",
      "identity": "51468B615127273A",
      "website": "",
      "security_contact": "",
      "details": "Witval is the validator arm from Vitwit. Vitwit is into software consulting and services business since 2015. We are working closely with Cosmos ecosystem since 2018. We are also building tools for the ecosystem, Aneka is our explorer for the cosmos ecosystem."
    },
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z",
    "commission": {
      "commission_rates": {
        "rate": "0.050000000000000000",
        "max_rate": "0.200000000000000000",
        "max_change_rate": "0.020000000000000000"
      },
      "update_time": "2021-10-01T19:24:52.663191049Z"
    },
    "min_self_delegation": "1"
  }
}
```

#### ValidatorDelegations

The `ValidatorDelegations` REST endpoint queries delegate information for given validator.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q/delegations" -H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_responses": [
    {
      "delegation": {
        "delegator_address": "cosmos190g5j8aszqhvtg7cprmev8xcxs6csra7xnk3n3",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "31000000000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "31000000000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1ddle9tczl87gsvmeva3c48nenyng4n56qwq4ee",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "628470000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "628470000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos10fdvkczl76m040smd33lh9xn9j0cf26kk4s2nw",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "838120000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "838120000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1n8f5fknsv2yt7a8u6nrx30zqy7lu9jfm0t5lq8",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "500000000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "500000000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos16msryt3fqlxtvsy8u5ay7wv2p8mglfg9hrek2e",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "61310000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "61310000"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "5"
  }
}
```

#### Delegation

The `Delegation` REST endpoint queries delegate information for given validator delegator pair.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations/{delegatorAddr}
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q/delegations/cosmos1n8f5fknsv2yt7a8u6nrx30zqy7lu9jfm0t5lq8" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_response": {
    "delegation": {
      "delegator_address": "cosmos1n8f5fknsv2yt7a8u6nrx30zqy7lu9jfm0t5lq8",
      "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
      "shares": "500000000.000000000000000000"
    },
    "balance": {
      "denom": "stake",
      "amount": "500000000"
    }
  }
}
```

#### UnbondingDelegation

The `UnbondingDelegation` REST endpoint queries unbonding information for given validator delegator pair.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations/{delegatorAddr}/unbonding_delegation
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu/delegations/cosmos1ze2ye5u5k3qdlexvt2e0nn0508p04094ya0qpm/unbonding_delegation" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "unbond": {
    "delegator_address": "cosmos1ze2ye5u5k3qdlexvt2e0nn0508p04094ya0qpm",
    "validator_address": "cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu",
    "entries": [
      {
        "creation_height": "153687",
        "completion_time": "2021-11-09T09:41:18.352401903Z",
        "initial_balance": "525111",
        "balance": "525111"
      }
    ]
  }
}
```

#### ValidatorUnbondingDelegations

The `ValidatorUnbondingDelegations` REST endpoint queries unbonding delegations of a validator.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/unbonding_delegations
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu/unbonding_delegations" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1q9snn84jfrd9ge8t46kdcggpe58dua82vnj7uy",
      "validator_address": "cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu",
      "entries": [
        {
          "creation_height": "90998",
          "completion_time": "2021-11-05T00:14:37.005841058Z",
          "initial_balance": "24000000",
          "balance": "24000000"
        }
      ]
    },
    {
      "delegator_address": "cosmos1qf36e6wmq9h4twhdvs6pyq9qcaeu7ye0s3dqq2",
      "validator_address": "cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu",
      "entries": [
        {
          "creation_height": "47478",
          "completion_time": "2021-11-01T22:47:26.714116854Z",
          "initial_balance": "8000000",
          "balance": "8000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```
