//TOTP 2FA
package handlers 

import (
   "fmt"
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/pquerna/otp/totp"
    "github.com/boombuler/barcode"
    "github.com/boombuler/barcode/qr"
    "image/png"
    "os"
)

// GenerateQRCode generates a QR code image for the TOTP URL.
func GenerateQRCode(url, filepath string, c echo.Context) error {
    // Generate QR code
    qrCode, err := qr.Encode(url, qr.L, qr.Auto )
    if err != nil {
        c.Logger().Error("Fail to generate QR Code", err)
        return fmt.Errorf("failed to generate QR code: %v", err)
    }

    // Scale the QR code to desired size
    qrCode, err = barcode.Scale(qrCode, 256, 256)
    if err != nil {
        c.Logger().Error("Fail to scale QR Code", err)
        return fmt.Errorf("failed to scale QR code: %v", err)
    }

    // Create file to save QR code image
    file, err := os.Create(filepath)
    if err != nil {
        c.Logger().Error("Fail to create file", err)
        return fmt.Errorf("failed to create file: %v", err)
    }
    defer file.Close()

    // Write the PNG image to the file
    err = png.Encode(file, qrCode)
    if err != nil {
        return fmt.Errorf("failed to encode PNG: %v", err)
    }

    return nil
}

// QRHandler handles the generation and serving of the QR code.
func QRHandler(c echo.Context) error {


    url := "otpauth://totp/MyApp?secret=JBSWY3DPEHPK3PXP&issuer=MyApp"
    filepath := "totp-qr.png"

    // Generate the QR code
    err := GenerateQRCode(url, filepath, c)
    if err != nil {
        c.Logger().Error("Error Generating QR Code: %s", err)
        return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
    }

    // Serve the generated QR code image as a response
    file, err := os.Open(filepath)
    if err != nil {
        c.Logger().Error("Error opening file to write QR Code:", err)
        return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error opening file: %v", err))
    }
    defer file.Close()

    return c.Stream(http.StatusOK, "image/png", file)
}

// GenerateTOTPKey creates a TOTP secret and returns the secret and URL for QR code.
func GenerateTOTPKey(userEmail string) (string, string, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "Hadash", // Replace with your app name
        AccountName: userEmail,     // The user’s email or username
    })
    if err != nil {
        return "", "", err
    }

    return key.Secret(), key.URL(), nil
}



// setup2FA handler for generating TOTP secret and QR code URL
func Setup2FA(c echo.Context) error {
    userEmail := c.QueryParam("email") // Get user email from query parameters

    if userEmail == "" {
        c.Logger().Error("Email is required")
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
    }

    // Generate TOTP secret and URL
    secret, url, err := GenerateTOTPKey(userEmail)
    if err != nil {
        c.Logger().Error("Failed to generate TOTP")
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate TOTP"})
    }

    // Generate the QR code image for the user to scan
    qrFile := fmt.Sprintf("%s_qr.png", userEmail) // File name based on user email
    if err := GenerateQRCode(url, qrFile, c); err != nil {
        c.Logger().Error("Failed to generate QR Code")
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate QR code"})
    }

    // Return the secret and QR code file path in the response
    return c.JSON(http.StatusOK, map[string]string{
        "secret":    secret,
        "qrCodePath": qrFile, // Path to the generated QR code
        "url":       url,     // TOTP URL for debugging or manual entry
    })
}

// verify2FA handler for verifying TOTP code
func Verify2FA(c echo.Context) error {
    type request struct {
        Code   string `json:"code"`
        Secret string `json:"secret"`
    }

    // Bind request data
    req := new(request)
    if err := c.Bind(req); err != nil {
        c.Logger().Error("Invalid request data: %s", err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
    }

    // Verify the TOTP code using the provided secret
    if totp.Validate(req.Code, req.Secret) {
        c.Logger().Error("Code verified successfully !")
        return c.JSON(http.StatusOK, map[string]string{"message": "Code verified successfully"})
    }
    c.Logger().Error("Code is invalid !")
    return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid code"})
}
