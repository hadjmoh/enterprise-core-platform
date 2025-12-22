package cloud

import "time"

// Provider represents a cloud service provider
type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderAzure Provider = "azure"
	ProviderGCP   Provider = "gcp"
)

// CloudAsset represents a unified cloud resource across providers
type CloudAsset struct {
	ID           string                 `json:"id"`            // Unified ID: aws:ec2:i-123456
	Provider     Provider               `json:"provider"`
	Type         string                 `json:"type"`          // ec2, vm, compute-instance
	Name         string                 `json:"name"`
	Region       string                 `json:"region"`
	Account      string                 `json:"account"`       // Account/Subscription/Project
	Tags         map[string]string      `json:"tags"`
	State        string                 `json:"state"`         // running, stopped, terminated
	CreatedAt    time.Time              `json:"created_at"`
	LastSeen     time.Time              `json:"last_seen"`
	
	// Network
	PrivateIP      string   `json:"private_ip,omitempty"`
	PublicIP       string   `json:"public_ip,omitempty"`
	VPC            string   `json:"vpc,omitempty"`
	SecurityGroups []string `json:"security_groups,omitempty"`
	
	// IAM
	IAMRole     string   `json:"iam_role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	
	// Risk (from Phase 6)
	RiskScore   int      `json:"risk_score"`
	RiskFactors []string `json:"risk_factors,omitempty"`
	
	// Raw metadata
	RawMetadata map[string]interface{} `json:"raw_metadata"`
}

// CloudEvent represents a unified cloud event across providers
type CloudEvent struct {
	ID        string   `json:"id"`
	Provider  Provider `json:"provider"`
	EventType string   `json:"event_type"` // iam:CreateUser, compute.instances.insert
	EventName string   `json:"event_name"`
	Timestamp time.Time `json:"timestamp"`
	
	// Actor (linked to Phase 6 identity)
	ActorID   string `json:"actor_id"`
	ActorType string `json:"actor_type"` // user, service, role
	ActorName string `json:"actor_name"`
	
	// Resource
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	
	// Context
	SourceIP  string `json:"source_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Region    string `json:"region"`
	Account   string `json:"account"`
	
	// Result
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	
	// Raw event
	RawEvent map[string]interface{} `json:"raw_event"`
}

// AssetRelationship represents a connection between cloud assets
type AssetRelationship struct {
	FromID   string                 `json:"from_id"`
	ToID     string                 `json:"to_id"`
	Type     string                 `json:"type"` // attached_to, member_of, routes_to
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AssetFilter for querying assets
type AssetFilter struct {
	Provider Provider
	Type     string
	Region   string
	Account  string
	Tags     map[string]string
	MinRisk  int
	MaxRisk  int
}
