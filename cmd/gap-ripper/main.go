package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
)

const (
	gapBaseURL            = "https://www.gap.com"
	gapProductPageBaseURL = gapBaseURL + "/browse/product.do?pid="
	gapImageSelector      = `div a img[src*="webcontent"]`
	outputDirBase         = "./gap_pictures/"
	cacheDir              = "./_gap_cache/"
	jpgSuffix             = ".jpg"
)

func main() {
	fmt.Println("gap-ripper")
	// Should have at least one product id.
	if len(os.Args[1:]) <= 0 {
		fmt.Println("Must supply Product Id.")
		os.Exit(1)
	}

	fmt.Println("ripping...")
	defer fmt.Print("Done\n\n")

	// Process each product id.
	for _, pid := range os.Args[1:] {
		processProduct(pid)
	}
}

func processProduct(pid string) {
	c := colly.NewCollector(
		colly.CacheDir(cacheDir),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	count := 1
	c.OnHTML(gapImageSelector, func(e *colly.HTMLElement) {
		fmt.Printf("%s %s\n", e.Attr("alt"), e.Attr("src"))

		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			if writeErr := writeImage(pid+strconv.Itoa(count), r.Body); writeErr != nil {
				fmt.Println("Encountered error: " + writeErr.Error())
				os.Exit(1)
			}
			count++
		})
		d.Visit(gapBaseURL + e.Attr("src"))
	})

	c.Visit(gapProductPageBaseURL + pid)
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
