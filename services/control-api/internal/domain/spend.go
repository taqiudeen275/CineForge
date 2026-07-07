package domain

type SpendRequest struct {
	AmountMicros              int64
	WorkspaceMonthSpendMicros int64
	MemberMonthSpendMicros    int64
}

type SpendDecision struct {
	Allowed          bool   `json:"allowed"`
	ApprovalRequired bool   `json:"approvalRequired"`
	Reason           string `json:"reason,omitempty"`
}

// AuthorizeSpend is the reusable policy decision used by future generation and
// rendering commands. Callers must pass committed month-to-date ledger totals.
func AuthorizeSpend(policy BudgetPolicy, member Membership, request SpendRequest) SpendDecision {
	if request.AmountMicros < 0 {
		return SpendDecision{Reason: "invalid_amount"}
	}
	if !member.CanSpend {
		return SpendDecision{Reason: "member_spending_disabled"}
	}
	if member.PerRunLimitMicros != nil && request.AmountMicros > *member.PerRunLimitMicros {
		return SpendDecision{Reason: "member_per_run_limit"}
	}
	if member.MonthlyLimitMicros != nil && request.MemberMonthSpendMicros+request.AmountMicros > *member.MonthlyLimitMicros {
		return SpendDecision{Reason: "member_monthly_limit"}
	}
	if policy.MonthlyLimitMicros != nil && request.WorkspaceMonthSpendMicros+request.AmountMicros > *policy.MonthlyLimitMicros {
		return SpendDecision{Reason: "workspace_monthly_limit"}
	}
	decision := SpendDecision{Allowed: true}
	if policy.PerRunApprovalMicros != nil && request.AmountMicros > *policy.PerRunApprovalMicros {
		decision.ApprovalRequired = true
		decision.Reason = "workspace_approval_threshold"
	}
	return decision
}
