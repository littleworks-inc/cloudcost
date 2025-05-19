package model

import (
	"time"
)

// Report represents a complete cost estimation report
type Report struct {
	// Basic information
	Timestamp     time.Time `json:"timestamp"`
	Currency      string    `json:"currency"`
	IaCFormat     string    `json:"iac_format"`
	IaCVersion    string    `json:"iac_version,omitempty"`
	ReportID      string    `json:"report_id"`
	ReportName    string    `json:"report_name,omitempty"`
	ReportVersion string    `json:"report_version"`

	// Resources and costs
	Resources    []Resource `json:"resources"`
	TotalHourly  float64    `json:"total_hourly"`
	TotalMonthly float64    `json:"total_monthly"`
	TotalYearly  float64    `json:"total_yearly"`

	// Breakdowns
	ByProvider     map[string]float64            `json:"by_provider,omitempty"`
	ByResourceType map[string]float64            `json:"by_resource_type,omitempty"`
	ByRegion       map[string]float64            `json:"by_region,omitempty"`
	ByTag          map[string]map[string]float64 `json:"by_tag,omitempty"`

	// Diff information (for comparison reports)
	IsDiff           bool           `json:"is_diff,omitempty"`
	PreviousReportID string         `json:"previous_report_id,omitempty"`
	AddedResources   []Resource     `json:"added_resources,omitempty"`
	RemovedResources []Resource     `json:"removed_resources,omitempty"`
	ChangedResources []ResourceDiff `json:"changed_resources,omitempty"`
	PriceDiff        float64        `json:"price_diff,omitempty"`
	PriceDiffPercent float64        `json:"price_diff_percent,omitempty"`

	// Errors and warnings
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`

	// Metadata
	MetaData map[string]string `json:"metadata,omitempty"`

	// Optimization suggestions
	Suggestions []Suggestion `json:"suggestions,omitempty"`
}

// ResourceDiff represents the difference between two versions of a resource
type ResourceDiff struct {
	ResourceID  string    `json:"resource_id"`
	OldResource *Resource `json:"old_resource"`
	NewResource *Resource `json:"new_resource"`
	PriceDiff   float64   `json:"price_diff"`
	Changes     []Change  `json:"changes"`
}

// Change represents a change in a resource property
type Change struct {
	Property     string  `json:"property"`
	OldValue     string  `json:"old_value"`
	NewValue     string  `json:"new_value"`
	ImpactOnCost float64 `json:"impact_on_cost,omitempty"`
}

// Suggestion represents a cost optimization suggestion
type Suggestion struct {
	ResourceID       string  `json:"resource_id,omitempty"`
	ResourceType     string  `json:"resource_type,omitempty"`
	Type             string  `json:"type"` // e.g., "rightsizing", "reserved_instance", "spot_instance"
	Description      string  `json:"description"`
	CurrentCost      float64 `json:"current_cost"`
	SuggestedCost    float64 `json:"suggested_cost"`
	PotentialSavings float64 `json:"potential_savings"`
	SavingsPercent   float64 `json:"savings_percent"`
	Confidence       float64 `json:"confidence,omitempty"`    // 0.0 to 1.0
	ApplyCommand     string  `json:"apply_command,omitempty"` // CLI command to apply the suggestion
}

// NewReport creates a new empty report
func NewReport() *Report {
	return &Report{
		Timestamp:      time.Now(),
		Currency:       "USD",
		Resources:      make([]Resource, 0),
		ByProvider:     make(map[string]float64),
		ByResourceType: make(map[string]float64),
		ByRegion:       make(map[string]float64),
		ByTag:          make(map[string]map[string]float64),
		Errors:         make([]string, 0),
		Warnings:       make([]string, 0),
		MetaData:       make(map[string]string),
		ReportVersion:  "1.0",
	}
}

// AddResource adds a resource to the report and updates totals
func (r *Report) AddResource(resource Resource) {
	r.Resources = append(r.Resources, resource)

	// Update totals
	r.TotalHourly += resource.HourlyPrice * float64(resource.Quantity)
	r.TotalMonthly += resource.MonthlyPrice * float64(resource.Quantity)
	r.TotalYearly += resource.YearlyPrice * float64(resource.Quantity)

	// Update breakdowns
	r.ByProvider[resource.Provider] += resource.MonthlyPrice * float64(resource.Quantity)
	r.ByResourceType[resource.ResourceType] += resource.MonthlyPrice * float64(resource.Quantity)
	r.ByRegion[resource.Region] += resource.MonthlyPrice * float64(resource.Quantity)

	// Update tag breakdowns
	for key, value := range resource.Tags {
		if _, ok := r.ByTag[key]; !ok {
			r.ByTag[key] = make(map[string]float64)
		}
		r.ByTag[key][value] += resource.MonthlyPrice * float64(resource.Quantity)
	}
}

// AddWarning adds a warning message to the report
func (r *Report) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

// AddError adds an error message to the report
func (r *Report) AddError(err string) {
	r.Errors = append(r.Errors, err)
}

// AddSuggestion adds a cost optimization suggestion to the report
func (r *Report) AddSuggestion(suggestion Suggestion) {
	r.Suggestions = append(r.Suggestions, suggestion)
}

// Summarize calculates summary information for the report
func (r *Report) Summarize() {
	// Reset totals
	r.TotalHourly = 0
	r.TotalMonthly = 0
	r.TotalYearly = 0

	// Clear breakdowns
	r.ByProvider = make(map[string]float64)
	r.ByResourceType = make(map[string]float64)
	r.ByRegion = make(map[string]float64)
	r.ByTag = make(map[string]map[string]float64)

	// Recalculate everything
	for _, resource := range r.Resources {
		// Update totals
		r.TotalHourly += resource.HourlyPrice * float64(resource.Quantity)
		r.TotalMonthly += resource.MonthlyPrice * float64(resource.Quantity)
		r.TotalYearly += resource.YearlyPrice * float64(resource.Quantity)

		// Update breakdowns
		r.ByProvider[resource.Provider] += resource.MonthlyPrice * float64(resource.Quantity)
		r.ByResourceType[resource.ResourceType] += resource.MonthlyPrice * float64(resource.Quantity)
		r.ByRegion[resource.Region] += resource.MonthlyPrice * float64(resource.Quantity)

		// Update tag breakdowns
		for key, value := range resource.Tags {
			if _, ok := r.ByTag[key]; !ok {
				r.ByTag[key] = make(map[string]float64)
			}
			r.ByTag[key][value] += resource.MonthlyPrice * float64(resource.Quantity)
		}
	}
}
