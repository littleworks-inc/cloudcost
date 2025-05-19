package model

import (
	"time"
)

// Resource represents a generic cloud resource
type Resource struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	ResourceType   string                 `json:"resource_type"` // e.g., "aws_instance", "azure_vm"
	Provider       string                 `json:"provider"`      // "aws", "azure", "gcp"
	Region         string                 `json:"region"`
	Size           string                 `json:"size"`     // e.g., "t3.micro", "Standard_B2s"
	Quantity       int                    `json:"quantity"` // Number of instances
	Tags           map[string]string      `json:"tags,omitempty"`
	Properties     map[string]interface{} `json:"properties,omitempty"` // Additional properties
	HourlyPrice    float64                `json:"hourly_price,omitempty"`
	MonthlyPrice   float64                `json:"monthly_price,omitempty"`
	YearlyPrice    float64                `json:"yearly_price,omitempty"`
	PricingDetails *PricingDetails        `json:"pricing_details,omitempty"`
	ParentID       string                 `json:"parent_id,omitempty"` // For resources that belong to others
	Children       []string               `json:"children,omitempty"`  // Child resource IDs
}

// PricingDetails contains detailed pricing information
type PricingDetails struct {
	Currency        string           `json:"currency"`
	EffectiveDate   time.Time        `json:"effective_date"`
	PricingTiers    []PriceTier      `json:"pricing_tiers,omitempty"`
	ReservedPricing bool             `json:"reserved_pricing"`
	UpfrontFee      float64          `json:"upfront_fee,omitempty"`
	PriceComponents []PriceComponent `json:"price_components,omitempty"`
	PricingSource   string           `json:"pricing_source"` // API URL or source of pricing data
	LastUpdated     time.Time        `json:"last_updated"`
}

// PriceTier represents a pricing tier (for tiered pricing models)
type PriceTier struct {
	StartQuantity float64 `json:"start_quantity"`
	EndQuantity   float64 `json:"end_quantity,omitempty"` // Omit for infinite tiers
	UnitPrice     float64 `json:"unit_price"`
}

// PriceComponent represents one component of a resource's price
type PriceComponent struct {
	Name      string  `json:"name"` // e.g., "Compute", "Storage", "Network"
	UnitPrice float64 `json:"unit_price"`
	Units     float64 `json:"units"`
	Total     float64 `json:"total"`
}

// NewResource creates a new resource with default values
func NewResource() Resource {
	return Resource{
		Quantity:   1,
		Tags:       make(map[string]string),
		Properties: make(map[string]interface{}),
	}
}

// CalculateDerivedPrices calculates monthly and yearly prices from hourly
func (r *Resource) CalculateDerivedPrices() {
	if r.HourlyPrice > 0 {
		// Average hours per month (365 days / 12 months * 24 hours)
		r.MonthlyPrice = r.HourlyPrice * 730
		r.YearlyPrice = r.HourlyPrice * 8760 // 365 days * 24 hours
	}
}
