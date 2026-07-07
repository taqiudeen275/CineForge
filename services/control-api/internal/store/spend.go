package store

import (
	"context"

	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *Store) AuthorizeSpend(ctx context.Context, userID, workspaceID string, request domain.SpendRequest) (domain.SpendDecision, error) {
	var policy domain.BudgetPolicy
	var member domain.Membership
	err := s.withUserTx(ctx, userID, func(tx pgx.Tx) error {
		var err error
		member, err = requireRole(ctx, tx, userID, workspaceID, domain.RoleOwner, domain.RoleAdmin, domain.RoleEditor, domain.RoleReviewer, domain.RoleViewer)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select workspace_id::text,currency,monthly_limit_micros,per_run_approval_micros,editor_can_publish_library,version,updated_at from workspace_policies where workspace_id=$1`, workspaceID).Scan(&policy.WorkspaceID, &policy.Currency, &policy.MonthlyLimitMicros, &policy.PerRunApprovalMicros, &policy.EditorCanPublishLibrary, &policy.Version, &policy.UpdatedAt)
	})
	if err != nil {
		return domain.SpendDecision{}, err
	}
	return domain.AuthorizeSpend(policy, member, request), nil
}
