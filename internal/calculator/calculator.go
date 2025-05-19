package calculator

import (
	"github.com/littleworks-inc/cloudcost/internal/pricing"
	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// Calculator calculates costs for cloud resources
type Calculator struct {
	PricingClients map[string]pricing.Client
}

// NewCalculator creates a new calculator with pricing clients
func NewCalculator() *Calculator {
	return &Calculator{
		PricingClients: make(map[string]pricing.Client),
	}
}

// RegisterPricingClient registers a pricing client for a cloud provider
func (c *Calculator) RegisterPricingClient(provider string, client pricing.Client) {
	c.PricingClients[provider] = client
}

// CalculateCosts calculates costs for all resources
func (c *Calculator) CalculateCosts(resources []model.Resource) (*model.Report, error) {
	// Initialize report
	report := &model.Report{
		Resources:      resources,
		ByProvider:     make(map[string]float64),
		ByResourceType: make(map[string]float64),
		ByRegion:       make(map[string]float64),
	}

	// Calculate costs for each resource
	for i := range resources {
		resource := &resources[i]

		// Get client for this provider
		client, ok := c.PricingClients[resource.Provider]
		if !ok {
			// Skip resources with no pricing client
			continue
		}

		// Get pricing data
		if err := client.GetPrice(resource); err != nil {
			// Log error and continue
			continue
		}

		// Add to totals
		report.TotalHourly += resource.HourlyPrice
		report.TotalMonthly += resource.MonthlyPrice
		report.TotalYearly += resource.YearlyPrice

		// Add to breakdowns
		report.ByProvider[resource.Provider] += resource.MonthlyPrice
		report.ByResourceType[resource.ResourceType] += resource.MonthlyPrice
		report.ByRegion[resource.Region] += resource.MonthlyPrice
	}

	return report, nil
}
