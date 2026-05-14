package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/timescale/ghost/internal/api"
)

// PricingOutput is the shared output shape for pricing data, used by both the
// CLI command and the MCP tool. Includes the wire-format hourly figures and
// derived per-month figures.
type PricingOutput struct {
	Dedicated DedicatedPricingOutput `json:"dedicated"`
}

// DedicatedPricingOutput holds pricing for dedicated databases.
type DedicatedPricingOutput struct {
	Compute []ComputePriceOutput `json:"compute"`
	Storage StoragePriceOutput   `json:"storage"`
}

// ComputePriceOutput holds the price and resource allocation for one dedicated
// compute size.
type ComputePriceOutput struct {
	Size          string  `json:"size"`
	MilliCPU      int     `json:"milli_cpu"`
	MemoryGiB     int     `json:"memory_gib"`
	PricePerHour  float64 `json:"price_per_hour"`
	PricePerMonth float64 `json:"price_per_month"`
}

// StoragePriceOutput holds the dedicated storage rate and free quota.
type StoragePriceOutput struct {
	PricePerGiBHour        float64 `json:"price_per_gib_hour"`
	PricePerGiBMonth       float64 `json:"price_per_gib_month"`
	IncludedGiBPerDatabase int     `json:"included_gib_per_database"`
}

// FetchPricing fetches dedicated pricing from the API.
func FetchPricing(ctx context.Context, client api.ClientWithResponsesInterface) (PricingOutput, error) {
	resp, err := client.GetPricingWithResponse(ctx)
	if err != nil {
		return PricingOutput{}, fmt.Errorf("failed to get pricing: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return PricingOutput{}, ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}
	if resp.JSON200 == nil {
		return PricingOutput{}, errors.New("empty response from API")
	}

	p := *resp.JSON200
	compute := make([]ComputePriceOutput, len(p.Dedicated.Compute))
	for i, c := range p.Dedicated.Compute {
		compute[i] = ComputePriceOutput{
			Size:          string(c.Size),
			MilliCPU:      c.MilliCpu,
			MemoryGiB:     c.MemoryGib,
			PricePerHour:  c.PricePerHour,
			PricePerMonth: c.PricePerMonth,
		}
	}
	return PricingOutput{
		Dedicated: DedicatedPricingOutput{
			Compute: compute,
			Storage: StoragePriceOutput{
				PricePerGiBHour:        p.Dedicated.Storage.PricePerGibHour,
				PricePerGiBMonth:       p.Dedicated.Storage.PricePerGibMonth,
				IncludedGiBPerDatabase: p.Dedicated.Storage.IncludedGibPerDatabase,
			},
		},
	}, nil
}
