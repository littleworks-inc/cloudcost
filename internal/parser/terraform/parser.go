package terraform

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/littleworks-inc/cloudcost/internal/parser"
	"github.com/littleworks-inc/cloudcost/pkg/model"
	"github.com/zclconf/go-cty/cty"
)

// Parser implements the parser.Parser interface for Terraform files
type Parser struct {
	// Configuration options can be added here
	analyzer *parser.ResourceAnalyzer
}

// NewParser creates a new Terraform parser
func NewParser() parser.Parser {
	return &Parser{
		analyzer: &parser.ResourceAnalyzer{},
	}
}

// Parse parses Terraform files and extracts resources
func (p *Parser) Parse(path string) ([]model.Resource, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path error: %v", err)
	}

	var tfFiles []string

	// If path is a directory, find all .tf files
	if info.IsDir() {
		files, err := filepath.Glob(filepath.Join(path, "*.tf"))
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %v", err)
		}
		tfFiles = files
	} else {
		// Single file
		if strings.HasSuffix(path, ".tf") {
			tfFiles = []string{path}
		} else {
			return nil, fmt.Errorf("not a Terraform file: %s", path)
		}
	}

	if len(tfFiles) == 0 {
		return nil, fmt.Errorf("no Terraform files found in: %s", path)
	}

	// Parse all Terraform files
	resources := []model.Resource{}

	hclParser := hclparse.NewParser()

	for _, file := range tfFiles {
		src, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file, err)
		}

		f, diags := hclParser.ParseHCL(src, file)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to parse file %s: %v", file, diags)
		}

		content, _, diags := f.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
			},
		})

		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to extract blocks from file %s: %v", file, diags)
		}

		for _, block := range content.Blocks {
			if block.Type == "resource" {
				resourceType := block.Labels[0]
				resourceName := block.Labels[1]

				// Create resource
				resource := model.NewResource()
				resource.ID = fmt.Sprintf("%s.%s", resourceType, resourceName)
				resource.Name = resourceName
				resource.ResourceType = resourceType

				// Extract provider from resource type
				parts := strings.Split(resourceType, "_")
				if len(parts) > 0 {
					resource.Provider = parts[0]
				}

				// Extract attributes
				attrs, _ := block.Body.JustAttributes()

				// Use our analyzer to find size field
				resource.Size = p.analyzer.FindSizeField(resourceType, attrs)

				// Use our analyzer to find region field
				region := p.analyzer.FindRegionField(resourceType, attrs)
				if region != "" {
					resource.Region = region
				}

				// Extract quantity and tags
				resource.Quantity = p.analyzer.FindQuantity(attrs)
				resource.Tags = p.analyzer.ExtractTags(attrs)

				// If some properties weren't determined, fall back to original method
				if resource.Size == "" {
					// Try to find size attribute using explicit checks
					for name, attr := range attrs {
						value, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() {
							if name == "instance_type" || name == "size" {
								if value.Type() == cty.String {
									resource.Size = value.AsString()
								}
							}
						}
					}
				}

				// If region wasn't determined, check explicitly
				if resource.Region == "" {
					// Try to find region attribute
					if attr, ok := attrs["region"]; ok {
						value, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() && value.Type() == cty.String {
							resource.Region = value.AsString()
						}
					}
				}

				// Ensure default quantity
				if resource.Quantity < 1 {
					resource.Quantity = 1
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// CanHandle checks if this parser can handle the given path
func (p *Parser) CanHandle(path string) bool {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// If it's a directory, check for .tf files
	if info.IsDir() {
		files, err := filepath.Glob(filepath.Join(path, "*.tf"))
		return err == nil && len(files) > 0
	}

	// If it's a file, check if it's a .tf file
	return strings.HasSuffix(path, ".tf")
}

// GetName returns the name of the parser
func (p *Parser) GetName() string {
	return "Terraform"
}
