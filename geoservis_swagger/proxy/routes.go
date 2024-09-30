package main

import (
	"net/http"

	"test/swagger"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"

	_ "github.com/mattn/go-sqlite3"
)

func (app *application) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	proxy := ReverseProxy{
		host: "hugo_task",
		port: "1313",
	}
	r.Use(proxy.ReverseProxy)

	r.Group(func(r chi.Router) {

		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token, _, err := jwtauth.FromContext(r.Context())

				if err != nil {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				if token == nil || jwt.Validate(token) != nil {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
			})
		})
		//r.Use(Authenticator(tokenAuth))

		r.Post("/api/address/search", app.SearchHandler)
		r.Post("/api/address/geocode", app.GeocodeHandler)

	})

	r.Post("/api/login", app.Login)
	r.Post("/api/register", app.Register)

	fileServer := http.FileServerFS(swagger.Swaggerfile)
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/swagger", fileServer)
		fs.ServeHTTP(w, r)
	})

	// r.Group(func(r chi.Router) {
	// 	r.Use(jwtauth.Verifier(tokenAuth))
	// 	r.Use(jwtauth.Authenticator)

	// 	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
	// 		_, claims, _ := jwtauth.FromContext(r.Context())
	// 		userID := claims["sub"].(string)
	// 		fmt.Fprintf(w, "Защищенный маршрут! Пользователь с ID %s авторизован.", userID)
	// 	})
	// })

	return r
}
