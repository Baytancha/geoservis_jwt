package models

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoUser        = errors.New("user doesn't exist")
	ErrWrongPassword = errors.New("wrong password")
)

type UserModelInterface interface {
	Insert(email, password string) error
	Authenticate(email, password string) error
}
type User struct {
	id       int    `db_field:"id" db_type:"SERIAL PRIMARY KEY"`
	email    string `db_field:"email" db_type:"VARCHAR(100)"`
	password string `db_field:"password" db_type:"VARCHAR(100)"`
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (email, hashed_password)
	 VALUES(?, ?)`

	_, err = m.DB.Exec(stmt, email, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) error {
	var hashedPassword []byte

	stmt := "SELECT hashed_password FROM users WHERE email = ?"

	err := m.DB.QueryRow(stmt, email).Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoUser
		} else {
			return err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrWrongPassword
		} else {
			return err
		}
	}

	return nil

}
