package aws

import (
	"context"
	"fmt"
	"strings"

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
}

// NewClient creates a new AWS pricing client
func NewClient() pricing.Client {
	return &Client{
		region: "us-east-1", // Default region for queries
	}
}

// GetPrice retrieves the price for a specific resource
func (c *Client) GetPrice(resource *model.Resource) error {
	if c.pricingClient == nil {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	// Set the region from the resource if available
	region := resource.Region
	if region == "" {
		region = c.region
	}

	// Determine the service and instance type based on the resource type
	var service, instanceType string

	switch {
	case strings.HasPrefix(resource.ResourceType, "aws_instance"):
		service = "AmazonEC2"
		instanceType = resource.Size
	case strings.HasPrefix(resource.ResourceType, "aws_rds"):
		service = "AmazonRDS"
		instanceType = resource.Size
	case strings.HasPrefix(resource.ResourceType, "aws_elasticache"):
		service = "AmazonElastiCache"
		instanceType = resource.Size
	default:
		// For unknown resource types, we can't determine pricing
		return fmt.Errorf("unable to determine pricing for resource type: %s", resource.ResourceType)
	}

	// Build filters for the pricing API
	filters := []types.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(service),
		},
		{
			Field: aws.String("regionCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(region),
		},
	}

	// Add instance type filter if available
	if instanceType != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("instanceType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(instanceType),
		})
	}

	// Call the AWS pricing API
	response, err := c.pricingClient.GetProducts(context.TODO(), &awspricing.GetProductsInput{
		Filters:     filters,
		MaxResults:  aws.Int32(100),
		ServiceCode: aws.String(service),
	})

	if err != nil {
		return fmt.Errorf("failed to get pricing data: %v", err)
	}

	// Process the pricing data
	if len(response.PriceList) > 0 {
		// Parse the price list (this is a simplified example)
		// In a real implementation, we would parse the JSON response and extract the pricing details

		// For now, let's use a placeholder hourly price
		resource.HourlyPrice = 0.01 // Placeholder, will be replaced with actual pricing data

		// Calculate monthly and yearly prices
		resource.MonthlyPrice = resource.HourlyPrice * 730 // Average hours per month
		resource.YearlyPrice = resource.HourlyPrice * 8760 // Hours per year
	} else {
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
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // Pricing API is only available in us-east-1
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %v", err)
	}

	// Create pricing client
	c.pricingClient = awspricing.NewFromConfig(cfg)

	return nil
}
