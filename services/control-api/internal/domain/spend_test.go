package domain

import "testing"

func TestAuthorizeSpend(t *testing.T) {
	workspaceLimit, memberLimit, runLimit, approval := int64(10_000), int64(5_000), int64(2_000), int64(1_000)
	policy := BudgetPolicy{MonthlyLimitMicros: &workspaceLimit, PerRunApprovalMicros: &approval}
	member := Membership{CanSpend: true, MonthlyLimitMicros: &memberLimit, PerRunLimitMicros: &runLimit}

	if got := AuthorizeSpend(policy, member, SpendRequest{AmountMicros: 1_500}); !got.Allowed || !got.ApprovalRequired {
		t.Fatalf("expected allowed request requiring approval, got %#v", got)
	}
	if got := AuthorizeSpend(policy, member, SpendRequest{AmountMicros: 2_001}); got.Allowed || got.Reason != "member_per_run_limit" {
		t.Fatalf("expected per-run rejection, got %#v", got)
	}
	member.CanSpend = false
	if got := AuthorizeSpend(policy, member, SpendRequest{AmountMicros: 1}); got.Allowed || got.Reason != "member_spending_disabled" {
		t.Fatalf("expected disabled rejection, got %#v", got)
	}
}
