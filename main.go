package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

// initialize project with user account
func InitializeFirebaseApp() (*firebase.App, error) {
	opt := option.WithCredentialsFile("testing-e3e8d-firebase-adminsdk-eoic0-21df8f66f2.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}
	return app, nil
}

// login system
func AuthenticateWithEmail(email, password string) (string, error) {
	client := &http.Client{}
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=AIzaSyAec4x7qAj27adl3KEtTdGy1nxpD7_dRAM"

	payload := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error authenticating with email and password: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed, status code: %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("error decoding response body: %v", err)
	}

	token, ok := respBody["idToken"].(string)
	if !ok {
		return "", fmt.Errorf("idToken not found in response")
	}

	return token, nil
}

// try to send data to realtime database
func SendDataToRealtimeDatabase(token, path string, data interface{}) error {
	client := &http.Client{}

	uniqueID := uuid.New().String()

	// Append the unique ID to the base path
	pathWithUniqueID := fmt.Sprintf("%s/%s", path, uniqueID)

	url := fmt.Sprintf("https://testing-e3e8d-default-rtdb.asia-southeast1.firebasedatabase.app/%s.json?auth=%s", pathWithUniqueID, token)

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending data to realtime database: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error sending data: ", err)
		return fmt.Errorf("failed to send data, status code: %d", resp.StatusCode)
	}

	return nil
}

// Handler untuk melayani beranda.html
func ServeBeranda(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/beranda.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handler untuk melayani signin.html
func ServeSignin(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/signin.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handler untuk login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ambil data dari form
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Autentikasi dengan Firebase
	_, err := AuthenticateWithEmail(email, password) // token tidak digunakan di sini
	if err != nil {
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Jika login berhasil, arahkan ke halaman dashboard.html
	http.Redirect(w, r, "/dashboard.html", http.StatusSeeOther)
}

// Handler untuk melayani dashboard.html
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func main() {
	_, err := InitializeFirebaseApp()
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v", err)
	}

	// Endpoint untuk beranda.html
	http.HandleFunc("/", ServeBeranda)

	// Endpoint untuk signin.html
	http.HandleFunc("/signin.html", ServeSignin)

	// Endpoint untuk login
	http.HandleFunc("/login", LoginHandler)

	// Endpoint untuk dashboard
	http.HandleFunc("/dashboard.html", DashboardHandler)

	// Jalankan server
	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
