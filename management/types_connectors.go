package management

// ConnectorProvider names a federation provider shape. The enum is open at
// the SDK boundary so a new provider on the server doesn't break consumers
// between SDK releases.
type ConnectorProvider string

const (
	// ConnectorProviderTerraformCloud is the Terraform Cloud federation
	// provider. The TerraformCloud sub-object carries the org+workspace
	// pair the SYSTEM TFC Trust Exchange resolves at policy time.
	ConnectorProviderTerraformCloud ConnectorProvider = "terraform_cloud"
	// ConnectorProviderGitHubActions is the GitHub Actions federation
	// provider. The GitHubActions sub-object carries the repo+workflow
	// pair the SYSTEM GHA Trust Exchange resolves at policy time.
	ConnectorProviderGitHubActions ConnectorProvider = "github_actions"
)

// TerraformCloudClaims is the provider-shaped claim block for a Terraform
// Cloud federation connector. The wire format keeps the names symmetric with
// what a TFC OIDC token actually carries, so the binding is self-documenting.
type TerraformCloudClaims struct {
	Organization string `json:"organization"`
	Workspace    string `json:"workspace"`
}

// GitHubActionsClaims is the provider-shaped claim block for a GitHub Actions
// federation connector.
type GitHubActionsClaims struct {
	Repository string `json:"repository"`
	Workflow   string `json:"workflow"`
}

// Connector is the read shape of a workspace federation connector. Exactly
// one of TerraformCloud or GitHubActions is populated, matching Provider.
type Connector struct {
	ID             string                `json:"id"`
	WorkspaceID    string                `json:"workspace_id"`
	Provider       ConnectorProvider     `json:"provider"`
	TerraformCloud *TerraformCloudClaims `json:"terraform_cloud,omitempty"`
	GitHubActions  *GitHubActionsClaims  `json:"github_actions,omitempty"`
	CreatedAt      Time                  `json:"created_at"`
	UpdatedAt      Time                  `json:"updated_at"`
}

// ConnectorInput is the write shape for Create. Exactly one of TerraformCloud
// or GitHubActions must be set and must match Provider; mismatch is rejected
// by the server with 400.
type ConnectorInput struct {
	Provider       ConnectorProvider     `json:"provider"`
	TerraformCloud *TerraformCloudClaims `json:"terraform_cloud,omitempty"`
	GitHubActions  *GitHubActionsClaims  `json:"github_actions,omitempty"`
}

// ConnectorList is the list response envelope. The server returns connectors
// under a Data key to match the rest of the workspace-scoped list endpoints.
type ConnectorList struct {
	Data []Connector `json:"data"`
}
