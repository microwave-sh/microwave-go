package management

// PermissionSet is the read shape of a permission set as returned by the
// Management API. Permissions is the full list of scope grants bound into this
// set.
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

// Permission is one scope grant inside a PermissionSet. Name is the scope
// string the API enforces (e.g. "deploys:write", or "*" for full access).
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Dangerous   bool   `json:"dangerous"`
}

// PermissionInput is the write shape for a single scope grant. Name is the
// scope string; Label is a human-readable title; Dangerous flags grants that
// warrant extra confirmation in UIs.
type PermissionInput struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Dangerous   bool   `json:"dangerous"`
}

// PermissionSetList is the paginated list response.
type PermissionSetList struct {
	Items []PermissionSet `json:"items"`
}
