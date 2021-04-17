package scraper

import (
	"context"
	"errors"
)

var (
	ErrEmptyRootPages = errors.New("empty root page slice")
)

const (
	UserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36`
)

// Scraper scrapes a retailers website.
type Scraper interface {
	Process(ctx context.Context) error
}
