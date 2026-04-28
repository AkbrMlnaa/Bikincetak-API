package controllers

import (
	"bikincetak-api/database"
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)


func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}


func GoogleLogin(c *fiber.Ctx) error {

	url := getGoogleOAuthConfig().AuthCodeURL(
		"random-state-string",
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

type GoogleUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func GoogleCallback(c *fiber.Ctx) error {
    state := c.Query("state")
    if state != "random-state-string" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "State invalid"})
    }

    code := c.Query("code")
    googleConfig := getGoogleOAuthConfig()

    tokenRes, err := googleConfig.Exchange(context.Background(), code)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Gagal menukar token dengan Google"})
    }

    // 2. Ambil data profil dari Google
    resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tokenRes.AccessToken)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Gagal mengambil data profil Google"})
    }
    defer resp.Body.Close()

    var googleUser GoogleUser
    if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membaca data dari Google"})
    }

    var user models.User
    if err := database.DB.Where("email = ?", googleUser.Email).First(&user).Error; err != nil {

        // Buat di ERPNext dulu
        customerID, errERP := erpnext.CreateCustomer(googleUser.Name, googleUser.Email, "")
        if errERP != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat customer di ERPNext via Google: " + errERP.Error()})
        }

        user = models.User{
            Email:      googleUser.Email,
            Password:   "",
            CustomerId: customerID,
        }

        if createErr := database.DB.Create(&user).Error; createErr != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Gagal mendaftarkan user baru"})
        }
    }

    claims := jwt.MapClaims{
        "email":       user.Email,
        "customer_id": user.CustomerId, 
        "expired":     time.Now().Add(time.Hour * 24).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    secret := os.Getenv("JWT_SECRET")
    t, errToken := token.SignedString([]byte(secret))

    if errToken != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal menggenerate token"})
    }

    frontendURL := os.Getenv("FRONTEND_URL")
    if frontendURL == "" {
        frontendURL = "http://localhost:5173" 
    }

    redirectURL := fmt.Sprintf("%s/login/success?token=%s", frontendURL, t)
    return c.Redirect(redirectURL, fiber.StatusFound)
}
