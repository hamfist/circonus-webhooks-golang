package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/andybons/hipchat"
	"github.com/gorilla/mux"
)

// HipchatHandler handles webhook requests by forwarding them to the Hipchat API
type HipchatHandler struct {
	HipchatClient    *hipchat.Client
	AlertTemplate    *template.Template
	RecoveryTemplate *template.Template
	AlertColor       string
	RecoveryColor    string
	From             string
}

// NewHipchatHandler creates a new by reading the environment
func NewHipchatHandler() *HipchatHandler {
	alertTemplate := os.Getenv("CIRCONUS_WEBHOOK_PROXY_HIPCHAT_ALERT_TEMPLATE")
	if alertTemplate == "" {
		alertTemplate = `Severity {{.Severity}} alert triggered by {{.CheckName}} ({{.MetricName}}: {{.Value}}). {{.URL}}`
	}

	recoveryTemplate := os.Getenv("CIRCONUS_WEBHOOK_PROXY_HIPCHAT_RECOVERY_TEMPLATE")
	if recoveryTemplate == "" {
		recoveryTemplate = `Recovery of {{.CheckName}} ({{.MetricName}}: {{.Value}}). {{.URL}}`
	}

	hh := &HipchatHandler{
		HipchatClient:    &hipchat.Client{AuthToken: os.Getenv("CIRCONUS_WEBHOOK_PROXY_HIPCHAT_API_TOKEN")},
		AlertTemplate:    template.Must(template.New("alert").Parse(alertTemplate)),
		RecoveryTemplate: template.Must(template.New("recovery").Parse(recoveryTemplate)),
		AlertColor:       os.Getenv("CIRCONUS_WEBHOOK_PROXY_HIPCHAT_ALERT_COLOR"),
		RecoveryColor:    os.Getenv("CIRCONUS_WEBHOOK_PROXY_HIPCHAT_RECOVERY_COLOR"),
		From:             os.Getenv("CIRCONUS_WEBHOOK_PROXY_HIPCHAT_FROM"),
	}

	if hh.AlertColor == "" {
		hh.AlertColor = hipchat.ColorRed
	}

	if hh.RecoveryColor == "" {
		hh.RecoveryColor = hipchat.ColorGreen
	}

	if hh.From == "" {
		hh.From = "Circonus"
	}

	return hh
}

// Name returns the name of the handler
func (hh *HipchatHandler) Name() string {
	return "Hipchat"
}

// Route returns the route to mount the handler at
func (hh *HipchatHandler) Route() string {
	return "/hipchat/{room}?format=json"
}

// Register adds the Hipchat handler to the router
func (hh *HipchatHandler) Register(r *mux.Router) {
	r.Handle("/hipchat/{room}", hh).Methods("POST").Queries("format", "json")
}

// Usage returns a string to show to the user about configuration of this hook
func (hh *HipchatHandler) Usage() string {
	return `
	Sends alerts to the Hipchat room identified in the URL as {room}.

	Requires the following environment variables:
	CIRCONUS_WEBHOOK_PROXY_HIPCHAT_API_TOKEN: API (version 1) token

	Permits the following environment variables:
	CIRCONUS_WEBHOOK_PROXY_HIPCHAT_ALERT_TEMPLATE: Golang text template to format alert (see circonus.go)
		Defaults to: Severity {{.Severity}} alert triggered by {{.CheckName}} ({{.MetricName}}: {{.Value}}). {{.URL}}
	CIRCONUS_WEBHOOK_PROXY_HIPCHAT_RECOVERY_TEMPLATE: Golang text template to format recovery (see circonus.go)
		Defaults to: Recovery of {{.CheckName}} ({{.MetricName}}: {{.Value}}). {{.URL}}
	CIRCONUS_WEBHOOK_PROXY_HIPCHAT_ALERT_COLOR: Color of Hipchat message for alerts (defaults to red)
	CIRCONUS_WEBHOOK_PROXY_HIPCHAT_RECOVERY_COLOR: Color of Hipchat message for recovery (defaults to green)
	CIRCONUS_WEBHOOK_PROXY_HIPCHAT_FROM: User to use as "From" (defaults to Circonus)
	`
}

// ServeHTTP parses Circonus webhook payload and sends messages to Hipchat
func (hh *HipchatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	circonusPayload := &CirconusPayload{}

	err := json.NewDecoder(r.Body).Decode(circonusPayload)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse request body: %s", err), http.StatusBadRequest)
	}

	for _, alert := range circonusPayload.Alerts {
		var (
			tpl   *template.Template
			color string
			buf   bytes.Buffer
		)

		if alert.IsRecovery() {
			tpl = hh.RecoveryTemplate
			color = hh.RecoveryColor
		} else {
			tpl = hh.AlertTemplate
			color = hh.AlertColor
		}

		err := tpl.Execute(&buf, alert)
		if err != nil {
			http.Error(w, fmt.Sprintf("error executing alert template", err), http.StatusInternalServerError)
			return
		}

		req := hipchat.MessageRequest{
			RoomId:        vars["room"],
			From:          hh.From,
			Message:       buf.String(),
			Color:         color,
			MessageFormat: hipchat.FormatText,
			Notify:        true,
		}

		if err := hh.HipchatClient.PostMessage(req); err != nil {
			http.Error(w, fmt.Sprintf("error sending message to hipchat: %v", err), http.StatusInternalServerError)
		}
	}
}
