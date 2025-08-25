package tx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nghiatrann0502/kyra-kit/postgres"
)

// ===== Context key =====
type ctxKeyTx struct{}

func TxFrom(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(ctxKeyTx{}).(pgx.Tx); ok {
		return tx
	}

	return nil
}

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxKeyTx{}, tx)
}

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func FromCtxOr(ctx context.Context, fallback DBTX) DBTX {
	if tx := TxFrom(ctx); tx != nil {
		return tx
	}
	return fallback
}

// ===== Implement =====
type UnitOfWork struct{ db postgres.DBEngine }

func NewUnitOfWord(db postgres.DBEngine) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func WithinTx(ctx context.Context, u *UnitOfWork, fn func(ctx context.Context) error) error {
	_, err := WithinTxR(ctx, u, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})

	return err
}

func WithinTxR[T any](ctx context.Context, u *UnitOfWork, fn func(ctx context.Context) (T, error)) (T, error) {
	if existing := TxFrom(ctx); existing != nil {
		return fn(ctx)
	}

	tx, err := u.db.GetDB().Begin(ctx)
	if err != nil {
		var zero T
		return zero, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	res, err := fn(withTx(ctx, tx))
	if err != nil {
		return res, err
	}

	if err := tx.Commit(ctx); err != nil {
		return res, err
	}

	return res, nil
}
