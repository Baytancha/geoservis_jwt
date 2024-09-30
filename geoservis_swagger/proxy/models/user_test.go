package models

import (
	"database/sql"
	"errors"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func inmemory_DB() *sql.DB {
	db, _ := sql.Open("sqlite3", ":memory:")

	_, err := db.Exec("CREATE TABLE users (id SERIAL PRIMARY KEY, email VARCHAR(100), hashed_password VARCHAR(100))")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TestUserModel(t *testing.T) {

	model := &UserModel{DB: inmemory_DB()}

	err := model.Insert("test", "test")
	if err != nil {
		t.Error(err)
	}

	t.Run("valid user", func(t *testing.T) {
		err = model.Authenticate("test", "test")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		err = model.Authenticate("testify", "test")
		if !errors.Is(err, ErrNoUser) {
			t.Error(err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		err = model.Authenticate("test", "tester")
		if !errors.Is(err, ErrWrongPassword) {
			t.Error(err)
		}
	})

}
