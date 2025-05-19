// File: internal/pricing/aws/client.go
package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awspricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/littleworks-inc/cloudcost/internal/pricing"
	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// Client implements the pricing.Client interface for AWS
type Client struct {
	pricingClient *awspricing.Client
	region        string
	initialized   bool
	error         error // Store initialization error
}

// NewClient creates a new AWS pricing client
func NewClient() pricing.Client {
	return &Client{
		region:      "us-east-1", // Default region for queries (AWS Pricing API only available in us-east-1)
		initialized: false,
	}
}

// GetPrice retrieves the price for a specific resource
func (c *Client) GetPrice(resource *model.Resource) error {
	// Initialize client if needed
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			// Set prices to zero but preserve the error for reporting
			resource.HourlyPrice = 0
			resource.MonthlyPrice = 0
			resource.YearlyPrice = 0

			// Set pricing details with error information
			resource.PricingDetails = &model.PricingDetails{
				Currency:      "USD",
				LastUpdated:   time.Now(),
				PricingSource: "Error: " + err.Error(),
			}

			return fmt.Errorf("AWS pricing data unavailable: %v", err)
		}
	} else if c.error != nil {
		// We tried to initialize before and failed
		resource.HourlyPrice = 0
		resource.MonthlyPrice = 0
		resource.YearlyPrice = 0

		// Set pricing details with error information
		resource.PricingDetails = &model.PricingDetails{
			Currency:      "USD",
			LastUpdated:   time.Now(),
			PricingSource: "Error: " + c.error.Error(),
		}

		return fmt.Errorf("AWS pricing data unavailable: %v", c.error)
	}

	// Set the region from the resource if available
	region := resource.Region
	if region == "" {
		// Use a default region if none specified
		region = "us-east-1"
	}

	// Determine the service and instance type based on the resource type
	var service, instanceType string

	switch {
	case strings.HasPrefix(resource.ResourceType, "aws_instance"):
		service = "AmazonEC2"
		instanceType = resource.Size
	case strings.HasPrefix(resource.ResourceType, "aws_db_instance"):
		service = "AmazonRDS"
		instanceType = resource.Size
	case strings.HasPrefix(resource.ResourceType, "aws_elasticache"):
		service = "AmazonElastiCache"
		instanceType = resource.Size
	default:
		// For unknown resource types
		resource.HourlyPrice = 0
		resource.MonthlyPrice = 0
		resource.YearlyPrice = 0

		// Set pricing details with error information
		resource.PricingDetails = &model.PricingDetails{
			Currency:      "USD",
			LastUpdated:   time.Now(),
			PricingSource: "Error: Unsupported resource type",
		}

		return fmt.Errorf("unsupported resource type for pricing: %s", resource.ResourceType)
	}

	// Build filters for the pricing API
	filters := []types.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(service),
		},
	}

	// Add region filter
	filters = append(filters, types.Filter{
		Field: aws.String("regionCode"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String(region),
	})

	// Add instance type filter if available
	if instanceType != "" {
		switch service {
		case "AmazonEC2":
			filters = append(filters, types.Filter{
				Field: aws.String("instanceType"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String(instanceType),
			})
		case "AmazonRDS":
			filters = append(filters, types.Filter{
				Field: aws.String("instanceType"),
				Type:  types.FilterTypeTermMatch,
				Value: aws.String(instanceType),
			})
		}
	}

	// Call the AWS pricing API
	response, err := c.pricingClient.GetProducts(context.TODO(), &awspricing.GetProductsInput{
		Filters:     filters,
		MaxResults:  aws.Int32(100),
		ServiceCode: aws.String(service),
	})

	if err != nil {
		// If API call fails, set prices to zero and return error
		resource.HourlyPrice = 0
		resource.MonthlyPrice = 0
		resource.YearlyPrice = 0

		// Set pricing details with error information
		resource.PricingDetails = &model.PricingDetails{
			Currency:      "USD",
			LastUpdated:   time.Now(),
			PricingSource: "Error: " + err.Error(),
		}

		return fmt.Errorf("failed to get pricing data: %v", err)
	}

	// Process the pricing data
	if len(response.PriceList) > 0 {
		// Initialize pricing details
		resource.PricingDetails = &model.PricingDetails{
			Currency:    "USD",
			LastUpdated: time.Now(),
		}

		// Parse the first price in the list (we'll pick the first match)
		var priceData map[string]interface{}
		if err := json.Unmarshal([]byte(response.PriceList[0]), &priceData); err != nil {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: Failed to parse pricing data"
			return fmt.Errorf("failed to parse pricing data: %v", err)
		}

		// Extract on-demand pricing (this structure depends on AWS API response)
		terms, ok := priceData["terms"].(map[string]interface{})
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: Invalid pricing data structure"
			return fmt.Errorf("invalid pricing data structure")
		}

		onDemand, ok := terms["OnDemand"].(map[string]interface{})
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: No on-demand pricing available"
			return fmt.Errorf("no on-demand pricing available")
		}

		// Get the first on-demand price SKU
		var sku string
		for k := range onDemand {
			sku = k
			break
		}

		if sku == "" {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: No pricing SKU found"
			return fmt.Errorf("no pricing SKU found")
		}

		priceData, ok = onDemand[sku].(map[string]interface{})
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: Invalid SKU pricing structure"
			return fmt.Errorf("invalid SKU pricing structure")
		}

		// Get the first price dimension
		dimensions, ok := priceData["priceDimensions"].(map[string]interface{})
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: No price dimensions found"
			return fmt.Errorf("no price dimensions found")
		}

		var dimensionKey string
		for k := range dimensions {
			dimensionKey = k
			break
		}

		if dimensionKey == "" {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: No price dimension key found"
			return fmt.Errorf("no price dimension key found")
		}

		dimension, ok := dimensions[dimensionKey].(map[string]interface{})
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: Invalid price dimension structure"
			return fmt.Errorf("invalid price dimension structure")
		}

		// Get the price per unit
		pricePerUnit, ok := dimension["pricePerUnit"].(map[string]interface{})
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: No price per unit found"
			return fmt.Errorf("no price per unit found")
		}

		// Get the USD price
		usdPrice, ok := pricePerUnit["USD"].(string)
		if !ok {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: No USD price found"
			return fmt.Errorf("no USD price found")
		}

		// Parse the price as float
		var hourlyPrice float64
		if _, err := fmt.Sscanf(usdPrice, "%f", &hourlyPrice); err != nil {
			// Set pricing details with error information
			resource.PricingDetails.PricingSource = "Error: Failed to parse price"
			return fmt.Errorf("failed to parse price: %v", err)
		}

		// Set the hourly price
		resource.HourlyPrice = hourlyPrice

		// Calculate monthly and yearly prices
		resource.MonthlyPrice = resource.HourlyPrice * 730 // Average hours per month
		resource.YearlyPrice = resource.HourlyPrice * 8760 // Hours per year

		// Set pricing details source
		resource.PricingDetails.PricingSource = "AWS Pricing API"

		// Get unit of measure
		unit, ok := dimension["unit"].(string)
		if ok {
			// Add a price component for the main unit
			resource.PricingDetails.PriceComponents = append(
				resource.PricingDetails.PriceComponents,
				model.PriceComponent{
					Name:      "On-Demand " + unit,
					UnitPrice: hourlyPrice,
					Units:     1,
					Total:     hourlyPrice,
				},
			)
		}

		// Get product details
		product, ok := priceData["product"].(map[string]interface{})
		if ok {
			attributes, ok := product["attributes"].(map[string]interface{})
			if ok {
				// Store product attributes
				if resource.PricingDetails.MetaData == nil {
					resource.PricingDetails.MetaData = make(map[string]string)
				}

				for k, v := range attributes {
					if str, ok := v.(string); ok {
						resource.PricingDetails.MetaData[k] = str
					}
				}
			}
		}
	} else {
		// If no pricing data found, set prices to zero
		resource.HourlyPrice = 0
		resource.MonthlyPrice = 0
		resource.YearlyPrice = 0

		// Set pricing details with error information
		resource.PricingDetails = &model.PricingDetails{
			Currency:      "USD",
			LastUpdated:   time.Now(),
			PricingSource: "Error: No pricing data found",
		}

		return fmt.Errorf("no pricing data found for resource: %s", resource.ID)
	}

	return nil
}

// GetName returns the name of the pricing client
func (c *Client) GetName() string {
	return "AWS"
}

// Initialize sets up the pricing client
func (c *Client) Initialize() error {
	// Mark as initialized to avoid repeated initialization attempts
	c.initialized = true

	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // Pricing API is only available in us-east-1
	)
	if err != nil {
		c.error = fmt.Errorf("AWS credentials not found: %v", err)
		return c.error
	}

	// Create pricing client
	c.pricingClient = awspricing.NewFromConfig(cfg)

	// Test the client with a simple API call
	_, err = c.pricingClient.DescribeServices(context.TODO(), &awspricing.DescribeServicesInput{
		MaxResults: aws.Int32(1),
	})

	if err != nil {
		c.error = fmt.Errorf("AWS API access failed: %v", err)
		return c.error
	}

	return nil
}
