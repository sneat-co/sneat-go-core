package dbmodels

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/crediterra/money"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/strongo/strongoapp/with"
)

type dummyValidatable struct {
	fail bool
}

func (d dummyValidatable) Validate() error {
	if d.fail {
		return errors.New("dummy error")
	}
	return nil
}

func TestAddress_ValidateCoverage(t *testing.T) {
	var nilAddr *Address
	if err := nilAddr.Validate(); err != nil {
		t.Fatalf("expected nil error for nil Address, got %v", err)
	}

	addr := &Address{}
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on empty countryID")
	}

	addr.CountryID = "XYZ"
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on invalid countryID")
	}

	addr.CountryID = "US"
	addr.ZipCode = " 12345 "
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on untrimmed zipCode")
	}

	addr.ZipCode = "12345"
	addr.City = strings.Repeat("a", 86)
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on city > 85 chars")
	}

	addr.City = "New York"
	addr.State = strings.Repeat("b", 31)
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on state > 30 chars")
	}

	addr.State = "NY"
	addr.Lines = " 123 Main St "
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on untrimmed lines")
	}

	addr.Lines = strings.Repeat("c", 1001)
	if err := addr.Validate(); err == nil {
		t.Fatal("expected error on lines > 1000 chars")
	}

	addr.Lines = "123 Main St"
	if err := addr.Validate(); err != nil {
		t.Fatalf("unexpected error on valid Address: %v", err)
	}
}

func TestByUser_ValidateCoverage(t *testing.T) {
	u := &ByUser{}
	if err := u.Validate(); err == nil {
		t.Fatal("expected error on empty UID")
	}
	u.UID = "user1"
	if err := u.Validate(); err != nil {
		t.Fatalf("unexpected error on valid UID: %v", err)
	}
}

func TestCreatedInfo_ValidateCoverage(t *testing.T) {
	ci := CreatedInfo{}
	if err := ci.Validate(); err == nil {
		t.Fatal("expected error on invalid Client in CreatedInfo")
	}
	ci.Client.HostOrApp = "test-host"
	ci.Client.RemoteAddr = "127.0.0.1"
	if err := ci.Validate(); err != nil {
		t.Fatalf("unexpected error on valid CreatedInfo: %v", err)
	}
}

func TestWithCustomFields_ValidateCoverage(t *testing.T) {
	wcf := &WithCustomFields{}
	if err := wcf.Validate(); err != nil {
		t.Fatalf("unexpected error on empty WithCustomFields: %v", err)
	}

	wcf.FieldsDate = map[string]string{"due": "invalid-date"}
	if err := wcf.Validate(); err == nil {
		t.Fatal("expected error on invalid date in FieldsDate")
	}

	wcf.FieldsDate = map[string]string{"due": "2025-01-01"}
	wcf.FieldsAmount = map[string]money.Amount{"fee": {Currency: "INVALID"}}
	if err := wcf.Validate(); err == nil {
		t.Fatal("expected error on invalid amount in FieldsAmount")
	}

	wcf.FieldsAmount = map[string]money.Amount{"fee": {Currency: money.CurrencyUSD}}
	if err := wcf.Validate(); err != nil {
		t.Fatalf("unexpected error on valid WithCustomFields: %v", err)
	}
}

func TestModified_ValidateCoverage(t *testing.T) {
	m := Modified{}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error on empty By in Modified")
	}
	m.By = "user1"
	if err := m.Validate(); err == nil {
		t.Fatal("expected error on zero At in Modified")
	}
	m.At = time.Now()
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error on valid Modified: %v", err)
	}
}

func TestWithModified_Coverage(t *testing.T) {
	now := time.Now()
	wm := NewWithModified(now, "creator")
	if wm.GetCreatedBy() != "creator" || wm.UpdatedBy != "creator" {
		t.Fatalf("unexpected NewWithModified values")
	}
	if err := wm.Validate(); err != nil {
		t.Fatalf("unexpected error on valid WithModified: %v", err)
	}

	wm.MarkAsUpdated("updater")
	if wm.UpdatedBy != "updater" {
		t.Fatalf("expected UpdatedBy=updater, got %s", wm.UpdatedBy)
	}

	// Validation errors
	badWm := WithModified{}
	if err := badWm.Validate(); err == nil {
		t.Fatal("expected errors on zero WithModified")
	}

	// Deleted validation error
	delWm := NewWithModified(now, "user1")
	delWm.DeletedAt = now
	delWm.DeletedBy = ""
	if err := delWm.Validate(); err == nil {
		t.Fatal("expected error on deleted fields invalid")
	}
}

func TestRemoteClientInfo_Coverage(t *testing.T) {
	rci := RemoteClientInfo{}
	if err := rci.Validate(); err == nil {
		t.Fatal("expected error on empty HostOrApp")
	}

	rci.HostOrApp = "web"
	if err := rci.Validate(); err == nil {
		t.Fatal("expected error on empty RemoteAddr without @")
	}

	// Messenger host with @
	rci.HostOrApp = "telegram@bot1"
	if err := rci.Validate(); err != nil {
		t.Fatalf("unexpected error for messenger host without remote addr: %v", err)
	}

	// Bad GeoCityPoint
	rci.GeoCityPoint = &GeoPoint{Lat: 100, Lng: 0}
	if err := rci.Validate(); err == nil {
		t.Fatal("expected error on invalid GeoPoint")
	}

	rci.GeoCityPoint = &GeoPoint{Lat: 45, Lng: 45}
	if err := rci.Validate(); err != nil {
		t.Fatalf("unexpected error on valid GeoPoint: %v", err)
	}

	// GeoPoint.Valid tests
	if (GeoPoint{Lat: -91, Lng: 0}).Valid() {
		t.Fatal("expected invalid for Lat < -90")
	}
	if (GeoPoint{Lat: 91, Lng: 0}).Valid() {
		t.Fatal("expected invalid for Lat > 90")
	}
	if (GeoPoint{Lat: 0, Lng: -181}).Valid() {
		t.Fatal("expected invalid for Lng < -180")
	}
	if (GeoPoint{Lat: 0, Lng: 181}).Valid() {
		t.Fatal("expected invalid for Lng > 180")
	}
	if !(GeoPoint{Lat: 0, Lng: 0}).Valid() {
		t.Fatal("expected valid for (0, 0)")
	}
}

func TestSpaceItemID_Coverage(t *testing.T) {
	sid := NewSpaceItemID("sp1", "it1")
	if string(sid) != "sp1_it1" {
		t.Fatalf("unexpected sid: %s", sid)
	}
	if sid.SpaceID() != "sp1" {
		t.Fatalf("expected SpaceID 'sp1', got %s", sid.SpaceID())
	}
	if sid.ItemID() != "it1" {
		t.Fatalf("expected ItemID 'it1', got %s", sid.ItemID())
	}

	// SpaceItemID.Validate
	if err := SpaceItemID("").Validate(); err == nil {
		t.Fatal("expected error on empty SpaceItemID")
	}
	if err := SpaceItemID("nosep").Validate(); err == nil {
		t.Fatal("expected error on missing separator")
	}
	if err := SpaceItemID("_item").Validate(); err == nil {
		t.Fatal("expected error on separator at 0")
	}
	if err := SpaceItemID("space_").Validate(); err == nil {
		t.Fatal("expected error on separator at end")
	}
	if err := SpaceItemID("sp_it_extra").Validate(); err == nil {
		t.Fatal("expected error on multiple separators")
	}
	if err := SpaceItemID("sp_it").Validate(); err != nil {
		t.Fatalf("unexpected error on valid SpaceItemID: %v", err)
	}
}

func TestTimestamp_ValidateCoverage(t *testing.T) {
	ts := &Timestamp{}
	if err := ts.Validate(); err == nil {
		t.Fatal("expected error on zero Time in Timestamp")
	}
	ts.Time = time.Now()
	if err := ts.Validate(); err == nil {
		t.Fatal("expected error on empty Operation in Timestamp")
	}
	ts.Operation = "start"
	if err := ts.Validate(); err != nil {
		t.Fatalf("unexpected error on valid Timestamp without By: %v", err)
	}
	ts.By = &ByUser{UID: ""}
	if err := ts.Validate(); err == nil {
		t.Fatal("expected error on invalid By in Timestamp")
	}
	ts.By.UID = "user1"
	if err := ts.Validate(); err != nil {
		t.Fatalf("unexpected error on valid Timestamp with By: %v", err)
	}
}

func TestValidateWithIdsAndBriefs_Coverage(t *testing.T) {
	// Empty ids
	if err := ValidateWithIdsAndBriefs("ids", "briefs", nil, map[string]dummyValidatable{}); err == nil {
		t.Fatal("expected error on empty ids")
	}

	// First id != "*"
	if err := ValidateWithIdsAndBriefs("ids", "briefs", []string{"foo"}, map[string]dummyValidatable{}); err == nil {
		t.Fatal("expected error on first id != '*'")
	}

	// Brief missing for an id
	if err := ValidateWithIdsAndBriefs("ids", "briefs", []string{"*", "id1"}, map[string]dummyValidatable{}); err == nil {
		t.Fatal("expected error on missing brief")
	}

	// Brief has id not in ids[1:]
	briefsExtra := map[string]dummyValidatable{
		"id1": {fail: false},
		"id2": {fail: false},
	}
	if err := ValidateWithIdsAndBriefs("ids", "briefs", []string{"*", "id1"}, briefsExtra); err == nil {
		t.Fatal("expected error on extra brief not in ids")
	}

	// Brief validate fails
	briefsBad := map[string]dummyValidatable{
		"id1": {fail: true},
	}
	if err := ValidateWithIdsAndBriefs("ids", "briefs", []string{"*", "id1"}, briefsBad); err == nil {
		t.Fatal("expected error on failing brief validate")
	}

	// Valid
	briefsOK := map[string]dummyValidatable{
		"id1": {fail: false},
	}
	if err := ValidateWithIdsAndBriefs("ids", "briefs", []string{"*", "id1"}, briefsOK); err != nil {
		t.Fatalf("unexpected error on valid ValidateWithIdsAndBriefs: %v", err)
	}
}

func TestValidateTitle_Coverage(t *testing.T) {
	if err := ValidateTitle(""); err == nil {
		t.Fatal("expected error on empty title")
	}
	if err := ValidateTitle("   "); err == nil {
		t.Fatal("expected error on whitespace title")
	}
	if err := ValidateTitle(" title"); err == nil {
		t.Fatal("expected error on untrimmed title")
	}
	if err := ValidateTitle("Valid Title"); err != nil {
		t.Fatalf("unexpected error on valid title: %v", err)
	}
}

func TestVersioned_Coverage(t *testing.T) {
	wv := &WithVersion{Version: 0}
	if err := wv.Validate(); err == nil {
		t.Fatal("expected error on version < 1")
	}
	wv.IncreaseVersion()
	if wv.Version != 1 {
		t.Fatalf("expected version 1, got %d", wv.Version)
	}
	if err := wv.Validate(); err != nil {
		t.Fatalf("unexpected error on version 1: %v", err)
	}
	updates := wv.GetUpdates()
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}

	wuv := &WithUpdatedAndVersion{}
	if err := wuv.Validate(); err == nil {
		t.Fatal("expected error on zero WithUpdatedAndVersion")
	}

	now := time.Now()
	newVer := wuv.IncreaseVersion(now, "updater")
	if newVer != 1 {
		t.Fatalf("expected newVer 1, got %d", newVer)
	}
	if wuv.UpdatedBy != "updater" || !wuv.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected values in WithUpdatedAndVersion")
	}
	if err := wuv.Validate(); err != nil {
		t.Fatalf("unexpected error on valid WithUpdatedAndVersion: %v", err)
	}
	wuvUpdates := wuv.GetUpdates()
	if len(wuvUpdates) == 0 {
		t.Fatal("expected updates from WithUpdatedAndVersion")
	}
}

func TestDtoWithID_Coverage(t *testing.T) {
	var nilDto *DtoWithID[dummyValidatable]
	if err := nilDto.Validate(); err != nil {
		t.Fatalf("expected nil error for nil DtoWithID, got %v", err)
	}

	dto := &DtoWithID[dummyValidatable]{}
	if err := dto.Validate(); err == nil {
		t.Fatal("expected error on empty ID")
	}

	dto.ID = "id1"
	dto.Data = dummyValidatable{fail: true}
	if err := dto.Validate(); err == nil {
		t.Fatal("expected error on failing Data in DtoWithID")
	}

	dto.Data = dummyValidatable{fail: false}
	if err := dto.Validate(); err != nil {
		t.Fatalf("unexpected error on valid DtoWithID: %v", err)
	}
}

func TestWithLastCurrencies_Coverage(t *testing.T) {
	wlc := &WithLastCurrencies{
		LastCurrencies: []money.CurrencyCode{money.CurrencyUSD, money.CurrencyEUR},
	}
	got := wlc.GetLastCurrencies()
	if len(got) != 2 || got[0] != money.CurrencyUSD {
		t.Fatalf("unexpected GetLastCurrencies result: %v", got)
	}

	// Unknown currency
	if _, err := wlc.SetLastCurrency("UNKNOWN"); err == nil {
		t.Fatal("expected error on unknown currency")
	}

	// Already first
	updates, err := wlc.SetLastCurrency(money.CurrencyUSD)
	if err != nil || updates != nil {
		t.Fatalf("expected nil updates when already first, got %v, %v", updates, err)
	}

	// In list at index 1 -> moved to front
	updates, err = wlc.SetLastCurrency(money.CurrencyEUR)
	if err != nil {
		t.Fatalf("unexpected error moving EUR to front: %v", err)
	}
	if wlc.LastCurrencies[0] != money.CurrencyEUR || wlc.LastCurrencies[1] != money.CurrencyUSD {
		t.Fatalf("expected EUR first, got %v", wlc.LastCurrencies)
	}

	// Not in list, added to front
	wlc.LastCurrencies = []money.CurrencyCode{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
	}
	updates, err = wlc.SetLastCurrency(money.CurrencyJPY)
	if err != nil || len(updates) != 1 {
		t.Fatalf("unexpected result adding JPY: %v, %v", updates, err)
	}
	if len(wlc.LastCurrencies) != 10 || wlc.LastCurrencies[0] != money.CurrencyJPY {
		t.Fatalf("expected 10 currencies with JPY first, got len=%d, first=%s", len(wlc.LastCurrencies), wlc.LastCurrencies[0])
	}
}

func TestWithPreferredLocale_Coverage(t *testing.T) {
	wpl := &WithPreferredLocale{PreferredLocale: "en-US"}
	if wpl.GetPreferredLocale() != "en-US" {
		t.Fatalf("expected en-US, got %s", wpl.GetPreferredLocale())
	}

	// Invalid length
	if _, err := wpl.SetPreferredLocale("en"); err == nil {
		t.Fatal("expected error on locale length 2")
	}

	// Valid length 5
	updates, err := wpl.SetPreferredLocale("fr-FR")
	if err != nil || len(updates) != 1 {
		t.Fatalf("unexpected result: %v, %v", updates, err)
	}
	if wpl.PreferredLocale != "fr-FR" {
		t.Fatalf("expected fr-FR, got %s", wpl.PreferredLocale)
	}

	// Valid length 0
	updates, err = wpl.SetPreferredLocale("")
	if err != nil || len(updates) != 1 {
		t.Fatalf("unexpected result: %v, %v", updates, err)
	}
	if wpl.PreferredLocale != "" {
		t.Fatalf("expected empty, got %s", wpl.PreferredLocale)
	}
}

func TestWithTimezone_ValidateCoverage(t *testing.T) {
	wtz := &WithTimezone{}
	if err := wtz.Validate(); err != nil {
		t.Fatalf("expected nil for nil Timezone, got %v", err)
	}

	wtz.Timezone = &Timezone{Iana: "UTC"}
	if err := wtz.Validate(); err != nil {
		t.Fatalf("unexpected error for UTC timezone: %v", err)
	}

	wtz.Timezone = &Timezone{Iana: "bad"}
	if err := wtz.Validate(); err == nil {
		t.Fatal("expected error on invalid timezone Iana")
	}
}

func TestWithSpaceDates_Comprehensive(t *testing.T) {
	wsd := &WithSpaceDates{}

	// Invalid SpaceIDs
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on invalid SpaceIDs")
	}

	wsd.SpaceIDs = []coretypes.SpaceID{"s1"}
	// Invalid Dates
	wsd.DatesFields = with.DatesFields{Dates: []string{"invalid-date"}}
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on invalid Dates")
	}

	wsd.DatesFields = with.DatesFields{Dates: []string{"2025-01-01"}}
	// Populates SpaceDates
	if err := wsd.Validate(); err != nil {
		t.Fatalf("unexpected error on valid WithSpaceDates: %v", err)
	}
	if len(wsd.SpaceDates) != 1 || wsd.SpaceDates[0] != "s1:2025-01-01" {
		t.Fatalf("unexpected SpaceDates: %v", wsd.SpaceDates)
	}

	// Length mismatch
	wsd.SpaceDates = []string{"s1:2025-01-01", "extra"}
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on length mismatch in SpaceDates")
	}

	// validateSpaceDates: duplicate space date
	wsd.SpaceIDs = []coretypes.SpaceID{"s1", "s2"}
	wsd.DatesFields.Dates = []string{"2025-01-01"}
	wsd.SpaceDates = []string{"s1:2025-01-01", "s1:2025-01-01"}
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on duplicate space date")
	}

	// validateSpaceDates: invalid format
	wsd.SpaceIDs = []coretypes.SpaceID{"s1"}
	wsd.DatesFields.Dates = []string{"2025-01-01"}
	wsd.SpaceDates = []string{"badformat"}
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on invalid format space date")
	}

	// validateSpaceDates: spaceID not in SpaceIDs
	wsd.SpaceIDs = []coretypes.SpaceID{"s1"}
	wsd.DatesFields.Dates = []string{"2025-01-01"}
	wsd.SpaceDates = []string{"s2:2025-01-01"}
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on spaceID not in SpaceIDs")
	}

	// validateSpaceDates: date in SpaceDate not in Dates
	wsd.SpaceIDs = []coretypes.SpaceID{"s1"}
	wsd.DatesFields.Dates = []string{"2025-01-01"}
	wsd.SpaceDates = []string{"s1:2025-01-02"}
	if err := wsd.Validate(); err == nil {
		t.Fatal("expected error on date in SpaceDate not in Dates")
	}
}
