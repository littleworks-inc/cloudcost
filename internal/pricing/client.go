package pricing

import (
	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// Client defines the interface for cloud pricing clients
type Client interface {
	// GetPrice returns the price for a specific resource
	GetPrice(resource *model.Resource) error

	// GetName returns the name of the pricing client (e.g., "AWS", "Azure")
	GetName() string

	// Initialize sets up the pricing client (e.g., authentication)
	Initialize() error
}
