package controller

import (
	"fmt"
	"time"

	"github.com/littleworks-inc/cloudcost/internal/calculator"
	"github.com/littleworks-inc/cloudcost/internal/parser"
	"github.com/littleworks-inc/cloudcost/internal/pricing"
	"github.com/littleworks-inc/cloudcost/internal/utils"
	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// Estimator is the main controller for cost estimation
type Estimator struct {
	Parsers        []parser.Parser
	PricingClients map[string]pricing.Client
	Calculator     *calculator.Calculator
}

// NewEstimator creates a new estimator
func NewEstimator() *Estimator {
	return &Estimator{
		Parsers:        make([]parser.Parser, 0),
		PricingClients: make(map[string]pricing.Client),
		Calculator:     calculator.NewCalculator(),
	}
}

// RegisterParser registers an IaC parser
func (e *Estimator) RegisterParser(p parser.Parser) {
	e.Parsers = append(e.Parsers, p)
}

// RegisterPricingClient registers a pricing client
func (e *Estimator) RegisterPricingClient(provider string, client pricing.Client) {
	e.PricingClients[provider] = client
	e.Calculator.RegisterPricingClient(provider, client)
}

// Estimate performs cost estimation on IaC files
func (e *Estimator) Estimate(path string) (*model.Report, error) {
	// Detect IaC type
	iacType, err := utils.DetectIaCType(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect IaC type: %v", err)
	}

	if iacType == utils.TypeUnknown {
		return nil, fmt.Errorf("could not determine IaC type for path: %s", path)
	}

	fmt.Printf("Detected IaC type: %s\n", iacType)

	// Find appropriate parser
	var selectedParser parser.Parser
	for _, p := range e.Parsers {
		if p.CanHandle(path) {
			selectedParser = p
			break
		}
	}

	if selectedParser == nil {
		return nil, fmt.Errorf("no parser available for IaC type: %s", iacType)
	}

	fmt.Printf("Using parser: %s\n", selectedParser.GetName())

	// Parse IaC files
	resources, err := selectedParser.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IaC files: %v", err)
	}

	fmt.Printf("Parsed %d resources\n", len(resources))

	// Calculate costs
	report, err := e.Calculator.CalculateCosts(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate costs: %v", err)
	}

	// Set report metadata
	report.IaCFormat = string(iacType)
	report.Timestamp = time.Now()

	return report, nil
}

// Compare compares current IaC costs with a previous report
func (e *Estimator) Compare(path string, previousReportPath string) (*model.Report, error) {
	// Estimate current costs
	currentReport, err := e.Estimate(path)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate current costs: %v", err)
	}

	// TODO: Load previous report and compare
	// For now, just return current report
	return currentReport, nil
}
