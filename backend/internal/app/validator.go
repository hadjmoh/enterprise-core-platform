package app

import (
	"fmt"
	"strings"
)

// ValidationResult represents the outcome of app validation
type ValidationResult struct {
	Passed   bool     `json:"passed"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// Validator performs automated validation checks
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateManifest checks manifest schema and content
func (v *Validator) ValidateManifest(manifest *AppManifest) *ValidationResult {
	result := &ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check API version
	if manifest.APIVersion != "prospect/v1alpha1" {
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported API version: %s", manifest.APIVersion))
		result.Passed = false
	}

	// Check kind
	if manifest.Kind != "App" {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid kind: %s (expected 'App')", manifest.Kind))
		result.Passed = false
	}

	// Check required metadata fields
	if manifest.Metadata.Name == "" {
		result.Errors = append(result.Errors, "App name is required")
		result.Passed = false
	}

	if manifest.Metadata.DisplayName == "" {
		result.Errors = append(result.Errors, "Display name is required")
		result.Passed = false
	}

	if manifest.Metadata.Version == "" {
		result.Errors = append(result.Errors, "Version is required")
		result.Passed = false
	}

	if manifest.Metadata.Description == "" {
		result.Errors = append(result.Errors, "Description is required")
		result.Passed = false
	}

	if manifest.Metadata.Author == "" {
		result.Errors = append(result.Errors, "Author is required")
		result.Passed = false
	}

	// Check name format (lowercase alphanumeric and hyphens)
	if !isValidAppName(manifest.Metadata.Name) {
		result.Errors = append(result.Errors, "App name must contain only lowercase letters, numbers, and hyphens")
		result.Passed = false
	}

	// Warnings for optional fields
	if manifest.Metadata.License == "" {
		result.Warnings = append(result.Warnings, "License not specified")
	}

	if len(manifest.Metadata.Tags) == 0 {
		result.Warnings = append(result.Warnings, "No tags specified (recommended for discoverability)")
	}

	return result
}

// ValidatePermissions checks if permissions are valid
func (v *Validator) ValidatePermissions(permissions []string) *ValidationResult {
	result := &ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	if len(permissions) == 0 {
		result.Warnings = append(result.Warnings, "No permissions requested")
		return result
	}

	// Check each permission against whitelist
	for _, perm := range permissions {
		cap := Capability(strings.ToLower(perm))
		if !IsValidCapability(cap) {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid permission: %s", perm))
			result.Passed = false
		}

		// Warn about high-risk permissions
		if RiskLevel(cap) == "high" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("High-risk permission requested: %s", perm))
		}
	}

	return result
}

// ValidateBundle checks bundle structure and size
func (v *Validator) ValidateBundle(bundlePath string) *ValidationResult {
	result := &ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check file exists
	// (In production, check file size, scan for malware, etc.)
	
	// Placeholder validation
	result.Warnings = append(result.Warnings, "Bundle validation is basic - consider adding malware scanning")

	return result
}

// ScanForIssues performs basic security checks
func (v *Validator) ScanForIssues(manifest *AppManifest) *ValidationResult {
	result := &ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check for path traversal attempts in dashboard paths
	for _, dashboard := range manifest.Dashboards {
		if strings.Contains(dashboard.Path, "..") {
			result.Errors = append(result.Errors, fmt.Sprintf("Path traversal detected in dashboard: %s", dashboard.Path))
			result.Passed = false
		}
	}

	// Warn about excessive permissions
	if len(manifest.Permissions) > 10 {
		result.Warnings = append(result.Warnings, "Large number of permissions requested (>10)")
	}

	return result
}

// ValidateAll runs all validation checks
func (v *Validator) ValidateAll(manifest *AppManifest, bundlePath string) *ValidationResult {
	combined := &ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Run all validations
	manifestResult := v.ValidateManifest(manifest)
	permResult := v.ValidatePermissions(manifest.Permissions)
	bundleResult := v.ValidateBundle(bundlePath)
	securityResult := v.ScanForIssues(manifest)

	// Combine results
	combined.Errors = append(combined.Errors, manifestResult.Errors...)
	combined.Errors = append(combined.Errors, permResult.Errors...)
	combined.Errors = append(combined.Errors, bundleResult.Errors...)
	combined.Errors = append(combined.Errors, securityResult.Errors...)

	combined.Warnings = append(combined.Warnings, manifestResult.Warnings...)
	combined.Warnings = append(combined.Warnings, permResult.Warnings...)
	combined.Warnings = append(combined.Warnings, bundleResult.Warnings...)
	combined.Warnings = append(combined.Warnings, securityResult.Warnings...)

	combined.Passed = manifestResult.Passed && permResult.Passed && 
		bundleResult.Passed && securityResult.Passed

	return combined
}

// Helper function
func isValidAppName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
