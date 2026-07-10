package domain

import "time"

type Role string

const (
	RoleOwner    Role = "owner"
	RoleAdmin    Role = "admin"
	RoleEditor   Role = "editor"
	RoleReviewer Role = "reviewer"
	RoleViewer   Role = "viewer"
)

type User struct {
	ID                  string     `json:"id"`
	IdentityProviderUID string     `json:"-"`
	Email               string     `json:"email"`
	EmailVerified       bool       `json:"emailVerified"`
	DisplayName         string     `json:"displayName"`
	AvatarURL           *string    `json:"avatarUrl,omitempty"`
	Locale              string     `json:"locale"`
	Timezone            string     `json:"timezone"`
	PersonalWorkspaceID *string    `json:"personalWorkspaceId,omitempty"`
	DeletionScheduledAt *time.Time `json:"deletionScheduledAt,omitempty"`
	Version             int64      `json:"version"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type Workspace struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Kind        string     `json:"kind"`
	OwnerUserID string     `json:"ownerUserId"`
	Version     int64      `json:"version"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
	PurgeAt     *time.Time `json:"purgeAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CurrentMembership *Membership `json:"currentMembership,omitempty"`
}

type Membership struct {
	ID                 string     `json:"id"`
	WorkspaceID        string     `json:"workspaceId"`
	UserID             string     `json:"userId"`
	Email              string     `json:"email,omitempty"`
	DisplayName        string     `json:"displayName,omitempty"`
	Role               Role       `json:"role"`
	Status             string     `json:"status"`
	CanSpend           bool       `json:"canSpend"`
	MonthlyLimitMicros *int64     `json:"monthlyLimitMicros,omitempty"`
	PerRunLimitMicros  *int64     `json:"perRunLimitMicros,omitempty"`
	LibraryPublish     bool       `json:"libraryPublish"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	RemovedAt          *time.Time `json:"removedAt,omitempty"`
}

type Invitation struct {
	ID string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Email string `json:"email"`
	Role Role `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type BudgetPolicy struct {
	WorkspaceID             string    `json:"workspaceId"`
	Currency                string    `json:"currency"`
	MonthlyLimitMicros      *int64    `json:"monthlyLimitMicros,omitempty"`
	PerRunApprovalMicros    *int64    `json:"perRunApprovalMicros,omitempty"`
	EditorCanPublishLibrary bool      `json:"editorCanPublishLibrary"`
	Version                 int64     `json:"version"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type ProjectSettings struct {
	Name                 string `json:"name"`
	ProjectType          string `json:"projectType"`
	ProductionFormat     string `json:"productionFormat"`
	AspectWidth          int    `json:"aspectWidth"`
	AspectHeight         int    `json:"aspectHeight"`
	FrameRateNumerator   int    `json:"frameRateNumerator"`
	FrameRateDenominator int    `json:"frameRateDenominator"`
	AudioLanguage        string `json:"audioLanguage"`
	Rating               string `json:"rating"`
	StyleDirection       string `json:"styleDirection"`
	QualityPolicy        string `json:"qualityPolicy"`
	CostCeilingMicros    *int64 `json:"costCeilingMicros,omitempty"`
}

type Project struct {
	ID             string          `json:"id"`
	WorkspaceID    string          `json:"workspaceId"`
	Slug           string          `json:"slug"`
	Status         string          `json:"status"`
	Privacy        string          `json:"privacy"`
	CurrentVersion int64           `json:"currentVersion"`
	Settings       ProjectSettings `json:"settings"`
	CreatedBy      string          `json:"createdBy"`
	DeletedAt      *time.Time      `json:"deletedAt,omitempty"`
	PurgeAt        *time.Time      `json:"purgeAt,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	SourceTemplateID *string       `json:"sourceTemplateId,omitempty"`
	SourceTemplateVersion *int64   `json:"sourceTemplateVersion,omitempty"`
}

type ProjectCreateInput struct {
	Name string `json:"name"`
	TemplateID *string `json:"templateId,omitempty"`
	SettingsOverrides *ProjectSettings `json:"settingsOverrides,omitempty"`
}

type ProjectVersion struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"projectId"`
	WorkspaceID string          `json:"workspaceId"`
	Version     int64           `json:"version"`
	Settings    ProjectSettings `json:"settings"`
	CreatedBy   string          `json:"createdBy"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type ProjectTemplate struct {
	ID             string          `json:"id"`
	WorkspaceID    *string         `json:"workspaceId,omitempty"`
	Key            *string         `json:"key,omitempty"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	System         bool            `json:"system"`
	CurrentVersion int64           `json:"currentVersion"`
	Settings       ProjectSettings `json:"settings"`
	LibraryLinks   []string        `json:"libraryLinks"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type AuditEvent struct {
	ID            string         `json:"id"`
	WorkspaceID   *string        `json:"workspaceId,omitempty"`
	ActorUserID   *string        `json:"actorUserId,omitempty"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      *string        `json:"targetId,omitempty"`
	Decision      string         `json:"decision"`
	SourceContext map[string]any `json:"sourceContext"`
	CorrelationID string         `json:"correlationId"`
	OccurredAt    time.Time      `json:"occurredAt"`
}

type Library struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspaceId"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Version     int64      `json:"version"`
	ArchivedAt  *time.Time `json:"archivedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type LibraryEntityVersion struct {
	ID              string         `json:"id"`
	EntityID        string         `json:"entityId"`
	Version         int64          `json:"version"`
	Type            string         `json:"type"`
	Name            string         `json:"name"`
	Aliases         []string       `json:"aliases"`
	Summary         string         `json:"summary"`
	Description     string         `json:"description"`
	Tags            []string       `json:"tags"`
	Attributes      map[string]any `json:"attributes"`
	Status          string         `json:"status"`
	ChangeSummary   string         `json:"changeSummary"`
	AuthorUserID    string         `json:"authorUserId"`
	SourceVersionID *string        `json:"sourceVersionId,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type LibraryEntity struct {
	ID               string               `json:"id"`
	WorkspaceID      string               `json:"workspaceId"`
	LibraryID        string               `json:"libraryId"`
	CurrentVersionID string               `json:"currentVersionId"`
	CurrentVersion   LibraryEntityVersion `json:"currentVersion"`
	ArchivedAt       *time.Time           `json:"archivedAt,omitempty"`
	CreatedAt        time.Time            `json:"createdAt"`
	UpdatedAt        time.Time            `json:"updatedAt"`
}

type MediaAsset struct {
	ID               string         `json:"id"`
	WorkspaceID      string         `json:"workspaceId"`
	Kind             string         `json:"kind"`
	State            string         `json:"state"`
	OriginalFilename string         `json:"originalFilename"`
	ContentType      string         `json:"contentType"`
	SizeBytes        int64          `json:"sizeBytes"`
	ChecksumSHA256   *string        `json:"checksumSha256,omitempty"`
	Metadata         map[string]any `json:"metadata"`
	RejectionCode    *string        `json:"rejectionCode,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}
