package retrieval

import (
	"sort"
)

type ImpactLevel int

const (
	ImpactLow ImpactLevel = iota
	ImpactMedium
	ImpactHigh
)

func (i ImpactLevel) String() string {
	switch i {
	case ImpactHigh:
		return "High"
	case ImpactMedium:
		return "Medium"
	case ImpactLow:
		return "Low"
	default:
		return "Unknown"
	}
}

type EffortLevel int

const (
	EffortLow EffortLevel = iota
	EffortMedium
	EffortHigh
)

func (e EffortLevel) String() string {
	switch e {
	case EffortHigh:
		return "High"
	case EffortMedium:
		return "Medium"
	case EffortLow:
		return "Low"
	default:
		return "Unknown"
	}
}

type Improvement struct {
	Title      string      `json:"title"`
	Impact     ImpactLevel `json:"impact"`
	Effort     EffortLevel `json:"effort"`
	Confidence float64     `json:"confidence"`
}

func RankImprovements(items []Improvement) []Improvement {
	sort.SliceStable(items, func(i, j int) bool {
		scoreI := scoreImprovement(items[i])
		scoreJ := scoreImprovement(items[j])
		return scoreI > scoreJ
	})
	if len(items) > 5 {
		items = items[:5]
	}
	return items
}

func scoreImprovement(item Improvement) float64 {
	impactScore := float64(item.Impact)
	effortScore := 2.0 - float64(item.Effort)
	return (impactScore*2.0 + effortScore*1.5) * item.Confidence
}
