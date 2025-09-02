package pg

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type BunTxActionFn func(tx bun.Tx) error

func HandleBunTransaction(ctx context.Context, bunTx BunTxActionFn, db *bun.DB) (err error) {
	var tx bun.Tx

	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = bunTx(tx)

	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("failed executing a transaction: %w; failed to rollback a transaction: %w", err, errRollback)
		}
		return fmt.Errorf("failed executing transaction: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
