package researchdata

import (
	"github.com/research-data-analysis/route"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("ResearchDataAnalysis", route.URL)
}