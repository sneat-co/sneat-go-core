package models4emails

import (
	"testing"
	"time"
)

func TestEmail_Validate(t *testing.T) {
	type fields struct {
		To      string
		From    string
		Created time.Time
		Queued  *time.Time
		Sent    *time.Time
		Type    string
		Subject string
		Body    struct {
			Text string
			HTML string
		}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "should_pass",
			wantErr: false,
			fields: fields{
				From:    "sender@example.com",
				To:      "receiver@example.com",
				Subject: "Test email",
				Type:    "test",
				Created: time.Now(),
				Body: struct {
					Text string
					HTML string
				}{
					Text: "text body",
					HTML: "<p>HTML body</p>",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Email{
				To:      tt.fields.To,
				From:    tt.fields.From,
				Created: tt.fields.Created,
				Queued:  tt.fields.Queued,
				Sent:    tt.fields.Sent,
				Type:    tt.fields.Type,
				Subject: tt.fields.Subject,
				Body:    tt.fields.Body,
			}
			if err := v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
