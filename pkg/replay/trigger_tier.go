package replay

import (
	"strings"

	"aoe2kit/pkg/triggergraph"
)

func classifyTriggerGraphTier(graph *triggergraph.Graph, region *triggergraph.Region) string {
	if graph == nil {
		return "unavailable"
	}
	for _, warning := range graph.Warnings {
		if warning == "modern embedded trigger count and display-order validated" {
			return "triggergraph_count_and_order_validated"
		}
		lower := strings.ToLower(warning)
		if strings.Contains(lower, "stopped trigger parse") ||
			strings.Contains(lower, "primary trigger region parse failed") ||
			strings.Contains(lower, "fallback") {
			return "triggergraph_partial"
		}
	}
	if region != nil {
		for _, warning := range region.Warnings {
			if strings.Contains(strings.ToLower(warning), "fallback") {
				return "triggergraph_heuristic_complete"
			}
		}
	}
	return "triggergraph_heuristic_complete"
}
