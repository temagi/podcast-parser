package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Status represents the status of a URL.
type Status struct {
	Date   string `json:"date"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

// ReadStatuses reads the statuses from the JSON file.
func ReadStatuses(filename string) ([]Status, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []Status{}, nil // Return an empty slice if the file does not exist.
		}
		return nil, err
	}
	defer file.Close()

	// Check if file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() == 0 {
		return []Status{}, nil // Return an empty slice if the file is empty.
	}

	var statuses []Status
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&statuses); err != nil {
		if err.Error() == "EOF" {
			return []Status{}, nil // Return an empty slice if the file is empty.
		}
		return nil, err
	}
	return statuses, nil
}

// WriteStatuses writes the statuses to the JSON file.
func WriteStatuses(filename string, statuses []Status) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(&statuses); err != nil {
		return err
	}
	return nil
}

// Function to check if a URL exists (returns true) or not (returns false)
func urlExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println("Error checking URL:", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Function to generate the HTML page with links to mp3 files in table format with Bootstrap styling
func generateHTMLPage(statuses []Status) string {
	var htmlBuilder strings.Builder

	// HTML page header with Bootstrap CSS included
	htmlBuilder.WriteString("<!DOCTYPE html><html><head><title>Latest RadioRecord K&H podcast direct links</title>")
	htmlBuilder.WriteString(`<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">`)
	htmlBuilder.WriteString("</head><body><div class=\"container\"><h1 class=\"mt-5 mb-4\">Latest RadioRecord K&H podcast direct links</h1><table class=\"table\"><thead><tr><th>#</th><th>Date</th><th>Day of the Week</th><th>Link</th><th>Status</th></tr></thead><tbody>")

	// Loop through statuses and generate table rows
	for i, status := range statuses {
		date, err := time.Parse("2006-01-02", status.Date)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}

		htmlBuilder.WriteString("<tr>")
		htmlBuilder.WriteString(fmt.Sprintf("<td>%d</td>", i+1))
		htmlBuilder.WriteString(fmt.Sprintf("<td>%s</td>", status.Date))
		htmlBuilder.WriteString(fmt.Sprintf("<td>%s</td>", date.Weekday().String()))
		if status.Status == "Available" {
			htmlBuilder.WriteString(fmt.Sprintf("<td><a href=\"%s\">Download file</a></td>", template.HTMLEscapeString(status.URL)))
			htmlBuilder.WriteString(fmt.Sprintf("<td style=\"background-color: lightgreen;\">%s</td>", status.Status))
		} else {
			htmlBuilder.WriteString("<td>-</td>")
			htmlBuilder.WriteString(fmt.Sprintf("<td style=\"background-color: lightcoral;\">%s</td>", status.Status))
		}
		htmlBuilder.WriteString("</tr>")
	}

	// HTML page footer
	htmlBuilder.WriteString("</tbody></table></div></body></html>")

	return htmlBuilder.String()
}

func main() {
	const filename = "statuses.json"
	const numLinks = 15

	// Read existing statuses from the JSON file
	statuses, err := ReadStatuses(filename)
	if err != nil {
		fmt.Println("Error reading statuses:", err)
		return
	}

	// Get the current date
	currentDate := time.Now()

	// Map existing statuses by date for quick lookup
	statusMap := make(map[string]Status)
	for _, status := range statuses {
		statusMap[status.Date] = status
	}

	// Loop through the last numLinks days and update statuses if necessary
	for i := 0; i < numLinks; i++ {
		date := currentDate.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		// Check if the status for this date is already known
		if _, exists := statusMap[dateStr]; exists {
			continue
		}

		// Construct the URL for the mp3 file
		url := fmt.Sprintf("https://itunes.radiorecord.ru/tmp_audio/itunes1/hik_-_rr_%s.mp3", dateStr)
		urlStatus := Status{Date: dateStr, URL: url, Status: "Not Available"}
		if urlExists(url) {
			urlStatus.Status = "Available"
		} else {
			url2 := strings.Replace(url, "itunes1", "itunes2", 1)
			if urlExists(url2) {
				urlStatus.URL = url2
				urlStatus.Status = "Available"
			}
		}

		// Update the status map
		statusMap[dateStr] = urlStatus
	}

	// Convert the status map back to a slice
	var updatedStatuses []Status
	for _, status := range statusMap {
		updatedStatuses = append(updatedStatuses, status)
	}

	// Sort statuses by date in descending order (newest first)
	sort.Slice(updatedStatuses, func(i, j int) bool {
		return updatedStatuses[i].Date > updatedStatuses[j].Date
	})

	// Write the updated statuses to the JSON file
	if err := WriteStatuses(filename, updatedStatuses); err != nil {
		fmt.Println("Error writing statuses:", err)
		return
	}

	// Generate HTML page content
	htmlContent := generateHTMLPage(updatedStatuses)

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
