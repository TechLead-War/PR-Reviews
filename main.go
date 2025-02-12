package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	owner = "julofinance"
	repo  = "whatsapp-service"
)

func getToken() string {
	fmt.Println(os.Getenv("GITHUB_TOKEN"))
	return os.Getenv("GITHUB_TOKEN")
}

type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
	Body      string `json:"body"`
}

func fetchPullRequests() ([]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GoReviewDashboard")
	if token := getToken(); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", body)
	}
	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}
	return prs, nil
}

func fetchPullRequest(number int) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GoReviewDashboard")
	if token := getToken(); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", body)
	}
	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

var indexTmpl = `
<!DOCTYPE html>
<html>
<head>
	<title>Review Dashboard</title>
</head>
<body>
	<h1>Pull Requests for Review</h1>
	<ul>
	{{ range . }}
		<li>
			<strong>{{ .User.Login }}</strong> - {{ .Title }} at {{ .CreatedAt }}
			<a href="/review/{{ .Number }}">Review</a>
		</li>
	{{ end }}
	</ul>
</body>
</html>
`

var reviewTmpl = `
<!DOCTYPE html>
<html>
<head>
	<title>Review PR #{{ .Number }}</title>
</head>
<body>
	<h1>{{ .Title }}</h1>
	<p>By: <strong>{{ .User.Login }}</strong></p>
	<p>Created at: {{ .CreatedAt }}</p>
	<p>{{ .Body }}</p>
	<a href="/">Back</a>
</body>
</html>
`

func indexHandler(w http.ResponseWriter, r *http.Request) {
	prs, err := fetchPullRequests()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("index").Parse(indexTmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, prs)
	if err != nil {
		return
	}
}

func reviewHandler(w http.ResponseWriter, r *http.Request) {
	numberStr := r.URL.Path[len("/review/"):]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	pr, err := fetchPullRequest(number)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("review").Parse(reviewTmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	err := tmpl.Execute(w, pr)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/review/", reviewHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
