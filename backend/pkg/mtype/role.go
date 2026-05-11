package mtype

const (
	RoleUser         = 0
	RolePartnerAdmin = 1
	RolePartner      = 2
	RoleAdmin        = 3
)

type Role uint8

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RolePartnerAdmin, RolePartner, RoleAdmin:
		return true
	}

	return false
}

func (r Role) User() bool {
	return r == RoleUser
}

func (r Role) Admin() bool {
	return r == RoleAdmin
}

func (r Role) Partner() bool {
	return r == RolePartner
}

func (r Role) PartnerAdmin() bool {
	return r == RolePartnerAdmin
}

// HasProjectAdminRights returns true for roles that can manage duels within a project:
// PartnerAdmin (own project), Partner (own project), Admin (any project).
func (r Role) HasProjectAdminRights() bool {
	return r == RolePartnerAdmin || r == RolePartner || r == RoleAdmin
}
