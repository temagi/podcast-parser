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

// Function to generate the HTML page with links to mp3 files in table format with Bootstrap styling
func generateHTMLPage() string {
	const numLinks = 15

	// Get the previous day's date
	prevDate := time.Now().AddDate(0, 0, -1)

	var htmlBuilder strings.Builder

	// HTML page header with Bootstrap CSS included
	htmlBuilder.WriteString("<!DOCTYPE html><html><head><title>Latest RadioRecord K&H podcast direct links</title>")
	htmlBuilder.WriteString(`<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">`)
	htmlBuilder.WriteString("</head><body><div class=\"container\"><h1 class=\"mt-5 mb-4\">Latest RadioRecord K&H podcast direct links</h1><table class=\"table\"><thead><tr><th>#</th><th>Date</th><th>Day of the Week</th><th>Link</th><th>Status</th></tr></thead><tbody>")

	// Loop through previous dates and generate table rows
	for i := 0; i < numLinks; i++ {
		// Construct the URL for the mp3 file
		dateStr := prevDate.Format("2006-01-02")
		url := fmt.Sprintf("https://itunes.radiorecord.ru/tmp_audio/itunes1/hik_-_rr_%s.mp3", dateStr)

		htmlBuilder.WriteString("<tr>")
		htmlBuilder.WriteString(fmt.Sprintf("<td>%d</td>", i+1))
		htmlBuilder.WriteString(fmt.Sprintf("<td>%s</td>", prevDate.Format("2006-01-02")))
		htmlBuilder.WriteString(fmt.Sprintf("<td>%s</td>", prevDate.Weekday().String()))
		htmlBuilder.WriteString(fmt.Sprintf("<td><a href=\"%s\">Download file</a></td>", url))

		// Check if the mp3 file is available
		status := "Not Available"
		if urlExists(url) {
			status = "Available"
			htmlBuilder.WriteString(fmt.Sprintf("<td style=\"background-color: lightgreen;\">%s</td>", status))
		} else {
			// Try the second dynamic part if the first one is not available
			url2 := strings.Replace(url, "itunes1", "itunes2", 1)
			if urlExists(url2) {
				status = "Available"
				htmlBuilder.WriteString(fmt.Sprintf("<td style=\"background-color: lightgreen;\">%s</td>", status))
			} else {
				htmlBuilder.WriteString(fmt.Sprintf("<td style=\"background-color: lightcoral;\">%s</td>", status))
			}
		}

		htmlBuilder.WriteString("</tr>")

		// Decrement the date for the next iteration
		prevDate = prevDate.AddDate(0, 0, -1)
	}

	// HTML page footer
	htmlBuilder.WriteString("</tbody></table></div></body></html>")

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
