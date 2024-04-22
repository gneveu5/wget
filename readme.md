# Go-Wget

Go-Wget is a basic clone of some essential functions of Wget implemented in Go (Golang). This project was developed as part of my training at Zone01.

## Description

Go-Wget provides functionalities similar to the popular command-line utility Wget, allowing users to download files from the internet. It supports features like downloading files, mirroring websites, limiting download speeds, rejecting specific file types, and more.

## Features

- **Downloading Files**: Download files from a specified URL to a local directory.
- **Website Mirroring**: Mirror a website by recursively downloading all linked resources (HTML pages, images, CSS files, etc.).
- **Rate Limiting**: Limit the download speed to prevent network congestion.
- **Rejecting File Types**: Exclude specific file types from being downloaded.
- **Converting Links**: Convert links in downloaded HTML files to make them relative.
- **Background Downloading**: Download files in the background, allowing users to continue using the terminal.

## Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   ```

2. Build the executable:

   ```bash
   go build
   ```

3. Run the executable:

   ```bash
   ./go-wget [options] <url>
   ```

   Replace `[options]` with any desired command-line options and `<url>` with the URL of the file or website to download.

## Usage

- Basic Usage:

  ```bash
  ./go-wget http://example.com/file.txt
  ```

- Mirror a Website:

  ```bash
  ./go-wget -mirror http://example.com
  ```

- Limit Download Speed:

  ```bash
  ./go-wget --rate-limit=1M http://example.com/large-file.zip
  ```

- Download in Background:

  ```bash
  ./go-wget -B http://example.com/large-file.zip
  ```

## Options

- `-O, --output`: Specify the output file name.
- `-P, --path`: Specify the output directory path.
- `--rate-limit`: Limit the download speed (e.g., "1M" for 1 MB/s).
- `-i, --input-file`: File containing multiple download links.
- `--mirror`: Mirror a website by recursively downloading linked resources.
- `--reject`: Exclude specific file types from being downloaded.
- `-X, --exclude`: Exclude specific directories from website mirroring.
- `--convert-links`: Convert links in downloaded HTML files to make them relative.
- `-B, --background`: Download files in the background.

## License

This project is licensed under the [MIT License](LICENSE).
