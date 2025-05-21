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

	fmt.Printf("Fetching price for: %s (%s) in region %s\n",
		resource.ResourceType, resource.Size, resource.Region)

	// Set the region from the resource if available
	region := resource.Region
	if region == "" {
		// Use a default region if none specified
		region = "us-east-1"
	}

	// Determine service code and build appropriate filters based on resource type pattern
	var serviceCode string
	var filters []types.Filter

	// Determine service based on resource type pattern
	switch {
	case strings.HasPrefix(resource.ResourceType, "aws_instance"):
		serviceCode = "AmazonEC2"
		filters = buildEC2Filters(resource.Size, region)
	case strings.HasPrefix(resource.ResourceType, "aws_db_instance"):
		serviceCode = "AmazonRDS"
		filters = buildRDSFilters(resource.Size, region)
	case strings.HasPrefix(resource.ResourceType, "aws_elasticache"):
		serviceCode = "AmazonElastiCache"
		filters = buildElastiCacheFilters(resource.Size, region)
	default:
		// For unknown resource types
		resource.HourlyPrice = 0
		resource.MonthlyPrice = 0
		resource.YearlyPrice = 0
		return fmt.Errorf("unsupported resource type for pricing: %s", resource.ResourceType)
	}

	fmt.Printf("Using service code: %s with %d filters\n", serviceCode, len(filters))

	// Call the AWS pricing API with a retry mechanism
	var response *awspricing.GetProductsOutput
	var err error

	for attempt := 1; attempt <= 3; attempt++ {
		response, err = c.pricingClient.GetProducts(context.TODO(), &awspricing.GetProductsInput{
			Filters:     filters,
			MaxResults:  aws.Int32(100),
			ServiceCode: aws.String(serviceCode),
		})

		if err == nil {
			break
		}

		fmt.Printf("Attempt %d failed: %v\n", attempt, err)
		if attempt < 3 {
			// Wait before retrying
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

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

	fmt.Printf("Got %d pricing results\n", len(response.PriceList))

	// Process the pricing data
	if len(response.PriceList) > 0 {
		// Initialize pricing details
		resource.PricingDetails = &model.PricingDetails{
			Currency:      "USD",
			LastUpdated:   time.Now(),
			PricingSource: "AWS Pricing API",
		}

		// Try a few results until we find one with pricing
		var priceFound bool

		for i, priceListItem := range response.PriceList {
			if i >= 5 { // Limit to first 5 results to avoid excessive processing
				break
			}

			// Parse the price data
			var priceData map[string]interface{}
			if err := json.Unmarshal([]byte(priceListItem), &priceData); err != nil {
				fmt.Printf("Failed to parse pricing data: %v\n", err)
				continue
			}

			// Extract product details for logging
			product, hasProduct := priceData["product"].(map[string]interface{})
			if hasProduct {
				attributes, hasAttrs := product["attributes"].(map[string]interface{})
				if hasAttrs {
					if instanceType, ok := attributes["instanceType"].(string); ok {
						fmt.Printf("Result %d is for instance type: %s\n", i, instanceType)
					}
				}
			}

			// Extract on-demand pricing
			terms, ok := priceData["terms"].(map[string]interface{})
			if !ok {
				fmt.Printf("Result %d: Invalid pricing terms structure\n", i)
				continue
			}

			onDemand, ok := terms["OnDemand"].(map[string]interface{})
			if !ok {
				fmt.Printf("Result %d: No on-demand pricing available\n", i)
				continue
			}

			// Try to find a valid price
			for _, priceData := range onDemand {
				priceDimensions, ok := priceData.(map[string]interface{})["priceDimensions"].(map[string]interface{})
				if !ok {
					continue
				}

				for _, dimension := range priceDimensions {
					dimensionData, ok := dimension.(map[string]interface{})
					if !ok {
						continue
					}

					pricePerUnit, ok := dimensionData["pricePerUnit"].(map[string]interface{})
					if !ok {
						continue
					}

					usdPrice, ok := pricePerUnit["USD"].(string)
					if !ok {
						continue
					}

					// Parse the price as float
					var hourlyPrice float64
					if _, err := fmt.Sscanf(usdPrice, "%f", &hourlyPrice); err != nil {
						fmt.Printf("Failed to parse price '%s': %v\n", usdPrice, err)
						continue
					}

					// Found a price!
					fmt.Printf("Found price: $%f/hour\n", hourlyPrice)
					resource.HourlyPrice = hourlyPrice
					resource.MonthlyPrice = hourlyPrice * 730 // Average hours per month
					resource.YearlyPrice = hourlyPrice * 8760 // Hours per year
					priceFound = true

					// Special handling for t3.micro with $0.00 price
					if resource.Size == "t3.micro" && resource.HourlyPrice == 0 {
						fmt.Printf("Special case: t3.micro shows $0.00 price, applying relative pricing calculation\n")

						// Find t3.small price for comparison
						t3SmallFilters := []types.Filter{
							{
								Field: aws.String("ServiceCode"),
								Type:  types.FilterTypeTermMatch,
								Value: aws.String("AmazonEC2"),
							},
							{
								Field: aws.String("instanceType"),
								Type:  types.FilterTypeTermMatch,
								Value: aws.String("t3.small"),
							},
							{
								Field: aws.String("operatingSystem"),
								Type:  types.FilterTypeTermMatch,
								Value: aws.String("Linux"),
							},
						}

						// Call the AWS pricing API
						t3SmallResponse, err := c.pricingClient.GetProducts(context.TODO(), &awspricing.GetProductsInput{
							Filters:     t3SmallFilters,
							MaxResults:  aws.Int32(10),
							ServiceCode: aws.String("AmazonEC2"),
						})

						if err == nil && len(t3SmallResponse.PriceList) > 0 {
							// Find t3.small price
							var t3SmallPrice float64
							var smallPriceFound bool

							// Process each result
							for _, priceListItem := range t3SmallResponse.PriceList {
								var priceData map[string]interface{}
								if err := json.Unmarshal([]byte(priceListItem), &priceData); err != nil {
									continue
								}

								// Extract terms
								terms, ok := priceData["terms"].(map[string]interface{})
								if !ok {
									continue
								}

								onDemand, ok := terms["OnDemand"].(map[string]interface{})
								if !ok {
									continue
								}

								// Process on-demand pricing
								for _, priceData := range onDemand {
									priceObj, ok := priceData.(map[string]interface{})
									if !ok {
										continue
									}

									priceDimensions, ok := priceObj["priceDimensions"].(map[string]interface{})
									if !ok {
										continue
									}

									// Process each price dimension
									for _, dimension := range priceDimensions {
										dimensionData, ok := dimension.(map[string]interface{})
										if !ok {
											continue
										}

										pricePerUnit, ok := dimensionData["pricePerUnit"].(map[string]interface{})
										if !ok {
											continue
										}

										usdPrice, ok := pricePerUnit["USD"].(string)
										if !ok {
											continue
										}

										// Parse price
										if _, err := fmt.Sscanf(usdPrice, "%f", &t3SmallPrice); err == nil && t3SmallPrice > 0 {
											fmt.Printf("Found t3.small price: $%f/hour\n", t3SmallPrice)
											smallPriceFound = true
											break
										}
									}

									if smallPriceFound {
										break
									}
								}

								if smallPriceFound {
									break
								}
							}

							if smallPriceFound && t3SmallPrice > 0 {
								// t3.micro is approximately half the price of t3.small
								// This is based on the fact that t3.micro has half the vCPUs and memory
								estimatedPrice := t3SmallPrice * 0.5
								fmt.Printf("Calculating t3.micro price as 50%% of t3.small: $%f/hour\n", estimatedPrice)

								resource.HourlyPrice = estimatedPrice
								resource.MonthlyPrice = estimatedPrice * 730
								resource.YearlyPrice = estimatedPrice * 8760

								resource.PricingDetails.PricingSource = "Estimated based on t3.small pricing (50% ratio)"
							}
						}
					}

					break
				}
				if priceFound {
					break
				}
			}
			if priceFound {
				break
			}
		}

		if !priceFound {
			fmt.Printf("Could not find valid pricing in any of the results\n")
			resource.HourlyPrice = 0
			resource.MonthlyPrice = 0
			resource.YearlyPrice = 0
			return fmt.Errorf("could not extract price from API response")
		}
	} else {
		// Try with fewer filters
		if len(filters) > 2 {
			fmt.Printf("No results with specific filters, trying with fewer filters\n")

			// Simplified filters - just service code and instance type
			simplifiedFilters := []types.Filter{
				{
					Field: aws.String("ServiceCode"),
					Type:  types.FilterTypeTermMatch,
					Value: aws.String(serviceCode),
				},
			}

			if resource.Size != "" {
				// The field name might be different depending on the service
				var fieldName string
				switch serviceCode {
				case "AmazonEC2":
					fieldName = "instanceType"
				case "AmazonRDS":
					fieldName = "instanceType"
				case "AmazonElastiCache":
					fieldName = "cacheNodeType"
				default:
					fieldName = "instanceType"
				}

				simplifiedFilters = append(simplifiedFilters, types.Filter{
					Field: aws.String(fieldName),
					Type:  types.FilterTypeTermMatch,
					Value: aws.String(resource.Size),
				})
			}

			fmt.Printf("Trying simplified filters: %+v\n", simplifiedFilters)

			simplifiedResponse, err := c.pricingClient.GetProducts(context.TODO(), &awspricing.GetProductsInput{
				Filters:     simplifiedFilters,
				MaxResults:  aws.Int32(100),
				ServiceCode: aws.String(serviceCode),
			})

			if err == nil && len(simplifiedResponse.PriceList) > 0 {
				fmt.Printf("Got %d results with simplified filters\n", len(simplifiedResponse.PriceList))

				// Process the simplified response
				var priceFound bool
				for i, priceListItem := range simplifiedResponse.PriceList {
					if i >= 5 { // Limit to first 5 results
						break
					}

					var priceData map[string]interface{}
					if err := json.Unmarshal([]byte(priceListItem), &priceData); err != nil {
						continue
					}

					// Extract product details for logging
					product, hasProduct := priceData["product"].(map[string]interface{})
					if hasProduct {
						attributes, hasAttrs := product["attributes"].(map[string]interface{})
						if hasAttrs {
							if instanceType, ok := attributes["instanceType"].(string); ok {
								fmt.Printf("Simplified result %d is for instance type: %s\n", i, instanceType)
							}
						}
					}

					// Extract terms
					if terms, ok := priceData["terms"].(map[string]interface{}); ok {
						if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
							for _, priceData := range onDemand {
								if priceDimensions, ok := priceData.(map[string]interface{})["priceDimensions"].(map[string]interface{}); ok {
									for _, dimension := range priceDimensions {
										if dimensionData, ok := dimension.(map[string]interface{}); ok {
											if pricePerUnit, ok := dimensionData["pricePerUnit"].(map[string]interface{}); ok {
												if usdPrice, ok := pricePerUnit["USD"].(string); ok {
													var hourlyPrice float64
													if _, err := fmt.Sscanf(usdPrice, "%f", &hourlyPrice); err == nil {
														fmt.Printf("Found price with simplified filters: $%f/hour\n", hourlyPrice)

														// Only use non-zero prices
														if hourlyPrice > 0 {
															// Initialize pricing details if needed
															if resource.PricingDetails == nil {
																resource.PricingDetails = &model.PricingDetails{
																	Currency:      "USD",
																	LastUpdated:   time.Now(),
																	PricingSource: "AWS Pricing API (simplified query)",
																}
															}

															resource.HourlyPrice = hourlyPrice
															resource.MonthlyPrice = hourlyPrice * 730
															resource.YearlyPrice = hourlyPrice * 8760
															priceFound = true
															break
														}
													}
												}
											}
										}
									}
									if priceFound {
										break
									}
								}
							}
						}
					}

					if priceFound {
						break
					}
				}

				if !priceFound {
					fmt.Printf("Could not find valid pricing in simplified results\n")
					resource.HourlyPrice = 0
					resource.MonthlyPrice = 0
					resource.YearlyPrice = 0
					return fmt.Errorf("no valid pricing found for resource: %s", resource.ID)
				}
			} else {
				fmt.Printf("Still no results with simplified filters\n")
				resource.HourlyPrice = 0
				resource.MonthlyPrice = 0
				resource.YearlyPrice = 0
				return fmt.Errorf("no pricing data found for resource: %s", resource.ID)
			}
		} else {
			// If no pricing data found and we've already tried simplified filters
			resource.HourlyPrice = 0
			resource.MonthlyPrice = 0
			resource.YearlyPrice = 0
			return fmt.Errorf("no pricing data found for resource: %s", resource.ID)
		}
	}

	return nil
}

// Helper functions to build filters for different services
func buildEC2Filters(instanceType, region string) []types.Filter {
	filters := []types.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AmazonEC2"),
		},
	}

	// Add region filter
	if region != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("regionCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(region),
		})
	}

	// Add instance type filter if available
	if instanceType != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("instanceType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(instanceType),
		})
	}

	// Add additional common filters for EC2
	filters = append(filters, types.Filter{
		Field: aws.String("operatingSystem"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String("Linux"),
	})

	filters = append(filters, types.Filter{
		Field: aws.String("tenancy"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String("Shared"),
	})

	return filters
}

func buildRDSFilters(instanceType, region string) []types.Filter {
	filters := []types.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AmazonRDS"),
		},
	}

	// Add region filter
	if region != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("regionCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(region),
		})
	}

	// Add instance type filter if available
	if instanceType != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("instanceType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(instanceType),
		})
	}

	// Add database engine filter (default to MySQL)
	filters = append(filters, types.Filter{
		Field: aws.String("databaseEngine"),
		Type:  types.FilterTypeTermMatch,
		Value: aws.String("MySQL"),
	})

	return filters
}

func buildElastiCacheFilters(instanceType, region string) []types.Filter {
	filters := []types.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String("AmazonElastiCache"),
		},
	}

	// Add region filter
	if region != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("regionCode"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(region),
		})
	}

	// Add instance type filter if available
	if instanceType != "" {
		filters = append(filters, types.Filter{
			Field: aws.String("cacheNodeType"),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(instanceType),
		})
	}

	return filters
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
