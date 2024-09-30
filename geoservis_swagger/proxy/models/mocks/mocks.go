package models

type MockUserModel struct {
	Insert_field       func(email, password string) error
	Authenticate_field func(email, password string) error
}

func (m *MockUserModel) Insert(email, password string) error {
	return m.Insert_field(email, password)
}

func (m *MockUserModel) Authenticate(email, password string) error {
	return m.Authenticate_field(email, password)
}
