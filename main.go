package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	funcs "wget/functions"
)

func main() {
	// Define flags
	background := flag.Bool("B", false, "Download in background")
	output := flag.String("O", "", "Output file name")
	path := flag.String("P", "./", "Output directory path")
	rateLimit := flag.String("rate-limit", "", "Download speed limit")
	inputFile := flag.String("i", "", "File containing multiple download links")
	mirror := flag.Bool("mirror", false, "Mirror a website")
	reject := flag.String("reject", "", "Reject specific file types")
	exclude := flag.String("X", "", "Exclude specific directories")
	convert := flag.Bool("convert-links", false, "Convert the links")

	// Parse command-line arguments
	flag.Parse()

	var URL string
	// Check for URL as positional argument
	if len(flag.Args()) > 0 {
		URL = flag.Args()[0]
	}

	// Validate that URL or other required arguments are provided
	if URL == "" && !*mirror && *inputFile == "" {
		fmt.Println("Please provide a URL, input file, or use the --mirror flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var rejectList []string

	if *reject != "" {
		rejectList = strings.Split(*reject, ",")
	}

	if *exclude != "" {
		funcs.ExcludeList = strings.Split(*exclude, ",")
	}

	filename := *output
	if *output == "" {
		parsedURL, err := url.Parse(URL)
		if err != nil {
			fmt.Println("Invalid URL:", err)
			os.Exit(1)
		}

		// Set filename to 'index.html' if URL is the root of a website or ends with a slash
		if parsedURL.Path == "" || strings.HasSuffix(parsedURL.Path, "/") {
			filename = "index.html"
		} else {
			filename = filepath.Base(URL)
		}
	}
	// Start time
	startTime := time.Now()
	fmt.Printf("start at %s\n", startTime.Format("2006-01-02 15:04:05"))

	if *mirror && URL != "" {
		domain, err := funcs.GetDomainName(URL)
		if err != nil {
			log.Fatal(err)
		}
		domainPath := filepath.Join(*path, domain)
		err1 := os.MkdirAll(domainPath, os.ModePerm)
		if err1 != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}

		fmt.Println(&domainPath)
		funcs.DownloadFile(URL, filename, &domainPath, rateLimit, true, rejectList, *convert, false)

		// End time
		endTime := time.Now()

		fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}

	if *background {
		var wg sync.WaitGroup
		wg.Add(1)
		go funcs.DownloadFileInBackground(URL, filename, path, rateLimit, &wg, rejectList)
		fmt.Println("Output will be written to \"wget-log\".")
		wg.Wait() // Wait for the background task to complete
		endTime := time.Now()
		fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}

	if *inputFile != "" {
		funcs.DownloadFromInput(*inputFile, path, rateLimit, rejectList)
	}

	// Download with progress bar
	if !*mirror {
		if err := funcs.DownloadFile(URL, filename, path, rateLimit, false, rejectList, false, false); err != nil {
			fmt.Printf("Error downloading file: %v\n", err)
		}
	}

	// End time
	endTime := time.Now()
	fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
}
