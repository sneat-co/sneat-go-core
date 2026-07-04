package twiliosms

import (
	"net/http"
	"testing"

	"github.com/strongo/gotwilio"
)

// fakeClient records calls and returns scripted results per call index.
type fakeClient struct {
	calls   []string // "to" of each SendSMS call, in order
	results []struct {
		resp *gotwilio.SmsResponse
		ex   *gotwilio.Exception
		err  error
	}
}

func (f *fakeClient) SendSMS(_, to, _, _, _ string) (*gotwilio.SmsResponse, *gotwilio.Exception, error) {
	i := len(f.calls)
	f.calls = append(f.calls, to)
	if i >= len(f.results) {
		return nil, nil, nil
	}
	r := f.results[i]
	return r.resp, r.ex, r.err
}

func TestCredentialsFromEnv(t *testing.T) {
	full := map[string]string{
		EnvAccountSID:        "AC123",
		EnvAuthToken:         "tok",
		EnvFromNumber:        "+15551230000",
		EnvApplicationSID:    "AP1",
		EnvStatusCallbackURL: "https://cb",
	}
	getenv := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}

	creds, ok := CredentialsFromEnv(getenv(full))
	if !ok {
		t.Fatal("expected ok with full env")
	}
	if creds.AccountSID != "AC123" || creds.FromNumber != "+15551230000" || creds.ApplicationSID != "AP1" || creds.StatusCallbackURL != "https://cb" {
		t.Errorf("unexpected creds: %+v", creds)
	}

	// Each required field missing → not ok.
	for _, missing := range []string{EnvAccountSID, EnvAuthToken, EnvFromNumber} {
		m := map[string]string{}
		for k, v := range full {
			m[k] = v
		}
		delete(m, missing)
		if _, ok := CredentialsFromEnv(getenv(m)); ok {
			t.Errorf("expected not-ok when %s is missing", missing)
		}
	}

	// Whitespace-only required field is treated as missing.
	m := map[string]string{EnvAccountSID: "  ", EnvAuthToken: "tok", EnvFromNumber: "+1"}
	if _, ok := CredentialsFromEnv(getenv(m)); ok {
		t.Error("expected not-ok when account SID is whitespace")
	}
}

func TestSend_PlainSuccess(t *testing.T) {
	client := &fakeClient{results: []struct {
		resp *gotwilio.SmsResponse
		ex   *gotwilio.Exception
		err  error
	}{
		{resp: &gotwilio.SmsResponse{Sid: "SM1"}},
	}}
	creds := Credentials{FromNumber: "+15551230000"}
	resp, ex, err := Send(client, creds, "+12025550100", "hi")
	if err != nil || ex != nil || resp == nil || resp.Sid != "SM1" {
		t.Fatalf("resp=%v ex=%v err=%v", resp, ex, err)
	}
	if len(client.calls) != 1 || client.calls[0] != "+12025550100" {
		t.Errorf("calls=%v, want one call to the original number", client.calls)
	}
}

func TestSend_TransportErrorNoRetry(t *testing.T) {
	client := &fakeClient{results: []struct {
		resp *gotwilio.SmsResponse
		ex   *gotwilio.Exception
		err  error
	}{
		{err: http.ErrHandlerTimeout},
	}}
	_, _, err := Send(client, Credentials{FromNumber: "+1"}, "+89991234567", "hi")
	if err == nil {
		t.Fatal("expected transport error propagated")
	}
	if len(client.calls) != 1 {
		t.Errorf("transport error must not retry; calls=%v", client.calls)
	}
}

func TestSend_Plus8ToPlus7Fixup(t *testing.T) {
	client := &fakeClient{results: []struct {
		resp *gotwilio.SmsResponse
		ex   *gotwilio.Exception
		err  error
	}{
		{ex: &gotwilio.Exception{Code: 21211, Message: "invalid"}}, // first: +8 rejected
		{resp: &gotwilio.SmsResponse{Sid: "SM2"}},                  // retry with +7 succeeds
	}}
	resp, ex, err := Send(client, Credentials{FromNumber: "+1"}, "+89991234567", "hi")
	if err != nil || ex != nil || resp == nil || resp.Sid != "SM2" {
		t.Fatalf("resp=%v ex=%v err=%v", resp, ex, err)
	}
	if len(client.calls) != 2 {
		t.Fatalf("expected 2 calls (original + fixup), got %v", client.calls)
	}
	if client.calls[1] != "+79991234567" {
		t.Errorf("retry number = %q, want +79991234567", client.calls[1])
	}
}

func TestSend_NoFixupWhenNotPlus8Or21211(t *testing.T) {
	// 21211 but not a +8 12-digit number → no retry.
	client := &fakeClient{results: []struct {
		resp *gotwilio.SmsResponse
		ex   *gotwilio.Exception
		err  error
	}{
		{ex: &gotwilio.Exception{Code: 21211}},
	}}
	_, ex, _ := Send(client, Credentials{FromNumber: "+1"}, "+12025550100", "hi")
	if ex == nil {
		t.Error("expected the 21211 exception to be returned unretried")
	}
	if len(client.calls) != 1 {
		t.Errorf("non-+8 number must not retry; calls=%v", client.calls)
	}
}

func TestIsPermanentException(t *testing.T) {
	tests := []struct {
		name string
		ex   *gotwilio.Exception
		want bool
	}{
		{"nil", nil, false},
		{"21211 bad number", &gotwilio.Exception{Code: 21211}, true},
		{"21610 blacklist", &gotwilio.Exception{Code: 21610}, true},
		{"generic 4xx", &gotwilio.Exception{Status: 400}, true},
		{"429 rate limit is transient", &gotwilio.Exception{Status: http.StatusTooManyRequests}, false},
		{"5xx transient", &gotwilio.Exception{Status: 503}, false},
		{"unknown code, no status", &gotwilio.Exception{Code: 99999}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPermanentException(tt.ex); got != tt.want {
				t.Errorf("IsPermanentException(%+v) = %v, want %v", tt.ex, got, tt.want)
			}
		})
	}
}
