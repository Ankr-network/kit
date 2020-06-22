package rdb

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type txKey struct{}

func ContextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetTxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	out, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	return out, ok
}

type Repository interface {
	FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	FindAll(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	AddOne(ctx context.Context, query string, args ...interface{}) (id int64, err error)
	SaveOne(ctx context.Context, query string, args ...interface{}) error
	UpdateOne(ctx context.Context, query string, args ...interface{}) error
	DeleteOne(ctx context.Context, query string, args ...interface{}) error
	NewReadTx(ctx context.Context) (*sqlx.Tx, error)
	NewWriteTx(ctx context.Context) (*sqlx.Tx, error)
	WithWriteTx(ctx context.Context, h func(ctx context.Context) error) error
	GetSQLOp(ctx context.Context) SQLOp
}

type SQLOp interface {
	sqlx.Ext
	sqlx.ExecerContext
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type MySQLRepository struct {
	*sqlx.DB
}

func NewMySQLRepository(cfg *Config) *MySQLRepository {
	db := sqlx.MustConnect("mysql", cfg.DSN)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.SetMaxOpenConns)
	return &MySQLRepository{
		DB: db,
	}
}

func (m *MySQLRepository) NewReadTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := m.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		log.Error("BeginTxx error", zap.Error(err))
		return nil, err
	}
	return tx, nil
}

func (m *MySQLRepository) NewWriteTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := m.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  false,
	})
	if err != nil {
		log.Error("BeginTxx error", zap.Error(err))
		return nil, err
	}
	return tx, nil
}

func (m *MySQLRepository) FindOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	LogSQL(query, args...)
	err := m.GetSQLOp(ctx).GetContext(ctx, dest, query, args...)
	if err != nil {
		if IsMySQLNotFoundError(err) {
			return ErrNotFound
		}
		log.Error("GetContext error", zap.Error(err))
		return err
	}
	return nil
}

func (m *MySQLRepository) FindAll(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	LogSQL(query, args...)
	err := m.GetSQLOp(ctx).SelectContext(ctx, dest, query, args...)
	if err != nil {
		log.Error("SelectContext error", zap.Error(err))
		return err
	}
	return nil
}

func (m *MySQLRepository) AddOne(ctx context.Context, query string, args ...interface{}) (id int64, err error) {
	LogSQL(query, args...)
	rs, err := m.GetSQLOp(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		if IsMySQLDuplicateError(err) {
			return 0, fmt.Errorf("%w:%v", ErrDuplicateKey, err)
		}
		if IsMySQLExecInReadOnlyTxError(err) {
			return 0, fmt.Errorf("%w:%v", ErrExecInReadOnlyTx, err)
		}
		log.Error("ExecContext error", zap.Error(err))
		return 0, err
	}
	id, err = rs.LastInsertId()
	if err != nil {
		log.Error("LastInsertId error", zap.Error(err))
		return 0, err
	}

	aft, err := rs.RowsAffected()
	if err != nil {
		log.Error("RowsAffected error", zap.Error(err))
		return 0, err
	}
	if aft > 1 {
		return 0, ErrAffectMany
	}

	return id, nil
}

func (m *MySQLRepository) SaveOne(ctx context.Context, query string, args ...interface{}) error {
	LogSQL(query, args...)
	rs, err := m.GetSQLOp(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		if IsMySQLDuplicateError(err) {
			return fmt.Errorf("%w:%v", ErrDuplicateKey, err)
		}
		if IsMySQLExecInReadOnlyTxError(err) {
			return fmt.Errorf("%w:%v", ErrExecInReadOnlyTx, err)
		}
		log.Error("ExecContext error", zap.Error(err))
		return err
	}
	aft, err := rs.RowsAffected()
	if err != nil {
		log.Error("RowsAffected error", zap.Error(err))
		return err
	}
	if aft > 2 { // aft: 0 nothing change, 1 insert, 2 update
		return ErrAffectMany
	}

	return nil
}

func (m *MySQLRepository) UpdateOne(ctx context.Context, query string, args ...interface{}) error {
	LogSQL(query, args...)
	op := m.GetSQLOp(ctx)
	rs, err := op.ExecContext(ctx, query, args...)
	if err != nil {
		if IsMySQLDuplicateError(err) {
			return fmt.Errorf("%w:%v", ErrDuplicateKey, err)
		}
		if IsMySQLExecInReadOnlyTxError(err) {
			return fmt.Errorf("%w:%v", ErrExecInReadOnlyTx, err)
		}
		log.Error("ExecContext error", zap.Error(err))
		return err
	}
	aft, err := rs.RowsAffected()
	if err != nil {
		log.Error("RowsAffected error", zap.Error(err))
		return err
	}
	if aft < 1 {
		return ErrNothingUpdated
	} else if aft > 1 {
		return ErrAffectMany
	}
	return nil
}

func (m *MySQLRepository) DeleteOne(ctx context.Context, query string, args ...interface{}) error {
	LogSQL(query, args...)
	op := m.GetSQLOp(ctx)
	rs, err := op.ExecContext(ctx, query, args...)
	if err != nil {
		if IsMySQLExecInReadOnlyTxError(err) {
			return fmt.Errorf("%w:%v", ErrExecInReadOnlyTx, err)
		}
		log.Error("ExecContext error", zap.Error(err))
		return err
	}
	aft, err := rs.RowsAffected()
	if err != nil {
		log.Error("RowsAffected error", zap.Error(err))
		return err
	}
	if aft < 1 {
		return ErrNotFound
	} else if aft > 1 {
		return ErrAffectMany
	}
	return nil
}

func (m *MySQLRepository) GetSQLOp(ctx context.Context) SQLOp {
	tx, ok := GetTxFromContext(ctx)
	if ok {
		return tx
	}
	return m
}

func (m *MySQLRepository) WithWriteTx(ctx context.Context, h func(ctx context.Context) error) error {
	tx, err := m.NewWriteTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = h(ContextWithTx(ctx, tx))
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Error("Commit error", zap.Error(err))
		return err
	}
	return nil
}
