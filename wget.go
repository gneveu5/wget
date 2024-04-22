package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	ppath "path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mxk/go-flowrate/flowrate"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/html"
)

// global var

var (
	// downloadedURLs tracks the URLs that have been downloaded to avoid duplicate downloads.
	downloadedURLs map[string]bool
	// ExcludeList stores a list of URLs to exclude from downloading.
	ExcludeList []string
)

// DownloadFile downloads a file from the provided URL and saves it to the specified path.
// It also supports rate limiting, mirroring, and modification of links.

func DownloadFile(url, filename string, path, rateLimit *string, isMirroring bool, rejectList []string, isModifylink bool, isBackground bool) error {
	if url == "" {
		return nil
	}
	fullPath := filepath.Join(*path, filename)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "MonAgentUtilisateur")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("sending request, awaiting response... ")
	if resp.StatusCode != http.StatusOK {
		fmt.Println(url)
		fmt.Println("ici le problmeme")

		fmt.Printf("Error: received status code %d\n", resp.StatusCode)
		os.Exit(1)
	}
	fmt.Println("status 200 OK")

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	reader := resp.Body
	if *rateLimit != "" {
		limit, err := ParseRateLimit(*rateLimit)
		if err != nil {
			return fmt.Errorf("invalid rate limit: %v", err)
		}
		reader = flowrate.NewReader(resp.Body, limit)
	}

	if resp.ContentLength == -1 {
		fmt.Println("content size: unknown")

		written, err := io.Copy(out, reader)
		if err != nil {
			return err
		}

		fmt.Printf("Downloaded %d bytes\n", written)
	} else {
		fmt.Printf("content size: %d [~%.2fMB]\n", resp.ContentLength, float64(resp.ContentLength)/1024/1024)

		if isBackground {

			multiWriter := io.MultiWriter(out)

			_, err = io.Copy(multiWriter, reader)
			if err != nil {
				return err
			}

		} else {
			bar := progressbar.DefaultBytes(
				resp.ContentLength,

				"downloading",
			)
			multiWriter := io.MultiWriter(out, bar)

			_, err = io.Copy(multiWriter, reader)
			if err != nil {
				return err
			}

		}
	}

	if isMirroring {
		if downloadedURLs == nil {
			downloadedURLs = make(map[string]bool)
		}
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "text/css") {
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			urls, err := ExtractURLs(file, url, *path)
			if err != nil {
				return err
			}
			if isModifylink {
				err = ModifyURLsInFile(fullPath)
				if err != nil {
					return err
				}

			}

			for _, u := range urls {
				if _, exists := downloadedURLs[url]; !exists {
					for _, ext := range rejectList {
						if strings.HasSuffix(u[0], ext) {
							fmt.Println("HERE", u, ext)
							return nil
						}
					}

					DownloadFile(u[0], ppath.Base(u[0]), &u[1], rateLimit, true, rejectList, isModifylink, false)

				}
			}

		}
	}
	fmt.Printf("saving file to: %s\n", fullPath)
	fmt.Printf("Downloaded [%s]\n", url)
	if isBackground {
		endTime := time.Now()

		fmt.Printf("finished at %s\n", endTime.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// DownloadFileInBackground is a concurrent version of DownloadFile that runs in the background.

func DownloadFileInBackground(url, filename string, path, rateLimit *string, wg *sync.WaitGroup, rejectList []string) {
	defer wg.Done()
	logFile, err := os.OpenFile("wget-log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()
	os.Stdout = logFile
	os.Stderr = logFile

	startTime := time.Now()
	fmt.Printf("start at %s\n", startTime.Format("2006-01-02 15:04:05"))

	if err := DownloadFile(url, filename, path, rateLimit, false, rejectList, false, true); err != nil {
		fmt.Fprintf(logFile, "Error downloading file: %v\n", err)
	}
}

// DownloadFromInput downloads files from a list of URLs provided in an input file.
func DownloadFromInput(inputFile string, path, rateLimit *string, rejectList []string) {
	urls, err := ReadURLsFromFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading URLs from file: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			filename := filepath.Base(url)
			fmt.Println(url, filename, *path, *rateLimit)
			if err := DownloadFile(url, filename, path, rateLimit, false, rejectList, false, false); err != nil {
				fmt.Printf("Error downloading file from %s: %v\n", url, err)
			} else {
				fmt.Printf("finished %s\n", filename)
			}
		}(url)
	}
	wg.Wait()
	fmt.Printf("Download finished: %v\n", urls)
}

// ParseRateLimit parses the rate limit string and returns the rate limit in bytes per second.

func ParseRateLimit(limitStr string) (int64, error) {
	var limit int64
	var unit string

	_, err := fmt.Sscanf(limitStr, "%d%s", &limit, &unit)
	if err != nil {
		return 0, err
	}

	unit = strings.ToLower(unit)
	switch unit {
	case "k":
		limit *= 1024
	case "m":
		limit *= 1024 * 1024
	}

	return limit, nil
}

// ReadURLsFromFile reads URLs from a file and returns them as a slice of strings.
func ReadURLsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

// ExtractURLs extracts URLs from HTML content and returns them as a slice of slices of strings.
func ExtractURLs(htmlContent io.Reader, baseURLString string, path string) ([][]string, error) {
	var urls [][]string
	doc, err := html.Parse(htmlContent)
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if n.Data == "link" || n.Data == "img" || n.Data == "script" {
					if (a.Key == "href" || a.Key == "src") && !strings.HasPrefix(a.Val, "http") && !strings.HasSuffix(a.Val, ".com") && a.Val != "/sites" && a.Val != "/sites/css/styles.css" {
						if a.Val == "/sites/css/styles.css" {
							fmt.Println(n)
							os.Exit(6)
							return
						}
						addURL(a.Val, baseURL, &urls, path)
					}

					if a.Key == "style" {
						extractURLsFromCSS(a.Val, baseURL, &urls, path)
					}
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "style" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					extractURLsFromCSS(c.Data, baseURL, &urls, path)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return urls, nil
}

// ModifyURLsInFile modifies URLs in an HTML file to make them relative.
func ModifyURLsInFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return err
	}

	var modifyNode func(*html.Node)
	modifyNode = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "a" || n.Data == "img" || n.Data == "link" || n.Data == "script") {
			for i, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					n.Attr[i].Val = modifyURL(attr.Val)
				}
			}
		} else if n.Type == html.ElementNode && n.Data == "style" {
			modifyCSSURLs(n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			modifyNode(c)
		}
	}

	modifyNode(doc)

	file.Seek(0, 0)
	file.Truncate(0)
	if err := html.Render(file, doc); err != nil {
		return err
	}

	return nil
}

// ModifyURLsInFile modifies URLs in an HTML file to make them relative.
func modifyURL(url string) string {
	if url[0] == '/' {
		return "." + url
	}
	return url
}

func modifyCSSURLs(styleNode *html.Node) {
	for c := styleNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			c.Data = modifyCSSContent(c.Data)
		}
	}
}

func modifyCSSContent(cssContent string) string {
	return strings.ReplaceAll(cssContent, "/", "./")
}

func addURL(rawurl string, baseURL *url.URL, urls *[][]string, path string) {
	resolvedURL, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	newpath := strings.TrimPrefix(rawurl, "./")

	dir := filepath.Dir(path + "/" + newpath)
	for _, dir := range ExcludeList {
		if strings.HasPrefix(newpath, strings.TrimPrefix(dir, "/")) {
			return
		}
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Erreur lors de la création du dossier:", err)
		os.Exit(1)
		return
	}

	resolvedURL2 := baseURL.ResolveReference(resolvedURL)

	*urls = append(*urls, []string{resolvedURL2.String(), dir})
}

func extractURLsFromCSS(css string, baseURL *url.URL, urls *[][]string, path string) {
	re := regexp.MustCompile(`url\(\s*(?:'([^']*)'|"([^"]*)"|([^'"\s][^)]*[^'"\s])|([^'"\s]))\s*\)`)
	matches := re.FindAllStringSubmatch(css, -1)
	for _, match := range matches {
		addURL(match[1], baseURL, urls, path)
	}
}

// GetDomainName extracts the domain name from a URL.

func GetDomainName(siteURL string) (string, error) {
	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Hostname(), nil
}
