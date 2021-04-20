package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
)

var (
	bloomingdalesLogger = log.New(os.Stdout, "bloomingdales: ", log.LUTC)
)

const (
	bloomingdalesBase                 = "https://www.bloomingdales.com"
	bloomingdalesProductBase          = "www.bloomingdales.com/shop/product/"
	bloomingdalesProductSelector      = "li div a"
	bloomingdalesNextCategorySelector = ".nextArrow a"
	bloomingdalesRootCategorySelector = ".adCatIcon a"

	outputDirBase = "./gap_pictures/"
	cacheDir      = "_gap_cache"
	jpgSuffix     = ".jpeg"
)

// Bloomingdales scrapes Bloomingdales.
type bloomingdales struct {
	rootPages []string
}

func Bloomingdales(rootPages []string) Scraper {
	return &bloomingdales{
		rootPages: rootPages,
	}
}

// Process processes Bloomingdales' pictures.
func (b *bloomingdales) Process(ctx context.Context) error {
	if b.rootPages == nil || len(b.rootPages) == 0 {
		bloomingdalesLogger.Println(ErrEmptyRootPages)
		return ErrEmptyRootPages
	}

	productURLChan := make(chan string, 1024)
	wg := sync.WaitGroup{}

	// For each root page, collect all of the product urls.
	for _, page := range b.rootPages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			page := page
			wg.Add(1)
			go func() {
				defer wg.Done()
				collectProductURLs(ctx, page, productURLChan)
			}()
		}
	}

	// Close the product channel after all product URL collection is done.
	go func() {
		wg.Wait()
		close(productURLChan)
	}()

	wg2 := sync.WaitGroup{}
	// Scrape product url for images.
	for i := 0; i < 4; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			scrapeProductURL(ctx, productURLChan)
		}()
	}

	wg2.Wait()

	return nil
}

// collectProductURLs takes a root page and returns all of the products reachable from the root.
func collectProductURLs(ctx context.Context, page string, productURLChan chan string) {
	// For each root page, scrape the product pages as well as the ext page if available.
	rootCollector := colly.NewCollector()
	rootCollector.UserAgent = UserAgent

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	rootCollector.Limit(&colly.LimitRule{
		DomainRegexp: "*",
		Parallelism:  2,
		Delay:        5 * time.Second,
	})

	// Get all the categories.
	rootCollector.OnHTML(bloomingdalesRootCategorySelector, func(e *colly.HTMLElement) {
		categoryURL := fmt.Sprintf("%s%s", bloomingdalesBase, e.Attr("href"))
		// bloomingdalesLogger.Println("CategoryURL: " + categoryURL)
		scrapeCategoryPage(ctx, rootCollector, categoryURL, productURLChan)
	})

	// Print All Errors.
	rootCollector.OnError(func(r *colly.Response, err error) {
		errMsg := fmt.Sprintf(
			"Error: %s\nURL: %s",
			err,
			r.Request.URL.RequestURI())
		bloomingdalesLogger.Println(errMsg)
	})

	rootCollector.Visit(page)
}

// scrapeCategoryPage scrapes a given URL. Gets all of the product URLs on a page and then finds the next category page if present.
func scrapeCategoryPage(ctx context.Context, c *colly.Collector, categoryURL string, productURLChan chan string) {
	d := c.Clone()

	// Get all product links on the current page.
	d.OnHTML(bloomingdalesProductSelector, func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
			// Only accept the URL from this element if it is a product URL
			if productURL := fmt.Sprintf("%s%s", bloomingdalesBase, e.Attr("href")); strings.Contains(productURL, bloomingdalesProductBase) {
				productURLChan <- productURL
			}
		}
	})

	// Get the next category page if available
	d.OnHTML(bloomingdalesNextCategorySelector, func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
			categoryURL := fmt.Sprintf("%s%s", bloomingdalesBase, e.Attr("href"))
			scrapeCategoryPage(ctx, c, categoryURL, productURLChan)
		}
	})

	d.Visit(categoryURL)
}

// scrapeProductURL scrapes the images off of the given product URL.
func scrapeProductURL(ctx context.Context, productURLChan chan string) {
	for productURL := range productURLChan {
		select {
		case <-ctx.Done():
			return
		default:
			c := colly.NewCollector()
			c.UserAgent = UserAgent

			c.OnHTML("picture img", func(e *colly.HTMLElement) {
				d := c.Clone()

				d.OnResponse(func(r *colly.Response) {
					filename := strings.Split(r.FileName(), ".")[0]

					// We don't want swatches
					if !strings.Contains(filename, "swatches") {
						filenameLen := len(filename)
						if len(filename) > 50 {
							filenameLen = 50
						}
						if writeErr := writeImage(filename[:filenameLen], r.Body); writeErr != nil {
							bloomingdalesLogger.Fatalln(writeErr)
						}
					}
				})

				d.Visit(e.Attr(("src")))
			})

			// Print All Errors.
			c.OnError(func(r *colly.Response, err error) {
				errMsg := fmt.Sprintf(
					"scrapeProductURL Error: %s\nURL: %s",
					err,
					r.Request.URL.RequestURI())
				bloomingdalesLogger.Println(errMsg)
			})

			c.Visit(productURL)

			time.Sleep(1 * time.Second)
		}
	}
}

func writeImage(name string, data []byte) error {
	imgPath := cacheDir + "/" + name + jpgSuffix
	bloomingdalesLogger.Println(imgPath + " being created")

	os.Mkdir(cacheDir, 0775)
	f, statErr := os.Create(imgPath)
	if statErr != nil {
		return errors.Wrap(statErr, fmt.Sprintf("could not create %s", imgPath))
	}
	defer f.Close()
	f.Write(data)

	return nil
}
