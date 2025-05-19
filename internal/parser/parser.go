package parser

import (
	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// Parser defines the interface for all IaC parsers
type Parser interface {
	// Parse parses IaC files and returns extracted resources
	Parse(path string) ([]model.Resource, error)

	// CanHandle checks if this parser can handle the given path/files
	CanHandle(path string) bool

	// GetName returns the name of the parser (e.g., "Terraform", "Ansible")
	GetName() string
}
