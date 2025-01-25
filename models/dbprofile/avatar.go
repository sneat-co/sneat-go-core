package dbprofile

import (
	"github.com/strongo/validation"
	"strings"
)

// Avatar record
type Avatar struct {
	Provider     string `json:"provider" firestore:"provider"`
	FileID       string `json:"fileID" firestore:"fileID"`                                 // Telegram supplies this
	UniqueFileID string `json:"uniqueFileID,omitempty" firestore:"uniqueFileID,omitempty"` // Telegram supplies this
	URL          string `json:"url" firestore:"url"`
	Width        int    `json:"width,omitempty" firestore:"width,omitempty"`
	Height       int    `json:"height,omitempty" firestore:"height,omitempty"`
	Size         int    `json:"size,omitempty" firestore:"size,omitempty"`
}

func (v *Avatar) Equal(v2 *Avatar) bool {
	return v == nil && v2 == nil || v != nil && v2 != nil && *v == *v2
}

// Validate validates Avatar record
func (v *Avatar) Validate() error {
	if strings.TrimSpace(v.URL) == "" && v.FileID == "" && v.UniqueFileID == "" {
		return validation.NewErrRecordIsMissingRequiredField("url|fileID|uniqueFileID")
	}
	if url := strings.TrimSpace(v.URL); url != v.URL {
		return validation.NewErrBadRecordFieldValue("url", "must not have leading or trailing spaces")
	}
	if fileID := strings.TrimSpace(v.FileID); fileID != v.FileID {
		return validation.NewErrBadRecordFieldValue("fileID", "must not have leading or trailing spaces")
	}
	if uniqueFileID := strings.TrimSpace(v.UniqueFileID); uniqueFileID != v.UniqueFileID {
		return validation.NewErrBadRecordFieldValue("uniqueFileID", "must not have leading or trailing spaces")
	}
	if v.Provider == "" {
		return validation.NewErrRecordIsMissingRequiredField("provider")
	}
	if strings.Contains(v.Provider, " ") {
		return validation.NewErrBadRecordFieldValue("provider", "must not contain spaces")
	}
	if v.Width < 0 {
		return validation.NewErrBadRecordFieldValue("width", "must be positive")
	}
	if v.Height < 0 {
		return validation.NewErrBadRecordFieldValue("height", "must be positive")
	}
	if v.Size < 0 {
		return validation.NewErrBadRecordFieldValue("size", "must be positive")
	}
	return nil
}
