package retrieval

import (
	"testing"
)

func TestImprovementRanking(t *testing.T) {
	items := []Improvement{
		{Title: "Low impact hard", Impact: ImpactLow, Effort: EffortHigh, Confidence: 0.9},
		{Title: "High impact easy", Impact: ImpactHigh, Effort: EffortLow, Confidence: 0.9},
		{Title: "Medium impact medium", Impact: ImpactMedium, Effort: EffortMedium, Confidence: 0.8},
	}

	ranked := RankImprovements(items)

	if len(ranked) != 3 {
		t.Fatalf("expected 3 items, got %d", len(ranked))
	}

	if ranked[0].Title != "High impact easy" {
		t.Errorf("expected first item 'High impact easy', got %q", ranked[0].Title)
	}
	if ranked[2].Title != "Low impact hard" {
		t.Errorf("expected last item 'Low impact hard', got %q", ranked[2].Title)
	}
}

func TestImprovementRankingMaxFive(t *testing.T) {
	items := make([]Improvement, 7)
	for i := 0; i < 7; i++ {
		items[i] = Improvement{
			Title:      "Item",
			Impact:     ImpactMedium,
			Effort:     EffortMedium,
			Confidence: 0.5,
		}
	}

	ranked := RankImprovements(items)
	if len(ranked) > 5 {
		t.Errorf("expected max 5 items, got %d", len(ranked))
	}
}

func TestImprovementImpactString(t *testing.T) {
	tests := []struct {
		level ImpactLevel
		want  string
	}{
		{ImpactLow, "Low"},
		{ImpactMedium, "Medium"},
		{ImpactHigh, "High"},
		{ImpactLevel(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.level.String()
			if got != tc.want {
				t.Errorf("(%d).String() = %q, want %q", tc.level, got, tc.want)
			}
		})
	}
}

func TestImprovementEffortString(t *testing.T) {
	tests := []struct {
		level EffortLevel
		want  string
	}{
		{EffortLow, "Low"},
		{EffortMedium, "Medium"},
		{EffortHigh, "High"},
		{EffortLevel(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.level.String()
			if got != tc.want {
				t.Errorf("(%d).String() = %q, want %q", tc.level, got, tc.want)
			}
		})
	}
}

func TestScoreImprovement(t *testing.T) {
	highImpactLowEffort := Improvement{Impact: ImpactHigh, Effort: EffortLow, Confidence: 1.0}
	lowImpactHighEffort := Improvement{Impact: ImpactLow, Effort: EffortHigh, Confidence: 1.0}

	scoreHigh := scoreImprovement(highImpactLowEffort)
	scoreLow := scoreImprovement(lowImpactHighEffort)

	if scoreHigh <= scoreLow {
		t.Errorf("expected high-impact/low-effort (%f) to score higher than low-impact/high-effort (%f)", scoreHigh, scoreLow)
	}
}

func TestScoreImprovementConfidenceAdjustment(t *testing.T) {
	highConf := Improvement{Impact: ImpactHigh, Effort: EffortLow, Confidence: 1.0}
	lowConf := Improvement{Impact: ImpactHigh, Effort: EffortLow, Confidence: 0.3}

	highScore := scoreImprovement(highConf)
	lowScore := scoreImprovement(lowConf)

	if highScore <= lowScore {
		t.Errorf("expected higher confidence to score higher: %f vs %f", highScore, lowScore)
	}
}

func TestImprovementZeroItems(t *testing.T) {
	ranked := RankImprovements(nil)
	if len(ranked) != 0 {
		t.Errorf("expected 0 items for nil input, got %d", len(ranked))
	}
}

func TestImprovementSingleItem(t *testing.T) {
	items := []Improvement{{Title: "Only one", Impact: ImpactHigh, Effort: EffortLow, Confidence: 0.5}}
	ranked := RankImprovements(items)
	if len(ranked) != 1 {
		t.Fatalf("expected 1 item, got %d", len(ranked))
	}
	if ranked[0].Title != "Only one" {
		t.Errorf("expected title 'Only one', got %q", ranked[0].Title)
	}
}
