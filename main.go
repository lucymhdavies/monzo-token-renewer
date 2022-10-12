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
		RedirectURL: getEnv("REDIRECT_URL", "http://localhost:8080"),
	}

	// TODO: Check if we have a token in Vault already we can reuse
	// if we do, check if it's valid

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(uuid.New().String())
	fmt.Printf("Visit the URL for the auth dialog: %v\n\n", url)

	codeChan := make(chan string)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// TODO: first page should be the link to the login
			// After login, page should check if we have API scopes yet (and prompt to use Monzo app if not)
			// Then we should check the token is still valid, and display that
			// and if the token is no longer valid, we should go back to the login page (or kill the app)

			if !hasToken {
				if r.URL.Query().Get("code") == "" {
					fmt.Fprintf(w, "Visit this URL to initiate auth:\n%s", url)
				} else {
					codeChan <- r.URL.Query().Get("code")
					fmt.Fprintf(w, "Code accepted. Open the Monzo app to approve, and get some API scopes")
				}
			} else {
				fmt.Fprintf(w, "App is running")
			}
		})
		// TODO: get port from env var
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

	// TODO: store this token in Vault

	// TODO: if renewing the token results in the token changing... update Vault
}

func apiLoop() {
	// TODO: modify this so that we can list pots from the monzo API, then
	// get the balance for each of those pots, and then store all of that
	// in Vault.
	// in that way, we never need to share the API token outside of this application

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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
