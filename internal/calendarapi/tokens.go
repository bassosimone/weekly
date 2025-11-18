// tokens.go - Code to manage tokens
// SPDX-License-Identifier: GPL-3.0-or-later

package calendarapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

// getToken retrieves a token from file or initiates the OAuth2 flow if needed.
func getToken(ctx context.Context, config *oauth2.Config, tokenPath string) (*oauth2.Token, error) {
	// Try to read token from file
	token, err := readTokenFromFile(tokenPath)
	if err == nil {
		return token, nil
	}

	// Token not found, get it from web
	token, err = getTokenFromWeb(ctx, config)
	if err != nil {
		return nil, err
	}

	// Save token for future use
	if err := saveTokenToFile(tokenPath, token); err != nil {
		return nil, err
	}

	return token, nil
}

// getTokenFromWeb initiates OAuth2 authorization code flow.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// TODO(bassosimone): we need to spin up a server on localhost to handle
	// the OAuth2 redirect. For now, the ghetto way to set things up is to copy
	// the `$code` value from the `code=$code` URL param and use it.
	//
	// Apparently, this form of CLI only authentication has been deprecated?

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n", authURL)
	fmt.Print("Enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange authorization code: %w", err)
	}

	return token, nil
}

// readTokenFromFile retrieves a token from a local file.
func readTokenFromFile(path string) (*oauth2.Token, error) {
	filep, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer filep.Close()

	token := &oauth2.Token{}
	if err := json.NewDecoder(filep).Decode(token); err != nil {
		return nil, err
	}

	return token, nil
}

// saveTokenToFile saves a token to a file.
func saveTokenToFile(path string, token *oauth2.Token) error {
	filep, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to create token file: %w", err)
	}

	if err := json.NewEncoder(filep).Encode(token); err != nil {
		filep.Close()
		return fmt.Errorf("unable to encode token: %w", err)
	}

	return filep.Close()
}
