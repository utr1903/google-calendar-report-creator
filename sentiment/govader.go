package sentiment

import (
	"github.com/jonreiter/govader"
)

type SentimentAnalyzerGoVader struct {
	analyzer *govader.SentimentIntensityAnalyzer
}

func NewGoVader() *SentimentAnalyzerGoVader {
	return &SentimentAnalyzerGoVader{
		analyzer: govader.NewSentimentIntensityAnalyzer(),
	}
}

func (a *SentimentAnalyzerGoVader) Run(
	text string,
) float64 {
	return a.analyzer.PolarityScores(text).Compound
}
