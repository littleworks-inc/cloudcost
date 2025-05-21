package parser

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// ResourceAnalyzer helps extract information from resources dynamically
type ResourceAnalyzer struct{}

// FindSizeField looks for the best field representing size
func (a *ResourceAnalyzer) FindSizeField(resourceType string, attrs hcl.Attributes) string {
	// Common size field patterns by priority
	sizePatterns := []struct {
		suffix   string
		priority int
	}{
		{"instance_type", 100},
		{"_type", 90},
		{"machine_type", 85},
		{"size", 80},
		{"instance_class", 75},
		{"_class", 70},
		{"_tier", 65},
		{"_size", 60},
		{"node_type", 55},
		{"bundle_id", 50},
		{"sku_name", 45},
		{"flavor", 40},
	}

	// Collect all potential size fields with priorities
	candidates := make(map[string]int)

	// First pass: exact matches based on known patterns
	for _, pattern := range sizePatterns {
		for attrName, attr := range attrs {
			// Check exact match
			if attrName == pattern.suffix || strings.HasSuffix(attrName, pattern.suffix) {
				// Extract value directly from the attribute expression
				if val, err := a.getExprStringValue(attr.Expr); err == nil && val != "" {
					candidates[val] = pattern.priority
				}
			}
		}
	}

	// Second pass: check for values that look like instance types
	for attrName, attr := range attrs {
		// Extract value directly from the attribute expression
		if val, err := a.getExprStringValue(attr.Expr); err == nil && val != "" {
			// Check if it looks like an instance type
			if a.looksLikeInstanceType(val) {
				// Boost priority if the attribute name contains size-related terms
				priority := 30
				if strings.Contains(attrName, "instance") ||
					strings.Contains(attrName, "type") ||
					strings.Contains(attrName, "size") ||
					strings.Contains(attrName, "class") {
					priority = 75
				}
				candidates[val] = priority
			}
		}
	}

	// Find the highest priority candidate
	var bestValue string
	bestPriority := -1

	for val, priority := range candidates {
		if priority > bestPriority {
			bestValue = val
			bestPriority = priority
		}
	}

	return bestValue
}

// looksLikeInstanceType determines if a string resembles a cloud instance type
func (a *ResourceAnalyzer) looksLikeInstanceType(value string) bool {
	// Common patterns for instance types across cloud providers:

	// AWS EC2: t2.micro, m5.large, c5n.xlarge
	if strings.Contains(value, ".") && len(value) >= 4 {
		parts := strings.Split(value, ".")
		if len(parts) >= 2 {
			// Check if first part contains letter+number pattern
			first := parts[0]
			if len(first) >= 2 {
				hasLetter := false
				hasNumber := false
				for _, c := range first {
					if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
						hasLetter = true
					} else if c >= '0' && c <= '9' {
						hasNumber = true
					}
				}
				if hasLetter && hasNumber {
					return true
				}
			}
		}
	}

	// Azure VMs: Standard_D2s_v3, Standard_B1s
	if strings.HasPrefix(value, "Standard_") && strings.Contains(value, "_") {
		return true
	}

	// GCP: n1-standard-1, e2-medium
	if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) >= 2 {
			return true
		}
	}

	// Check for common size suffixes
	sizeSuffixes := []string{"small", "medium", "large", "xlarge", "2xlarge", "micro", "nano"}
	for _, suffix := range sizeSuffixes {
		if strings.HasSuffix(value, suffix) || strings.HasSuffix(value, "."+suffix) {
			return true
		}
	}

	return false
}

// FindRegionField looks for the region information
func (a *ResourceAnalyzer) FindRegionField(resourceType string, attrs hcl.Attributes) string {
	// Common region field names
	regionFields := []string{"region", "location", "zone", "availability_zone"}

	// Try direct matches first
	for _, fieldName := range regionFields {
		if attr, ok := attrs[fieldName]; ok {
			// Extract value directly from the attribute expression
			if val, err := a.getExprStringValue(attr.Expr); err == nil && val != "" {
				return val
			}
		}
	}

	// Look for fields containing region-related keywords
	for attrName, attr := range attrs {
		lowerName := strings.ToLower(attrName)
		if strings.Contains(lowerName, "region") ||
			strings.Contains(lowerName, "location") ||
			strings.Contains(lowerName, "zone") {
			// Extract value directly from the attribute expression
			if val, err := a.getExprStringValue(attr.Expr); err == nil && val != "" {
				return val
			}
		}
	}

	// Look for values that look like regions
	for _, attr := range attrs {
		// Extract value directly from the attribute expression
		if val, err := a.getExprStringValue(attr.Expr); err == nil && val != "" {
			if a.looksLikeRegion(val) {
				return val
			}
		}
	}

	return ""
}

// looksLikeRegion checks if a string looks like a cloud region
func (a *ResourceAnalyzer) looksLikeRegion(value string) bool {
	// AWS regions: us-east-1, eu-west-2
	if strings.Count(value, "-") == 2 && len(value) >= 7 {
		parts := strings.Split(value, "-")
		if len(parts) == 3 && len(parts[0]) == 2 && len(parts[1]) >= 4 {
			_, err := fmt.Sscanf(parts[2], "%d", new(int))
			return err == nil
		}
	}

	// Azure regions: eastus, westeurope
	commonAzureRegions := []string{
		"eastus", "westus", "centralus", "northeurope", "westeurope",
		"eastasia", "southeastasia", "australiaeast",
	}

	for _, region := range commonAzureRegions {
		if strings.EqualFold(value, region) {
			return true
		}
	}

	// GCP regions: us-central1, europe-west1
	if strings.Count(value, "-") == 1 && strings.HasSuffix(value, "1") {
		parts := strings.Split(value, "-")
		if len(parts) == 2 && (len(parts[0]) == 2 || len(parts[0]) > 4) {
			return true
		}
	}

	return false
}

// FindQuantity determines the quantity/count of resources
func (a *ResourceAnalyzer) FindQuantity(attrs hcl.Attributes) int {
	// Look for count attribute
	if attr, ok := attrs["count"]; ok {
		// Extract number directly from the attribute expression
		value, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() && value.Type() == cty.Number {
			f, _ := value.AsBigFloat().Float64()
			if f > 0 {
				return int(f)
			}
		}
	}

	// Default to 1 if no count found
	return 1
}

// ExtractTags extracts resource tags
func (a *ResourceAnalyzer) ExtractTags(attrs hcl.Attributes) map[string]string {
	tags := make(map[string]string)

	// Look for tags attribute
	if attr, ok := attrs["tags"]; ok {
		value, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() && value.Type().IsMapType() {
			value.ForEachElement(func(key cty.Value, val cty.Value) bool {
				if key.Type() == cty.String && val.Type() == cty.String {
					tags[key.AsString()] = val.AsString()
				}
				return true
			})
		}
	}

	return tags
}

// getExprStringValue extracts a string value from an HCL expression
func (a *ResourceAnalyzer) getExprStringValue(expr hcl.Expression) (string, error) {
	value, diags := expr.Value(nil)
	if diags.HasErrors() || value.Type() != cty.String {
		return "", fmt.Errorf("not a string value")
	}
	return value.AsString(), nil
}
