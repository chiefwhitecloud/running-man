package dataimport

type RaceFetcher interface {
	GetRawResults(url string) ([]byte, error)
}
