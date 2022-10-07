# Monzo API Token Renewer

Simple app that kicks off the Monzo API OAuth dance, gets a token back, and then
(allegedly) renews that token. I've not run it long enough to find out if
that renewal actually works, but hey, let's find out.

UNLESS YOU ARE ME, READ AND UNDERSTAND THE CODE BEFORE YOU GO NEAR THIS

(This is generally good advice, but for an app that touches your bank account,
this is much more important)

## Usage

Populate `.env` file with appropriate values (or set the corresponding env vars)

Run with:

```
go run .
```

First thing you'll see is a URL you need to visit to get auth'd. Go to it.

Once auth'd, the redirect should take you to an incredibly thrilling localhost
webserver which will tell you something like:

```
Open the Monzo app to approve, and get some API scopes
```

You're almost there. Now open the Monzo app, allow access to your data, and
you're good.

Every hour (or whatever duration you specified in your env), you'll get logs
like these in your terminal:

```
Response (200): {"authenticated":true,"client_id":"REDACTED","user_id":"REDACTED"}

Response (200): {"balance":69,"total_balance":69,"balance_including_flexible_savings":69,"currency":"GBP","spend_today":69,"local_currency":"","local_exchange_rate":0,"local_spend":[{"spend_today":69,"currency":"USD"},{"spend_today":69,"currency":"GBP"}]}
```