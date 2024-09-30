package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	models "test/models"
	mocks "test/models/mocks"
	"testing"
	"time"
)

type MockGeoService struct {
	AddressSearch_field func(input string) ([]*Address, error)
	GeoCode_field       func(lat, lng string) ([]*Address, error)
}

func (m *MockGeoService) AddressSearch(input string) ([]*Address, error) {
	return m.AddressSearch_field(input)
}

func (m *MockGeoService) GeoCode(lat, lng string) ([]*Address, error) {
	return m.GeoCode_field(lat, lng)
}

func newApp(mock *mocks.MockUserModel) *application {

	app := &application{
		geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		user:   mock,
	}

	return app

}

func TestAddressSearch(t *testing.T) {

	geo := NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9")
	addresses, err := geo.AddressSearch("Москва, ул Сухонская")
	if err != nil {
		t.Error(err)
	}
	if len(addresses) == 0 {
		t.Error("no addresses")
	}

	empty, err := geo.AddressSearch("Босква, ул Бухонская")
	if err != nil {
		t.Error(err)
	}
	if len(empty) != 0 {
		t.Error("should be empty addresses")
	}

}

func TestGeoCode(t *testing.T) {
	geo := NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9")
	geoCode, err := geo.GeoCode("55.878", "37.653")
	if err != nil {
		t.Error(err)
	}
	if len(geoCode) == 0 {
		t.Error("no addresses")
	}

	empty, err := geo.GeoCode("-7575", "-867868")
	if err != nil {
		t.Error(err)
	}
	if len(empty) != 0 {
		t.Error("should be empty")
	}

	empty2, err := geo.GeoCode("sdfsfsfsf", "fsfsf")
	if err != nil {
		t.Error(err)
	}
	if len(empty2) != 0 {
		t.Error("should be empty")
	}
}

func TestMarshalUnMarshalGeoCode(t *testing.T) {
	client := NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9")
	lat, lng := "55.878", "37.653"
	httpClient := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{"lat": %s, "lon": %s}`, lat, lng))
	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address", data)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", client.apiKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("buffer", string(buf))

	geocode, err := UnmarshalGeoCode(buf)
	if err != nil {
		t.Error(err)
	}

	if geocode.Suggestions[0].Data.City != "Москва" {
		t.Error("wrong city")
	}

	_, err = geocode.Marshal()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("end")
}

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"SearchHandler1_valid", "/api/address/search", http.StatusOK, []byte(`{"query":"Москва, ул Сухонская"}`)},
		{"SearchHandler2_invalid", "/api/address/search", http.StatusBadRequest, []byte(`"name":"John","age":30}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateToken(tt.name)

			cookie := &http.Cookie{
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Second),
				SameSite: http.SameSiteLaxMode,
				Name:     "jwt",
				Value:    token,
			}
			req := httptest.NewRequest("POST", tt.path, bytes.NewReader(tt.body))
			req.AddCookie(cookie)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestGeoCodeHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"GeocodeHandler1_valid", "/api/address/geocode", http.StatusOK, []byte(`{"lat": "55.878", "lng": "37.653"}`)},
		{"GeocodeHandler2_invalid", "/api/address/geocode", http.StatusBadRequest, []byte(` "lat": "55.878", "lng": "37.653"`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateToken(tt.name)

			cookie := &http.Cookie{
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Second),
				SameSite: http.SameSiteLaxMode,
				Name:     "jwt",
				Value:    token,
			}

			req := httptest.NewRequest("POST", tt.path, bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestSearchHandler_500(t *testing.T) {
	t.Run("internal server error", func(t *testing.T) {
		token := GenerateToken("test")

		cookie := &http.Cookie{
			HttpOnly: true,
			Expires:  time.Now().Add(7 * 24 * time.Second),
			SameSite: http.SameSiteLaxMode,
			Name:     "jwt",
			Value:    token,
		}

		req := httptest.NewRequest("POST", "/api/address/search", bytes.NewReader([]byte(`{"query":"Москва, ул Сухонская"}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		app := &application{
			geo: &MockGeoService{
				AddressSearch_field: func(input string) ([]*Address, error) { return nil, errors.New("some error") },
				GeoCode_field:       func(lat, lng string) ([]*Address, error) { return nil, errors.New("some error") },
			},
			logger: logger,
		}
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d but got %d", http.StatusInternalServerError, w.Code)
		}

	})
}

func TestGeoCodeHandler_500(t *testing.T) {
	t.Run("internal server error", func(t *testing.T) {
		token := GenerateToken("test")

		cookie := &http.Cookie{
			HttpOnly: true,
			Expires:  time.Now().Add(7 * 24 * time.Second),
			SameSite: http.SameSiteLaxMode,
			Name:     "jwt",
			Value:    token,
		}

		req := httptest.NewRequest("POST", "/api/address/geocode", bytes.NewReader([]byte(`{"lat": "55.878", "lng": "37.653"}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		app := &application{
			geo: &MockGeoService{
				AddressSearch_field: func(input string) ([]*Address, error) { return nil, errors.New("some error") },
				GeoCode_field:       func(lat, lng string) ([]*Address, error) { return nil, errors.New("some error") },
			},
			logger: logger,
		}
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d but got %d", http.StatusInternalServerError, w.Code)
		}

	})
}

func TestFileHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"Fileserver_valid", "/swagger/", http.StatusOK, nil},
		{"Fileserver_invalid", "/swagger/smth", http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		token := GenerateToken(tt.name)

		cookie := &http.Cookie{
			HttpOnly: true,
			Expires:  time.Now().Add(7 * 24 * time.Second),
			SameSite: http.SameSiteLaxMode,
			Name:     "jwt",
			Value:    token,
		}

		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("GET", tt.path, nil)
			req.AddCookie(cookie)
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	t.Run("failed to parse form", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Authenticate_field: func(email, password string) error {
				return models.ErrNoUser
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader("foo%3z1%26bar%3D2"))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}

	})

	t.Run("missing loging or password", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Authenticate_field: func(email, password string) error {
				return models.ErrNoUser
			},
		}
		data := url.Values{}
		data.Set("email", "")
		data.Set("password", "")
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}

	})

	t.Run("failed to authenticate", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Authenticate_field: func(email, password string) error {
				return models.ErrNoUser
			},
		}
		data := url.Values{}
		data.Set("email", "foo")
		data.Set("password", "bar")
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %d", http.StatusOK, w.Code)
		}

	})

	t.Run("failed to authenticate", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Authenticate_field: func(email, password string) error {
				return models.ErrWrongPassword
			},
		}
		data := url.Values{}
		data.Set("email", "foo")
		data.Set("password", "bar")
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %d", http.StatusOK, w.Code)
		}

	})

	t.Run("other errors", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Authenticate_field: func(email, password string) error {
				return errors.New("some error")
			},
		}
		data := url.Values{}
		data.Set("email", "foo")
		data.Set("password", "bar")
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d but got %d", http.StatusInternalServerError, w.Code)
		}

	})

	t.Run("happy path", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Authenticate_field: func(email, password string) error {
				return nil
			},
		}
		data := url.Values{}
		data.Set("email", "foo")
		data.Set("password", "bar")
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %d", http.StatusOK, w.Code)
		}

		if w.Body.String() != "successfully logged in" {
			t.Errorf("expected 'true' but got %s", w.Body.String())
		}

	})

}

func TestRegisterHandler(t *testing.T) {
	t.Run("failed to parse form", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Insert_field: func(email, password string) error {
				return models.ErrNoUser
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader("foo%3z1%26bar%3D2"))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}

	})

	t.Run("missing loging or password", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Insert_field: func(email, password string) error {
				return models.ErrNoUser
			},
		}
		data := url.Values{}
		data.Set("email", "")
		data.Set("password", "")
		req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}

	})

	t.Run("failed to insert a user", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Insert_field: func(email, password string) error {
				return errors.New("some error")
			},
		}
		data := url.Values{}
		data.Set("email", "foo")
		data.Set("password", "bar")
		req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %d", http.StatusOK, w.Code)
		}

		if w.Body.String() != "failed to insert user" {
			t.Errorf("expected 'true' but got %s", w.Body.String())
		}
	})

	t.Run("happy path", func(t *testing.T) {
		service := &mocks.MockUserModel{
			Insert_field: func(email, password string) error {
				return nil
			},
		}
		data := url.Values{}
		data.Set("email", "foo")
		data.Set("password", "bar")
		req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(data.Encode()))
		w := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app := newApp(service)
		r := app.setupRouter()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %d", http.StatusOK, w.Code)
		}

		if w.Body.String() != "successfully signed up" {
			t.Errorf("expected 'true' but got %s", w.Body.String())
		}

	})

}
