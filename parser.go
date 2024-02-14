package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Function to check if a URL exists (returns true) or not (returns false)
func urlExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println("Error checking URL:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}

// Function to generate the HTML page with links to mp3 files
func generateHTMLPage() string {
	const numLinks = 15

	// Get the previous day's date
	prevDate := time.Now().AddDate(0, 0, -1)

	// Base URL with dynamic part
	baseURL := "https://itunes.radiorecord.ru/tmp_audio/"

	var htmlBuilder strings.Builder

	// HTML page header
	htmlBuilder.WriteString("<!DOCTYPE html><html><head><title>MP3 Files</title></head><body><h1>MP3 Files</h1><ul>")

	// Loop through previous dates and generate links
	for i := 0; i < numLinks; i++ {
		// Construct the URL for the mp3 file
		dateStr := prevDate.Format("2006-01-02")
		url := fmt.Sprintf("%situnes1/hik_-_rr_%s.mp3", baseURL, dateStr)

		htmlBuilder.WriteString("<li>")
		htmlBuilder.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a>", url, url))

		// Check if the mp3 file is available
		if urlExists(url) {
			htmlBuilder.WriteString(" (Available)")
		} else {
			// Try the second dynamic part if the first one is not available
			url2 := strings.Replace(url, "itunes1", "itunes2", 1)
			if urlExists(url2) {
				htmlBuilder.WriteString(" (Available in itunes2)")
			} else {
				htmlBuilder.WriteString(" (Not Available)")
			}
		}

		htmlBuilder.WriteString("</li>")

		// Decrement the date for the next iteration
		prevDate = prevDate.AddDate(0, 0, -1)
	}

	// HTML page footer
	htmlBuilder.WriteString("</ul></body></html>")

	return htmlBuilder.String()
}

func main() {
	// Generate HTML page content
	htmlContent := generateHTMLPage()

	// Write HTML content to a file
	file, err := os.Create("index.html")
	if err != nil {
		fmt.Println("Error creating HTML file:", err)
		return
	}
	defer file.Close()

	// Write HTML content to the file
	_, err = file.WriteString(htmlContent)
	if err != nil {
		fmt.Println("Error writing to HTML file:", err)
		return
	}

	fmt.Println("HTML page generated successfully at index.html")
}
