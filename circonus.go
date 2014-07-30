package webhook

import (
	"strings"
	"time"
)

// CirconusTimeFormat represents the datetime format Circonus sends times in the webhook payload
const CirconusTimeFormat = "Mon, 02 Jan 2006 15:04:05"

// CirconusAccountTimezone represents the timezone the account is in since the webhook payload lacks
// this information. Exposed globally so it can be overridden.
var CirconusAccountTimezone = time.Local

// CirconusTime represents a time in the Circonus webhook payload
// This is needed to parse the format Circonus sends times in
type CirconusTime struct {
	time.Time
}

// UnmarshalJSON parse the time in the format expected from Circonus in the timezone specified globally
func (ct *CirconusTime) UnmarshalJSON(body []byte) (err error) {
	ct.Time, err = time.ParseInLocation(CirconusTimeFormat, strings.Trim(string(body), "\""), CirconusAccountTimezone)
	return err
}

// CirconusAlertValue represents the value of an alert in Circonus
// Circonus will encode string or a numeric value in the JSON payload, but we
// will just consider it a string to be safe.
type CirconusAlertValue struct {
	value string
}

// String returns the underlying string value
func (cav *CirconusAlertValue) String() string {
	return cav.value
}

// UnmarshalJSON interprets the value as a string regardless of whether it is a numeric value or not
func (cav *CirconusAlertValue) UnmarshalJSON(body []byte) (err error) {
	cav.value = strings.Trim(string(body), "\"")
	return nil
}

// CirconusPayload represents the JSON payload sent via Circonus webhook
type CirconusPayload struct {
	AccountName string           `json:"account_name"`
	Alerts      []*CirconusAlert `json:"alerts"`
}

// CirconusAlert represents an alert in the JSON payload sent via Circonus webhook
type CirconusAlert struct {
	ID         int                `json:"alert_id"`
	Severity   int                `json:"severity"`
	Value      CirconusAlertValue `json:"alert_value"`
	Time       CirconusTime       `json:"alert_time"`
	URL        string             `json:"alert_url"`
	Agent      string             `json:"agent"`
	CheckName  string             `json:"check_name"`
	MetricName string             `json:"metric_name"`

	ClearTime  CirconusTime       `json:"clear_time"`
	ClearValue CirconusAlertValue `json:"clear_value"`
}

// IsRecovery returns whether the alert represents a recovery rather than an alert
func (a *CirconusAlert) IsRecovery() bool {
	return !a.ClearTime.IsZero()
}
