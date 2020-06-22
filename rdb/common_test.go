//+build integration

package rdb

import "testing"

var (
	testRepo *MySQLRepository
)

func TestMain(m *testing.M) {
	testRepo = NewMySQLRepository(MustLoadConfig())
	defer testRepo.Close()
	m.Run()
}
