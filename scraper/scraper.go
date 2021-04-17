package scraper

import (
	"context"
	"errors"
)

var (
	ErrEmptyRootPages = errors.New("empty root page slice")
)

// Scraper scrapes a retailers website.
type Scraper interface {
	Process(ctx context.Context) error
}
