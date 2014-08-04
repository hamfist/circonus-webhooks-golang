package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/modcloth-labs/circonus-webhooks-golang"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
)

func main() {
	webhooks := []webhook.Handler{
		webhook.NewHipchatHandler(),
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, `
Contains plugins for proxying Circonus webhook requests to non-natively supported services (e.g. Hipchat).
See http://www.circonus.com/webhook-notifications for more information.

Reads the following environment configuration globally:
PORT: Port to listen on (defaults to 3000).
CIRCONUS_WEBHOOK_PROXY_ACCOUNT_TIMEZONE: Timezone your Circonus organization
	account is in (this information is not included in the webhook payload).
	Expects locations to be in IANA Time Zone format.
	Defaults to local system time.
		`)

		fmt.Fprintf(os.Stderr, "\nPlugins:\n\n")
		for _, handler := range webhooks {
			fmt.Fprintf(os.Stderr, "%s - %s\n", handler.Name(), handler.Route())
			fmt.Fprintln(os.Stderr, handler.Usage())
		}
	}

	flag.Parse()

	accountTimezoneString := os.Getenv("CIRCONUS_WEBHOOK_PROXY_ACCOUNT_TIMEZONE")
	if accountTimezoneString != "" {
		var err error
		webhook.CirconusAccountTimezone, err = time.LoadLocation(accountTimezoneString)
		if err != nil {
			fmt.Printf("Could not parse timezone location: %s\n", accountTimezoneString)
			flag.Usage()
			os.Exit(1)
		}
	}

	r := mux.NewRouter()
	for _, handler := range webhooks {
		handler.Register(r)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	n := negroni.Classic()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(r)
	n.Run(fmt.Sprintf(":%s", port))
}
