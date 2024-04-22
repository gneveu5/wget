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
)

func main() {
	// Parse command-line flags
	background := flag.Bool("B", false, "Download in background")
	output := flag.String("O", "", "Output file name")
	path := flag.String("P", "./", "Output directory path")
	rateLimit := flag.String("rate-limit", "", "Download speed limit")
	inputFile := flag.String("i", "", "File containing multiple download links")
	mirror := flag.Bool("mirror", false, "Mirror a website")
	reject := flag.String("reject", "", "Reject specific file types")
	exclude := flag.String("X", "", "Exclude specific directories")
	convert := flag.Bool("convert-links", false, "Convert the links")

	flag.Parse()

	var URL string
	if len(flag.Args()) > 0 {
		URL = flag.Args()[0]
	}
	// Check if URL is provided, or mirror flag is set, or input file is specified
	if URL == "" && !*mirror && *inputFile == "" {
		fmt.Println("Please provide a URL, input file, or use the --mirror flag")
		flag.PrintDefaults()
		os.Exit(1)
	}
	// Initialize reject list and exclude list

	var rejectList []string

	if *reject != "" {
		rejectList = strings.Split(*reject, ",")
	}

	if *exclude != "" {
		ExcludeList = strings.Split(*exclude, ",")
	}
	// Determine output filename

	filename := *output
	if *output == "" {
		parsedURL, err := url.Parse(URL)
		if err != nil {
			fmt.Println("Invalid URL:", err)
			os.Exit(1)
		}

		if parsedURL.Path == "" || strings.HasSuffix(parsedURL.Path, "/") {
			filename = "index.html"
		} else {
			filename = filepath.Base(URL)
		}
	}
	// Record start time
	startTime := time.Now()
	fmt.Printf("start at %s\n", startTime.Format("2006-01-02 15:04:05"))

	// Handle website mirroring

	if *mirror && URL != "" {
		domain, err := GetDomainName(URL)
		if err != nil {
			log.Fatal(err)
		}
		domainPath := filepath.Join(*path, domain)
		err1 := os.MkdirAll(domainPath, os.ModePerm)
		if err1 != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}

		fmt.Println(&domainPath)
		DownloadFile(URL, filename, &domainPath, rateLimit, true, rejectList, *convert, false)

		endTime := time.Now()

		fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}
	// Handle background downloading
	if *background {
		var wg sync.WaitGroup
		wg.Add(1)
		go DownloadFileInBackground(URL, filename, path, rateLimit, &wg, rejectList)
		fmt.Println("Output will be written to \"wget-log\".")
		wg.Wait()
		endTime := time.Now()
		fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}
	// Handle input file containing multiple download links
	if *inputFile != "" {
		DownloadFromInput(*inputFile, path, rateLimit, rejectList)
	}
	// Download single file if not mirroring
	if !*mirror {
		if err := DownloadFile(URL, filename, path, rateLimit, false, rejectList, false, false); err != nil {
			fmt.Printf("Error downloading file: %v\n", err)
		}
	}
	// Record end time

	endTime := time.Now()
	fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
}
