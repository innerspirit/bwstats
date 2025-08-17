package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

// findReplayFiles scans the user's replay directory and returns a list of all .rep files
func findReplayFiles(progressCallback func(float64)) ([]string, error) {
	// Get the replay directory path
	replayDir := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "StarCraft", "Maps", "Replays", "AutoSave")
	
	// Check if the replay directory exists
	if _, err := os.Stat(replayDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("autoreplays directory not found: %s", replayDir)
	}

	// Get subdirectories (date folders)
	subdirs, err := os.ReadDir(replayDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read autoreplays directory: %v", err)
	}

	if len(subdirs) == 0 {
		return nil, fmt.Errorf("no date folders found in autoreplays directory")
	}

	// Find all .rep files in the directory and subdirectories
	var repFiles []string

	// Walk through each subdirectory to find all .rep files
	for i, subdir := range subdirs {
		if !subdir.IsDir() {
			continue
		}

		subDirPath := filepath.Join(replayDir, subdir.Name())
		err := filepath.Walk(subDirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if it's a .rep file
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".rep") {
				repFiles = append(repFiles, path)
			}

			return nil
		})

		// Update progress after each subdirectory is processed
		if progressCallback != nil {
			progressCallback(float64(i+1) / float64(len(subdirs)))
		}

		if err != nil {
			return nil, fmt.Errorf("failed to walk subdirectory %s: %v", subdir.Name(), err)
		}
	}
	
	return repFiles, nil
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
