package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Account represents a user account from CSettings.json
type Account struct {
	Account string `json:"account"`
}

// Settings mirrors akafinder's structure for CSettings.json
// See akafinder `main.go` for reference.
type Settings struct {
	GatewayHistory []Account `json:"Gateway History"`
}

func main() {
	// Create a new Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("BW Stats - User Info")
	myWindow.Resize(fyne.NewSize(420, 180))

	// Load user settings
	userNickname, err := loadUserNickname()
	if err != nil {
		userNickname = "Error loading settings: " + err.Error()
	}

	// UI elements
	welcomeLabel := widget.NewLabel("Welcome to BW Stats")
	welcomeLabel.TextStyle = fyne.TextStyle{Bold: true}

	userLabel := widget.NewLabel("Current User: " + userNickname)
	userLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		welcomeLabel,
		userLabel,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// loadUserNickname loads the user's nickname from CSettings.json
func loadUserNickname() (string, error) {
	// Get user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	// Construct path to CSettings.json similarly to akafinder
	settingsPath := filepath.Join(userHome, "Documents", "StarCraft", "CSettings.json")

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return "Settings file not found", nil
	}

	// Open and read the file
	f, err := os.Open(settingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to open settings file: %v", err)
	}
	defer f.Close()

	// Read file content
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read settings file: %v", err)
	}

	// Parse JSON using the same structure as akafinder
	var settings Settings
	if err := json.Unmarshal(b, &settings); err != nil {
		return "", fmt.Errorf("failed to parse settings: %v", err)
	}

	// Use the first account as the current user
	if len(settings.GatewayHistory) > 0 {
		return settings.GatewayHistory[0].Account, nil
	}

	return "No accounts found", nil
}
