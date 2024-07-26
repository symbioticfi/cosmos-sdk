package keeper

import (
	"context"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator instance.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx context.Context) error {
	return nil
}

// Migrate2to3 migrates x/symStaking state from consensus version 2 to 3.
func (m Migrator) Migrate2to3(ctx context.Context) error {
	return nil
}

// Migrate3to4 migrates x/symStaking state from consensus version 3 to 4.
func (m Migrator) Migrate3to4(ctx context.Context) error {
	return nil
}

// Migrate4to5 migrates x/symStaking state from consensus version 4 to 5.
func (m Migrator) Migrate4to5(ctx context.Context) error {
	return nil
}

// Migrate4to5 migrates x/symStaking state from consensus version 5 to 6.
func (m Migrator) Migrate5to6(ctx context.Context) error {
	return nil
}
