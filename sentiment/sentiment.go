package sentiment

type SentimentAnalyzer interface {
	Run(texts string) float64
}
