package app

import (
	"fmt"
	"strings"
)

// Capability represents a permission that an app can request
type Capability string

// Standard capabilities
const (
	CapReadLogs        Capability = "read:logs"
	CapWriteAlerts     Capability = "write:alerts"
	CapReadConfig      Capability = "read:config"
	CapWriteConfig     Capability = "write:config"
	CapWriteDashboards Capability = "write:dashboards"
	CapReadThreats     Capability = "read:threats"
	CapWriteThreats    Capability = "write:threats"
	CapManagePipelines Capability = "manage:pipelines"
	CapReadCloud       Capability = "read:cloud"
	CapWriteAssets     Capability = "write:assets"
	CapReadData        Capability = "read:data"
	CapWriteData       Capability = "write:data"
	CapReadIncidents   Capability = "read:incidents"
	CapWriteIncidents  Capability = "write:incidents"
	CapReadReports     Capability = "read:reports"
	CapWriteReports    Capability = "write:reports"
)

// AllCapabilities is the whitelist of valid capabilities
var AllCapabilities = []Capability{
	CapReadLogs,
	CapWriteAlerts,
	CapReadConfig,
	CapWriteConfig,
	CapWriteDashboards,
	CapReadThreats,
	CapWriteThreats,
	CapManagePipelines,
	CapReadCloud,
	CapWriteAssets,
	CapReadData,
	CapWriteData,
	CapReadIncidents,
	CapWriteIncidents,
	CapReadReports,
	CapWriteReports,
}

// CapabilitySet represents a set of capabilities granted to an app
type CapabilitySet struct {
	capabilities map[Capability]bool
}

// NewCapabilitySet creates a new empty capability set
func NewCapabilitySet() *CapabilitySet {
	return &CapabilitySet{
		capabilities: make(map[Capability]bool),
	}
}

// Grant adds a capability to the set
func (cs *CapabilitySet) Grant(cap Capability) error {
	if !IsValidCapability(cap) {
		return fmt.Errorf("invalid capability: %s", cap)
	}
	cs.capabilities[cap] = true
	return nil
}

// Revoke removes a capability from the set
func (cs *CapabilitySet) Revoke(cap Capability) {
	delete(cs.capabilities, cap)
}

// Has checks if a capability is granted
func (cs *CapabilitySet) Has(cap Capability) bool {
	return cs.capabilities[cap]
}

// List returns all granted capabilities
func (cs *CapabilitySet) List() []Capability {
	caps := make([]Capability, 0, len(cs.capabilities))
	for cap := range cs.capabilities {
		caps = append(caps, cap)
	}
	return caps
}

// IsValidCapability checks if a capability is in the whitelist
func IsValidCapability(cap Capability) bool {
	for _, valid := range AllCapabilities {
		if cap == valid {
			return true
		}
	}
	return false
}

// ValidateCapabilities checks if all requested capabilities are valid
func ValidateCapabilities(requested []string) error {
	for _, req := range requested {
		cap := Capability(strings.ToLower(req))
		if !IsValidCapability(cap) {
			return fmt.Errorf("invalid capability requested: %s", req)
		}
	}
	return nil
}

// ParseCapabilities converts string slice to Capability slice
func ParseCapabilities(perms []string) ([]Capability, error) {
	caps := make([]Capability, 0, len(perms))
	for _, p := range perms {
		cap := Capability(strings.ToLower(p))
		if !IsValidCapability(cap) {
			return nil, fmt.Errorf("invalid capability: %s", p)
		}
		caps = append(caps, cap)
	}
	return caps, nil
}

// CapabilityDescription returns a human-readable description
func CapabilityDescription(cap Capability) string {
	descriptions := map[Capability]string{
		CapReadLogs:        "Read system logs and event data",
		CapWriteAlerts:     "Create and modify security alerts",
		CapReadConfig:      "Read system configuration",
		CapWriteConfig:     "Modify system configuration",
		CapWriteDashboards: "Create and modify dashboards",
		CapReadThreats:     "Access threat intelligence data",
		CapWriteThreats:    "Submit threat intelligence",
		CapManagePipelines: "Create and manage data pipelines",
		CapReadCloud:       "Access cloud asset information",
		CapWriteAssets:     "Modify asset inventory",
		CapReadData:        "Read indexed data",
		CapWriteData:       "Write data to indexes",
		CapReadIncidents:   "View security incidents",
		CapWriteIncidents:  "Create and modify incidents",
		CapReadReports:     "Access reports and analytics",
		CapWriteReports:    "Generate custom reports",
	}
	return descriptions[cap]
}

// RiskLevel returns the risk level of a capability
func RiskLevel(cap Capability) string {
	highRisk := []Capability{
		CapWriteConfig,
		CapWriteData,
		CapManagePipelines,
	}
	
	mediumRisk := []Capability{
		CapWriteAlerts,
		CapWriteThreats,
		CapWriteAssets,
		CapWriteIncidents,
		CapWriteReports,
	}
	
	for _, hr := range highRisk {
		if cap == hr {
			return "high"
		}
	}
	
	for _, mr := range mediumRisk {
		if cap == mr {
			return "medium"
		}
	}
	
	return "low"
}
