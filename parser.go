package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type PodcastStatus struct {
	Date          string `json:"date"`
	Url           string `json:"url"`
	Status        string `json:"status"`
	PodcastNumber *int   `json:"number,omitempty"`
}

func urlExists(url string) bool {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	fmt.Println("Checking URL:", url)
	resp, err := client.Head(url)
	if err != nil {
		fmt.Println("Error checking URL:", err)
		return false
	}
	defer resp.Body.Close()

	if len(resp.Request.URL.String()) > len(url) {
		fmt.Println("Redirection occurred:", resp.Request.URL.String())
	} else {
		fmt.Println("No redirection.")
	}

	result := resp.StatusCode == http.StatusOK
	fmt.Printf("URL: %s, Status Code: %d, Exists: %t\n", url, resp.StatusCode, result)
	return result
}

func loadStatuses(filename string) ([]PodcastStatus, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return []PodcastStatus{}, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var statuses []PodcastStatus
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&statuses)
	if err != nil {
		if err == io.EOF {
			return []PodcastStatus{}, nil
		}
		return nil, err
	}

	return statuses, nil
}

func saveStatuses(filename string, statuses []PodcastStatus) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(statuses)
}

func updatePodcastStatuses(statuses []PodcastStatus, numDays int, urlTemplate string, filename string) ([]PodcastStatus, error) {
	prevDate := time.Now().AddDate(0, 0, -1)
	statusMap := make(map[string]PodcastStatus)

	for _, status := range statuses {
		statusMap[status.Date] = status
	}

	for i := 0; i < numDays; i++ {
		dateStr := prevDate.Format("2006-01-02")
		if _, exists := statusMap[dateStr]; exists {
			prevDate = prevDate.AddDate(0, 0, -1)
			continue
		}

		url := fmt.Sprintf(urlTemplate, dateStr)
		status := "Not Available"
		if urlExists(url) {
			status = "Available"
		} else {
			url2 := strings.Replace(url, "itunes1", "itunes2", 1)
			if urlExists(url2) {
				url = url2
				status = "Available"
			}
		}

		statusMap[dateStr] = PodcastStatus{
			Date:   dateStr,
			Url:    url,
			Status: status,
		}

		prevDate = prevDate.AddDate(0, 0, -1)
	}

	statuses = make([]PodcastStatus, 0, len(statusMap))
	for _, status := range statusMap {
		statuses = append(statuses, status)
	}

	// Sort statuses by date in descending order
	sort.Slice(statuses, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", statuses[i].Date)
		dateJ, _ := time.Parse("2006-01-02", statuses[j].Date)
		return dateI.After(dateJ)
	})

	err := saveStatuses(filename, statuses)
	return statuses, err
}

func updateRTStatuses(statuses []PodcastStatus, filename string) ([]PodcastStatus, error) {
	statusMap := make(map[string]PodcastStatus)

	for _, status := range statuses {
		statusMap[status.Date] = status
	}

	// Find the latest podcast number
	latestPodcast := -1
	for _, status := range statuses {
		if status.PodcastNumber != nil && *status.PodcastNumber > latestPodcast {
			latestPodcast = *status.PodcastNumber
		}
	}
	if latestPodcast == -1 {
		latestPodcast = 0
	}

	// Check the next podcast
	nextPodcast := latestPodcast + 1
	dateStr := fmt.Sprintf("Podcast %d", nextPodcast)
	if _, exists := statusMap[dateStr]; !exists {
		url := fmt.Sprintf("https://cdn.radio-t.com/rt_podcast%d.mp3", nextPodcast)
		status := "Not Available"
		if urlExists(url) {
			status = "Available"
			statusMap[dateStr] = PodcastStatus{
				Date:          dateStr,
				Url:           url,
				Status:        status,
				PodcastNumber: &nextPodcast,
			}
		}
	}

	statuses = make([]PodcastStatus, 0, len(statusMap))
	for _, status := range statusMap {
		statuses = append(statuses, status)
	}

	// Sort statuses by podcast number in descending order
	sort.Slice(statuses, func(i, j int) bool {
		return *statuses[i].PodcastNumber > *statuses[j].PodcastNumber
	})

	err := saveStatuses(filename, statuses)
	return statuses, err
}

func generateHTMLPage(khStatuses, rtStatuses []PodcastStatus) string {
	const tpl = `
<!DOCTYPE html>
<html>
<head>
	<title>Latest Podcasts</title>
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
	<script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/@popperjs/core@2.5.4/dist/umd/popper.min.js"></script>
	<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
</head>
<body>
<div class="container">
	<h1 class="mt-5 mb-4">Latest Podcasts</h1>
	<ul class="nav nav-tabs" id="myTab" role="tablist">
	  <li class="nav-item">
		<a class="nav-link active" id="kh-tab" data-toggle="tab" href="#kh" role="tab" aria-controls="kh" aria-selected="true">K&H</a>
	  </li>
	  <li class="nav-item">
		<a class="nav-link" id="rt-tab" data-toggle="tab" href="#rt" role="tab" aria-controls="rt" aria-selected="false">Radio-T</a>
	  </li>
	</ul>
	<div class="tab-content" id="myTabContent">
	  <div class="tab-pane fade show active" id="kh" role="tabpanel" aria-labelledby="kh-tab">
		<table class="table mt-4">
			<thead>
				<tr><th>#</th><th>Date</th><th>Day of the Week</th><th>Link</th><th>Status</th></tr>
			</thead>
			<tbody>
				{{range $index, $status := .KH}}
				<tr>
					<td>{{$index}}</td>
					<td>{{$status.Date}}</td>
					<td>{{(parseTime $status.Date).Weekday}}</td>
					{{if eq $status.Status "Available"}}
					<td><a href="{{$status.Url}}">Download file</a></td>
					<td style="background-color: lightgreen;">{{$status.Status}}</td>
					{{else}}
					<td>-</td>
					<td style="background-color: lightcoral;">{{$status.Status}}</td>
					{{end}}
				</tr>
				{{end}}
			</tbody>
		</table>
	  </div>
	  <div class="tab-pane fade" id="rt" role="tabpanel" aria-labelledby="rt-tab">
		<table class="table mt-4">
			<thead>
				<tr><th>#</th><th>Date</th><th>Link</th><th>Status</th></tr>
			</thead>
			<tbody>
				{{range $index, $status := .RT}}
				<tr>
					<td>{{$index}}</td>
					<td>{{$status.Date}}</td>
					{{if eq $status.Status "Available"}}
					<td><a href="{{$status.Url}}">Download file</a></td>
					<td style="background-color: lightgreen;">{{$status.Status}}</td>
					{{else}}
					<td>-</td>
					<td style="background-color: lightcoral;">{{$status.Status}}</td>
					{{end}}
				</tr>
				{{end}}
			</tbody>
		</table>
	  </div>
	</div>
</div>
</body>
</html>
`
	funcMap := template.FuncMap{
		"parseTime": func(dateStr string) time.Time {
			t, _ := time.Parse("2006-01-02", dateStr)
			return t
		},
	}

	tmpl, err := template.New("webpage").Funcs(funcMap).Parse(tpl)
	if err != nil {
		panic(err)
	}

	data := struct {
		KH []PodcastStatus
		RT []PodcastStatus
	}{
		KH: khStatuses,
		RT: rtStatuses,
	}

	var htmlBuilder strings.Builder
	writer := bufio.NewWriter(&htmlBuilder)
	err = tmpl.Execute(writer, data)
	if err != nil {
		panic(err)
	}
	writer.Flush()

	return htmlBuilder.String()
}

func main() {
	// Load statuses
	khStatuses, err := loadStatuses("statuses.json")
	if err != nil {
		fmt.Println("Error reading KH statuses:", err)
		return
	}
	rtStatuses, err := loadStatuses("rt_statuses.json")
	if err != nil {
		fmt.Println("Error reading RT statuses:", err)
		return
	}

	// Update K&H podcast statuses
	khStatuses, err = updatePodcastStatuses(khStatuses, 14, "https://itunes.radiorecord.ru/tmp_audio/itunes1/hik_-_rr_%s.mp3", "statuses.json")
	if err != nil {
		fmt.Println("Error updating KH statuses:", err)
		return
	}

	// Update RT podcast statuses if it's Sunday and 11 AM
	now := time.Now()
	if now.Weekday() == time.Sunday && now.Hour() == 11 {
		rtStatuses, err = updateRTStatuses(rtStatuses, "rt_statuses.json")
		if err != nil {
			fmt.Println("Error updating RT statuses:", err)
			return
		}
	}

	// Generate HTML page content
	htmlContent := generateHTMLPage(khStatuses, rtStatuses)

	// Write HTML content to a file
	file, err := os.Create("index.html")
	if err != nil {
		fmt.Println("Error creating HTML file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(htmlContent)
	if err != nil {
		fmt.Println("Error writing to HTML file:", err)
		return
	}

	fmt.Println("HTML page generated successfully at index.html")
}
