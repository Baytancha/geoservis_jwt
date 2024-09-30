package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"test/models"
	"time"
)

// swagger:parameters GetAddress
type SearchRequest struct {
	//A search request in JSON format
	//example: Москва Обуховская 11
	Query string `json:"query"`
}

//
//swagger:model
type SearchResponse struct {
	// An array of addresses
	Addresses []*Address `json:"addresses"`
}

// swagger:parameters GetAddressByGeocode
type GeocodeRequest struct {
	//latitude
	Lat string `json:"lat"`
	//longitude
	Lng string `json:"lng"`
}

//swagger:model
type GeocodeResponse struct {
	//An array of addresses
	Addresses []*Address `json:"addresses"`
}

func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	//swagger:route POST /api/register SignUp
	// swagger:operation POST /api/register SignUp
	//
	// signup handler
	//
	//
	//
	// ---
	// consumes:
	// - x-www-form-urlencoded
	// parameters:
	// - name: email
	//   in: body
	//   type: string
	// - name: password
	//   in: body
	//   type: string
	// responses:
	//   '200':
	//     description: success or error
	//     schema:
	//         type: string
	//
	//   '400':
	//      description: invalid request body
	//      schema:
	//	        type: string
	//   '500':
	//        description: internal server error
	//        schema:
	//	        type: string

	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, fmt.Sprint("failed to parse form"))
	}
	userName := r.PostForm.Get("email")
	userPassword := r.PostForm.Get("password")
	fmt.Println("data", userName, userPassword)
	if userName == "" || userPassword == "" {
		http.Error(w, "empty username and password", http.StatusBadRequest)
		return
	}
	err = app.user.Insert(userName, userPassword)
	if err != nil {
		app.logger.Error("failed to insert user", "error", err.Error())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, fmt.Sprint("failed to insert user"))
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, fmt.Sprint("successfully signed up"))

}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	//swagger:route POST /api/login Login
	// swagger:operation POST /api/login Login
	//
	// login handler
	//
	//
	//
	// ---
	// consumes:
	// - x-www-form-urlencoded
	// parameters:
	// - name: email
	//   in: body
	//   type: string
	// - name: password
	//   in: body
	//   type: string
	// responses:
	//   '200':
	//     description: success of error
	//     schema:
	//         type: string
	//
	//   '400':
	//      description: invalid request body
	//      schema:
	//	        type: string
	//   '500':
	//        description: internal server error
	//        schema:
	//	        type: string

	err := r.ParseForm()
	if err != nil {
		app.logger.Error("failed to parse form", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userName := r.PostForm.Get("email")
	userPassword := r.PostForm.Get("password")
	fmt.Println("data", userName, userPassword)

	if userName == "" || userPassword == "" {
		http.Error(w, "Missing email or password.", http.StatusBadRequest)
		return
	}
	err = app.user.Authenticate(userName, userPassword)
	if err != nil {
		if errors.Is(err, models.ErrNoUser) {
			app.logger.Error("failed to authenticate", "error", err.Error())
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, fmt.Sprint("user doesn't exist"))
			return
		} else if errors.Is(err, models.ErrWrongPassword) {
			app.logger.Error("failed to authenticate", "error", err.Error())
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, fmt.Sprint("wrong password"))
			return
		}
		app.logger.Error("failed to authenticate", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, fmt.Sprint("failed to authenticate"))
		return
	}
	token := GenerateToken(userName)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		Name:     "jwt",
		Value:    token,
	})
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, fmt.Sprint("successfully logged in"))

}

func (app *application) SearchHandler(w http.ResponseWriter, r *http.Request) {
	//swagger:route POST /api/address/search GetAddress
	// swagger:operation POST /api/address/search GetAddress
	//
	// gets addresses either from URL query param or request body
	//
	//
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: addr_query
	//   in: query
	//   type: string
	// - name: addr_query
	//   in: body
	//   type: string
	// responses:
	//   '200':
	//     description: an array of addresses
	//     schema:
	//         items:
	//         "$ref": "#/definitions/SearchResponse"
	//   '400':
	//      description: invalid request body
	//      schema:
	//	        type: string
	//   '500':
	//        description: internal server error
	//        schema:
	//	        type: string

	var req SearchRequest
	req.Query = r.URL.Query().Get("query")
	if req.Query == "" {

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	fmt.Println(req.Query)
	addresses, err := app.geo.AddressSearch(req.Query)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := SearchResponse{Addresses: addresses}
	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	w.Write(responseJSON)

}

func (app *application) GeocodeHandler(w http.ResponseWriter, r *http.Request) {
	//swagger:route POST /api/address/geocode GetAddressByGeocode
	// swagger:operation POST /api/address/geocode GetAddressByGeocode
	//
	// gets addresses based on geographic coordinates submitted in URL query param or request body
	//
	//
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: lat
	//   in: query
	//   type: string
	// - name: lng
	//   in: query
	//   type: string
	// - name: lat_lng
	//   in: body
	//   type: string
	// responses:
	//  '200':
	//     description: an array of addresses
	//     schema:
	//         items:
	//         "$ref": "#/definitions/GeocodeResponse"
	//  '400':
	//      description: invalid request body
	//      schema:
	//	        type: string
	//  '500':
	//        description: internal server error
	//        schema:
	//	        type: string

	//

	var req GeocodeRequest
	req.Lat = r.URL.Query().Get("lat")
	req.Lng = r.URL.Query().Get("lng")
	if req.Lat == "" || req.Lng == "" {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	addresses, err := app.geo.GeoCode(req.Lat, req.Lng)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := GeocodeResponse{Addresses: addresses}

	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	w.Write(responseJSON)
}
