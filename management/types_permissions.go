package management

// PermissionSet is the read shape of a permission set as returned by the
// Management API. Permissions is the full list of (resource, action) grants
// bound into this set.
type PermissionSet struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   Time         `json:"created_at"`
	UpdatedAt   Time         `json:"updated_at"`
}

// PermissionSetInput is the write shape for Create and Update. Updates replace
// the Permissions list wholesale — the API does not support partial diffs.
type PermissionSetInput struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Permissions []PermissionInput `json:"permissions"`
}

// Permission is one grant inside a PermissionSet.
type Permission struct {
	ID         string `json:"id"`
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	Constraint string `json:"constraint,omitempty"`
}

// PermissionInput is the write shape for a single grant. Constraint is an
// optional CEL expression scoping the grant (e.g. `resource.owner == subject`).
type PermissionInput struct {
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	Constraint string `json:"constraint,omitempty"`
}

// PermissionSetList is the paginated list response.
type PermissionSetList struct {
	Items []PermissionSet `json:"items"`
}
