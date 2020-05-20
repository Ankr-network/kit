//+build integration

package rdb

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMySQLRepository(t *testing.T) {
	t.Logf("status:%+v", testRepo.DB.Stats())
}

func TestMySQLRepository_NewReadTx(t *testing.T) {
	tx, err := testRepo.NewReadTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	err = testRepo.UpdateOne(ctx, `INSERT INTO test.food (name) VALUES ()`, "apple")
	assert.True(t, errors.Is(err, ErrExecInReadOnlyTx))
}

func TestMySQLRepository_AddOne(t *testing.T) {
	tx, err := testRepo.NewWriteTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	id, err := testRepo.AddOne(ctx, `INSERT INTO test.food (name) VALUES (?)`, "apple")
	require.NoError(t, err)
	assert.True(t, id > 0)

	id, err = testRepo.AddOne(ctx, `INSERT INTO test.food (id, name) VALUES (?, ?)`, id, "orange")
	assert.True(t, errors.Is(err, ErrDuplicateKey))
	assert.Equal(t, int64(0), id)
}

func TestMySQLRepository_SaveOne(t *testing.T) {
	tx, err := testRepo.NewWriteTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	err = testRepo.SaveOne(ctx, `INSERT INTO test.food (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?`, 1, "apple", "apple")
	require.NoError(t, err)

	type food struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}

	out := new(food)

	err = tx.Get(out, `SELECT id, name FROM test.food WHERE id = ?`, 1)
	require.NoError(t, err)
	assert.Equal(t, "apple", out.Name)

	err = testRepo.SaveOne(ctx, `INSERT INTO test.food (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?`, 1, "orange", "orange")
	require.NoError(t, err)

	err = tx.Get(out, `SELECT id, name FROM test.food WHERE id = ?`, 1)
	require.NoError(t, err)
	assert.Equal(t, "orange", out.Name)
}

func TestMySQLRepository_UpdateOne(t *testing.T) {
	tx, err := testRepo.NewWriteTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	err = testRepo.UpdateOne(ctx, `UPDATE test.food SET name = ? WHERE id = ?`, "apple", -1)
	assert.True(t, errors.Is(err, ErrNotFound))

	res, err := tx.Exec(`INSERT INTO test.food (name) VALUES (?),(?)`, "apple", "orange")
	require.NoError(t, err)
	id, err := res.LastInsertId()
	assert.NoError(t, err)
	assert.True(t, id > 0)
	aft, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), aft)

	var oid int64
	err = tx.Get(&oid, `SELECT id FROM test.food WHERE name = ?`, "orange")
	require.NoError(t, err)

	err = testRepo.UpdateOne(ctx, `UPDATE test.food SET id = ? WHERE name = ?`, oid, "apple")
	assert.True(t, errors.Is(err, ErrDuplicateKey))
}

func TestMySQLRepository_DeleteOne(t *testing.T) {
	tx, err := testRepo.NewWriteTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	err = testRepo.DeleteOne(ctx, `DELETE FROM test.food WHERE id = ?`, -1)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestMySQLRepository_FindOne(t *testing.T) {
	tx, err := testRepo.NewWriteTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	type food struct {
		ID   int64
		Name string
	}

	o := new(food)
	err = testRepo.FindOne(ctx, o, `SELECT id, name FROM test.food WHERE id = ?`, -1)
	assert.True(t, errors.Is(err, ErrNotFound))

}

func TestMySQLRepository_FindAll(t *testing.T) {
	tx, err := testRepo.NewWriteTx(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()

	ctx := ContextWithTx(context.Background(), tx)

	type food struct {
		ID   int64
		Name string
	}

	var o []food
	err = testRepo.FindAll(ctx, &o, `SELECT id, name FROM test.food`)
	assert.NoError(t, err)
	assert.Len(t, o, 0)

	_, err = tx.Exec(`INSERT INTO test.food (name) VALUES (?),(?)`, "apple", "orange")
	require.NoError(t, err)

	err = tx.Select(&o, `SELECT id, name FROM test.food`)
	assert.NoError(t, err)
	assert.Len(t, o, 2)
}
