package main

func GenerateToken(name string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"username": name})
	return tokenString
}
