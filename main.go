package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	client   *http.Client
	tok      *oauth2.Token
	hasToken bool
)

func main() {
	auth()
	apiLoop()
}

func auth() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     os.Getenv("MONZO_CLIENT_ID"),
		ClientSecret: os.Getenv("MONZO_CLIENT_SECRET"),
		Scopes:       []string{}, // Monzo API has no documented scopes
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.monzo.com/",
			TokenURL: "https://api.monzo.com/oauth2/token",
		},
		RedirectURL: "http://localhost:8080",
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(uuid.New().String())
	fmt.Printf("Visit the URL for the auth dialog: %v\n\n", url)

	codeChan := make(chan string)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !hasToken {
				codeChan <- r.URL.Query().Get("code")
				fmt.Fprintf(w, "Open the Monzo app to approve, and get some API scopes")
			} else {
				fmt.Fprintf(w, "App is running")
			}
		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	code := <-codeChan

	fmt.Printf("Attempting token exchange with '%v'\n\n", code)

	var err error
	tok, err = conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}
	hasToken = true

	fmt.Printf("We got token %s\n\n", tok)
	client = conf.Client(ctx, tok)
}

func apiLoop() {
	urls := []string{
		"https://api.monzo.com/ping/whoami",
		"https://api.monzo.com/balance?account_id=" + os.Getenv("MONZO_ACCOUNT_ID"),
	}

	sleepString := os.Getenv("SLEEP_TIMER")
	if sleepString == "" {
		sleepString = "1h"
	}
	sleepDuration, err := time.ParseDuration(sleepString)
	if err != nil {
		fmt.Printf("Could not parse %s as duration", sleepString)
		sleepDuration = time.Hour
	}

	for {
		for _, url := range urls {
			resp, err := client.Get(url)
			if err != nil {
				log.Fatal(err)
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Printf("Response (%v): %s\n", resp.StatusCode, bodyString)
		}

		fmt.Println()

		time.Sleep(sleepDuration)
	}
}
