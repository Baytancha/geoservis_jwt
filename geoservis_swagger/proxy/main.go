//   Product Api:
//    version: 0.1
//    title: Product Api
//   Schemes: http, https
//   Host:
//   BasePath: /api/v1
//      Consumes:
//      - application/json
//   Produces:
//   - application/json
//   SecurityDefinitions:
//    Bearer:
//     type: apiKey
//     name: Authorization
//     in: header
//   swagger:meta

package main

import (
	"fmt"
	"log"
	"log/slog"

	"os"

	"database/sql"

	"github.com/go-chi/jwtauth/v5"

	_ "github.com/mattn/go-sqlite3"

	"test/models"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

func inmemory_DB() *sql.DB {
	db, _ := sql.Open("sqlite3", ":memory:")

	_, err := db.Exec("CREATE TABLE users (id SERIAL PRIMARY KEY, email VARCHAR(100), hashed_password VARCHAR(100))")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type application struct {
	geo    GeoProvider
	logger *slog.Logger
	user   models.UserModelInterface
}

func main() {
	fmt.Println("starting server")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
		logger: logger,
		user:   &models.UserModel{DB: inmemory_DB()},
	}
	err := app.serve()
	if err != nil {
		log.Fatal(err)
	}

}
