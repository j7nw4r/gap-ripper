package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gocolly/colly/v2"
)

var (
	bloomingdalesLogger = log.New(os.Stdout, "bloomingdales: ", log.LUTC)
)

const (
	bloomingdalesBase                 = "https://www.bloomingdales.com"
	bloomingdalesProductSelector      = ""
	bloomingdalesNextCategorySelector = ""
	bloomingdalesRootCategorySelector = ".adCatIcon a"
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

	// Scrape product url for images.
	for productURL := range productURLChan {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			scrapeProductURL(productURL)
		}
	}

	return nil
}

// collectProductURLs takes a root page and returns all of the products reachable from the root.
func collectProductURLs(ctx context.Context, page string, productURLChan chan string) {
	// For each root page, scrape the product pages as well as the ext page if available.
	rootCollector := colly.NewCollector()
	rootCollector.UserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36`

	// Print Response metadata.
	rootCollector.OnResponse(func(r *colly.Response) {
		statusCodeMsg := fmt.Sprintf("URL: %s | Status Code: %d", r.Request.URL, r.StatusCode)
		bloomingdalesLogger.Printf(statusCodeMsg)
	})

	// Get all the categories.
	rootCollector.OnHTML(bloomingdalesRootCategorySelector, func(e *colly.HTMLElement) {
		categoryURL := fmt.Sprintf("%s%s", bloomingdalesBase, e.Attr("href"))
		bloomingdalesLogger.Println("Category: " + categoryURL)
		scrapeCategoryPage(rootCollector, categoryURL, productURLChan)
	})

	// Print All Errors.
	rootCollector.OnError(func(r *colly.Response, err error) {
		errMsg := fmt.Sprintf(
			"Error: %s\nURL: %s\nBody: %d",
			err,
			r.Request.URL.RequestURI(),
			r.StatusCode)
		bloomingdalesLogger.Println(errMsg)
	})

	rootCollector.Visit(page)
}

// scrapeCategoryPage scrapes a given URL. Gets all of the product URLs on a page and then finds the next category page if present.
func scrapeCategoryPage(c *colly.Collector, categoryURL string, productURLChan chan string) {
	d := c.Clone()

	// Get all product links on the current page.
	d.OnHTML(bloomingdalesProductSelector, func(e *colly.HTMLElement) {
		productURL := fmt.Sprintf("a product URL from %s", categoryURL)
		productURLChan <- productURL
	})

	// Get the next category page if available
	d.OnHTML(bloomingdalesNextCategorySelector, func(e *colly.HTMLElement) {
		categoryURL := fmt.Sprintf("%s%s", "", e.Attr("href"))
		bloomingdalesLogger.Println("Category Next Page: " + categoryURL)
		scrapeCategoryPage(c, categoryURL, productURLChan)
	})
}

// scrapeProductURL scrapes the images off of the given product URL.
func scrapeProductURL(productURL string) {
	bloomingdalesLogger.Println("Viewing " + productURL)
}
