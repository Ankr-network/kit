package rdb

import "testing"

var (
	testRepo *MySQLRepository
)

func TestMain(m *testing.M) {
	testRepo = NewMySQLRepositoryWithConfig()
	defer testRepo.Close()
	m.Run()
}
