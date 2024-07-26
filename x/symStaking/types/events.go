package types

// staking module event types
const (
	EventTypeCompleteUnbonding = "complete_unbonding"
	EventTypeCreateValidator   = "create_validator"
	EventTypeEditValidator     = "edit_validator"
	EventTypeUnbond            = "unbond"

	AttributeKeyValidator      = "validator"
	AttributeKeyCommissionRate = "commission_rate"
	AttributeKeyCreationHeight = "creation_height"
	AttributeKeyCompletionTime = "completion_time"
)
