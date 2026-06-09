package management

type TrustBindingType string

const (
	TrustBindingTypeTerraformCloud TrustBindingType = "terraform_cloud"
	TrustBindingTypeGitHubActions  TrustBindingType = "github_actions"
)

type TrustBindingTypeDefinition struct {
	Key                    TrustBindingType `json:"key"`
	DisplayName            string           `json:"display_name"`
	Description            string           `json:"description"`
	LogoURL                string           `json:"logo_url"`
	DocsURL                string           `json:"docs_url"`
	RequiredIdentityClaims []string         `json:"required_identity_claims"`
}

type TrustBindingTypeList struct {
	Data []TrustBindingTypeDefinition `json:"data"`
}

type TrustBinding struct {
	ID           string           `json:"id"`
	WorkspaceID  string           `json:"workspace_id"`
	BindingType  TrustBindingType `json:"binding_type"`
	Identity     map[string]any   `json:"identity"`
	OutputClaims map[string]any   `json:"output_claims,omitempty"`
	CreatedAt    Time             `json:"created_at"`
	UpdatedAt    Time             `json:"updated_at"`
}

type TrustBindingInput struct {
	BindingType  TrustBindingType `json:"binding_type"`
	Identity     map[string]any   `json:"identity"`
	OutputClaims map[string]any   `json:"output_claims,omitempty"`
}

type TrustBindingList struct {
	Data []TrustBinding `json:"data"`
}
