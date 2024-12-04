package main

import (
	"github.com/golang-jwt/jwt/v5"
  echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	h "github.com/adam-fraga/hadash/handlers"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Enable CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	AllowOrigins: []string{"http://localhost:5173"}, // React frontend
		AllowMethods: []string{echo.GET, echo.POST, echo.DELETE, echo.PUT },
	}))

	// Login route (issue JWT token)
	e.POST("/login", h.JWTHandler)

	// Protected routes (need JWT token)
	r := e.Group("/protected")
	r.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(h.JwtCustomClaims)
		},
		SigningKey: h.SecretKey,
	}))

  // Protected route by JWT
	r.GET("/resource", h.RestrictedRoute)

  // Route to generate and serve the QR code to the user (Need frontend to serve the qr code image to the user)
  e.GET("/generate-qr", h.QRHandler)

  // Setup 2FA with user email (Here the client need to develop a proper frontend to serve the qrcode image)
  //curl "http://localhost:8080/setup2fa?email=user@example.com"
  e.GET("/setup-2fa", h.Setup2FA)

  // Route to verify 2FA after scanning QR code
  //curl -X POST http://localhost:8080/verify2fa -d '{"code":"123456", "secret":"<secret_from_setup2fa>"}' -H "Content-Type: application/json"
  e.GET("/verify-2fa", h.Verify2FA)

	// Start the server
	e.Logger.Fatal(e.Start(":8081"))
}
