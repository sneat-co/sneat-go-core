// Package twiliosms is the ecosystem's single shared Twilio SMS-sending kernel.
//
// It is the ONE place the Sneat ecosystem builds a Twilio client, applies the
// +8→+7 phone-number fixup, and classifies a Twilio exception as permanent vs
// transient. Both the sneat-go notificator SMS provider
// (pkg/notificator/channels4notificator/twilio4notificator) and the debtus bot's
// sms package delegate here, so there is exactly one copy of this logic. Previously
// the fixup + classification were duplicated: originally in debtus's
// debtus/sms/send.go, then ported (copied) into the notificator provider. This
// package collapses both onto a shared kernel.
//
// It lives in sneat-go-core — the one module BOTH sneat-go (the host) and the
// debtus/backend extension already depend on — because an extension module must
// not import the host sneat-go (import-cycle + dependency-weight in its
// public-only CI). A neutral low-level home is the only thing both can import.
//
// The package is credential-agnostic about identity: callers pass Credentials.
// The generalized env var names (Env*) and CredentialsFromEnv give every consumer
// the SAME ecosystem-wide Twilio account/number when they read the environment —
// which is how debtus now sends from the shared Sneat Twilio identity rather than
// its own. A consumer that needs a different identity (e.g. debtus's Twilio test
// sandbox, or its own delivery-status webhook) just builds Credentials itself.
package twiliosms

import (
	"net/http"
	"strings"

	"github.com/strongo/gotwilio"
)

// Generalized ecosystem Twilio env var names. Flat naming matching
// env_variables_template.yaml (AWS_*, ANTHROPIC_*, TMDB_*); deliberately NOT
// debtus's historical TWILIO_LIVE_* names — a single set of names means a single
// shared identity across the ecosystem.
const (
	EnvAccountSID        = "TWILIO_ACCOUNT_SID"         // required
	EnvAuthToken         = "TWILIO_AUTH_TOKEN"          // required
	EnvFromNumber        = "TWILIO_FROM_NUMBER"         // required (E.164 sender number)
	EnvApplicationSID    = "TWILIO_APPLICATION_SID"     // optional
	EnvStatusCallbackURL = "TWILIO_STATUS_CALLBACK_URL" // optional (delivery-status webhook)
)

// Credentials carries everything needed to send an SMS: the account identity, the
// sender number, and the optional application SID + delivery-status callback URL.
type Credentials struct {
	AccountSID        string
	AuthToken         string
	FromNumber        string
	ApplicationSID    string
	StatusCallbackURL string
}

// CredentialsFromEnv reads the generalized TWILIO_* env vars. ok is false (and
// Credentials zero) when any REQUIRED credential (account SID, auth token, from
// number) is absent, so a caller can leave its SMS path unconfigured rather than
// building a broken sender. It never panics and never makes a network call.
func CredentialsFromEnv(getenv func(string) string) (creds Credentials, ok bool) {
	accountSID := strings.TrimSpace(getenv(EnvAccountSID))
	authToken := strings.TrimSpace(getenv(EnvAuthToken))
	fromNumber := strings.TrimSpace(getenv(EnvFromNumber))
	if accountSID == "" || authToken == "" || fromNumber == "" {
		return Credentials{}, false
	}
	return Credentials{
		AccountSID:        accountSID,
		AuthToken:         authToken,
		FromNumber:        fromNumber,
		ApplicationSID:    strings.TrimSpace(getenv(EnvApplicationSID)),
		StatusCallbackURL: strings.TrimSpace(getenv(EnvStatusCallbackURL)),
	}, true
}

// Client is the subset of gotwilio.Twilio this package uses. It is an interface
// so consumers' tests can inject a fake without making live Twilio calls (the
// production client is built by NewClient).
type Client interface {
	SendSMS(from, to, body, statusCallback, applicationSid string) (*gotwilio.SmsResponse, *gotwilio.Exception, error)
}

// NewClient builds the production gotwilio client for the given credentials. When
// httpClient is nil gotwilio uses http.DefaultClient. It never makes a network
// call, so it is safe to build eagerly.
func NewClient(creds Credentials, httpClient *http.Client) Client {
	return gotwilio.NewTwilioClientCustomHTTP(creds.AccountSID, creds.AuthToken, httpClient)
}

// Send sends an SMS through client using creds' FromNumber / StatusCallbackURL /
// ApplicationSID, applying the +8→+7 fixup (see below). It returns gotwilio's
// response and exception VERBATIM so each caller can build its own result shape
// (the notificator maps to a permanent/transient flag via IsPermanentException;
// debtus surfaces the exception to its bot UI via TwilioExceptionToMessage).
//
// The +8→+7 fixup: some "+8..." numbers are really "+7..." (Russia/Kazakhstan)
// and Twilio rejects them with error 21211; on that exact case (12-digit, "+8"
// prefix) Send retries once with the corrected prefix before giving up.
func Send(client Client, creds Credentials, to, text string) (*gotwilio.SmsResponse, *gotwilio.Exception, error) {
	resp, ex, err := client.SendSMS(creds.FromNumber, to, text, creds.StatusCallbackURL, creds.ApplicationSID)
	if err != nil {
		return nil, nil, err
	}
	if ex != nil && ex.Code == 21211 && len(to) == 12 && strings.HasPrefix(to, "+8") {
		corrected := strings.Replace(to, "+8", "+7", 1)
		return client.SendSMS(creds.FromNumber, corrected, text, creds.StatusCallbackURL, creds.ApplicationSID)
	}
	return resp, ex, err
}

// permanentCodes are the Twilio error codes that mean "this number will never
// accept this message" — a permanent classification lets a caller advance a
// fallback chain instead of retrying a doomed send.
//
//	21211 — invalid phone number          https://www.twilio.com/docs/errors/21211
//	21614 — not an SMS-capable/mobile no.  https://www.twilio.com/docs/errors/21614
//	21612 — not reachable from From no.    https://www.twilio.com/docs/errors/21612
//	21408 — SMS not enabled for region     https://www.twilio.com/docs/errors/21408
//	21610 — From/To pair is blacklisted    https://www.twilio.com/docs/errors/21610
var permanentCodes = map[int]bool{
	21211: true,
	21614: true,
	21612: true,
	21408: true,
	21610: true,
}

// IsPermanentException classifies a Twilio exception as permanent (true) or
// transient (false). Permanent for the known bad-number codes above, and for any
// generic 4xx client error (which the same retry will never fix) EXCEPT HTTP 429
// rate-limiting. Transient for 429, 5xx and anything else (network hiccups),
// which the caller should retry. A nil exception is not permanent.
func IsPermanentException(ex *gotwilio.Exception) bool {
	if ex == nil {
		return false
	}
	if permanentCodes[ex.Code] {
		return true
	}
	if ex.Status >= 400 && ex.Status < 500 && ex.Status != http.StatusTooManyRequests {
		return true
	}
	return false
}
