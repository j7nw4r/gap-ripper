package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/johnathan-walker/gap-ripper/scraper"
	"github.com/pkg/errors"
)

const (
	outputDirBase = "./gap_pictures/"
	cacheDir      = "./_gap_cache/"
	jpgSuffix     = ".jpg"
)

func main() {
	// Process retailers
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	for _, retailer := range orchestrateScrapers() {
		retailer.Process(ctx)
	}
}

// orchestrateScrapers organizes scrapers.
func orchestrateScrapers() []scraper.Scraper {
	var scrapers []scraper.Scraper

	// Bloomingdales
	scrapers = append(scrapers, scraper.Bloomingdales([]string{
		`https://www.bloomingdales.com/shop/mens?id=3864&cm_sp=NAVIGATION-_-TOP_NAV-_-MEN-n-n`,
		`https://www.bloomingdales.com/shop/womens-apparel?id=2910&cm_sp=NAVIGATION-_-TOP_NAV-_-WOMEN-n-n`,
	}))

	return scrapers
}

func writeImage(name string, data []byte) error {
	imgPath := cacheDir + name + jpgSuffix

	f, statErr := os.Create(imgPath)
	if statErr != nil {
		return errors.Wrap(statErr, fmt.Sprintf("could not create %s", imgPath))
	}
	defer f.Close()
	f.Write(data)

	return nil
}
