package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// init mysql and redis
	// initDB()
	// defer db.Close()
	// initRedis()
	// defer rdb.Close()
	// m.Errorf("main test")
	os.Exit(m.Run())
}
