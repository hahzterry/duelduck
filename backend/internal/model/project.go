package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProjectStatus represents the lifecycle state of a partner project.
// 1=active, 2=disabled (by partner), 3=blocked (by Duel Duck team).
type ProjectStatus int8

const (
	ProjectStatusActive   ProjectStatus = 1
	ProjectStatusDisabled ProjectStatus = 2
	ProjectStatusBlocked  ProjectStatus = 3
)

type Project struct {
	bun.BaseModel `bun:"table:projects,alias:p" json:"-"`

	ID                    uuid.UUID     `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	PartnerID             uuid.UUID     `bun:",type:uuid,notnull" json:"partner_id"`
	APIKey                string        `bun:"api_key,notnull,unique" json:"api_key"`
	Name                  string        `bun:"name,notnull" json:"name"`
	SiteURL               string        `bun:"site_url,notnull" json:"site_url"`
	WalletAddress         string        `bun:"wallet_address,notnull,default:''" json:"wallet_address"`
	Status                ProjectStatus `bun:"status,notnull,default:1" json:"status"`
	IsBlocked             bool          `bun:"is_blocked,notnull,default:false" json:"-"`
	IsUsersDuelsEnabled   bool          `bun:"is_users_duels_enabled,notnull,default:false" json:"is_users_duels_enabled"`
	IsSelfResolvedEnabled bool          `bun:"is_self_resolved_enabled,notnull,default:false" json:"is_self_resolved_enabled"`
	CreatedAt             time.Time     `bun:",notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt             time.Time     `bun:",notnull,default:current_timestamp" json:"updated_at"`

	StatusHistory []ProjectStatusHistory `bun:"rel:has-many,join:id=project_id" json:"status_history,omitempty"`
}

// IsAccessible returns true when the API key should be accepted for requests.
func (p *Project) IsAccessible() bool {
	return p.Status == ProjectStatusActive
}

type ProjectStatusHistory struct {
	bun.BaseModel `bun:"table:project_status_history,alias:psh" json:"-"`

	ID        int64         `bun:",pk,autoincrement" json:"id"`
	ProjectID uuid.UUID     `bun:",type:uuid,notnull" json:"project_id"`
	Status    ProjectStatus `bun:"status,notnull" json:"status"`
	ChangedAt time.Time     `bun:",notnull,default:current_timestamp" json:"changed_at"`
}

func NewProject(
	id uuid.UUID,
	apiKey string,
	partnerID uuid.UUID,
) *Project {
	now := time.Now().UTC()
	return &Project{
		ID:        id,
		PartnerID: partnerID,
		APIKey:    apiKey,
		Status:    ProjectStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type EditProjectReq struct {
	Name    string `json:"name"`
	SiteURL string `json:"site_url"`
}

type UpdateProjectPermissionsReq struct {
	IsUsersDuelsEnabled   bool `json:"is_users_duels_enabled"`
	IsSelfResolvedEnabled bool `json:"is_self_resolved_enabled"`
}
