// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"reflect"
	"testing"
)

// This registry cohort may activate only after separately verified cleanup.
// Restoring a retired name must fail the monitor's Unexpected-root check;
// held operational names remain observation debt, never write permissions.
func TestPostCleanupRetirementDetectsEveryRecreatedRoot(t *testing.T) {
	r, _, err := DefaultRootRegistry()
	if err != nil {
		t.Fatal(err)
	}
	held := []string{"currentSpaceServiceRefreshDuePartitionsV1", "datatugGithubProjectLocators", "datatugPersonalLogOperations", "datatugProjectQuickSearch", "datatugWatchControl", "ovdbWritableSandboxCapacity"}
	retired := []string{"aiCloudAllowance", "aiUsage", "botPlatforms", "bttt-matches", "checkoutAccountAcceptedRevisions", "checkoutAccountFinancialEvents", "checkoutAccountInvoiceIndex", "checkoutAccountOwners", "checkoutAccountPaymentIndex", "checkoutAccountSubscriptionIndex", "checkoutAccountSubscriptions", "checkoutBillingAccountIndex", "checkoutCustomerIndex", "checkoutCustomers", "checkoutEntitlements", "checkoutOfferBuyers", "checkoutOfferLedgers", "checkoutOfferPlaces", "checkoutQuoteHistory", "checkoutSessionIndex", "checkoutUnmappedFinancialEvents", "checkoutUnmappedPayments", "commitiusCommitments", "datatugGithubProjectCreateOperations", "datatugGithubQueryOperations", "datatugPaidMoneyMonths", "datatugPersonalLogGrants", "datatugPersonalLogSubjects", "datatugPersonalLogUsers", "datatugPlanApplications", "datatug_github_app_credentials", "datatug_github_oauth_app_credentials", "datatug_projects", "moneyEvents", "oauthDeviceAuthorizations", "oauthDeviceUserCodes", "ovdbAccessTokens", "ovdbDeviceAuthorizations", "ovdbDeviceUserCodes", "slugs", "transfers"}
	if !reflect.DeepEqual(r.LegacyRoots, held) || len(r.LegacyWritableRoots) != 0 {
		t.Fatalf("wrong held cohort: %#v", r)
	}
	got := r.CompareRoots(append(append(append([]string{}, r.ApprovedRoots...), held...), retired...))
	if !reflect.DeepEqual(got.Legacy, held) || !reflect.DeepEqual(got.Unexpected, retired) {
		t.Fatalf("retired roots must be unexpected: %#v", got)
	}
}
