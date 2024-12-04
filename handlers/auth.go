//JWT ISSUER
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var SecretKey = []byte("secret") // Use a secure key in production

type JwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

// Login handler to issue JWT token
func JWTHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

  fmt.Println(username, password)

	// Basic username/password check (you can improve this)
	if username != "Adam" || password != "655957" {
		return echo.ErrUnauthorized
	}

	// Set custom claims
	claims := &JwtCustomClaims{
		"Adam Fraga",
		true,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// Log the token
	// Generate encoded token and send it as response
	t, err := token.SignedString(SecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}
