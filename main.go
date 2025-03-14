package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type GitHubUser struct {
	Login string `json:"login"`
}

type GitHubPR struct {
	ID                 int          `json:"id"`
	Number             int          `json:"number"`
	Title              string       `json:"title"`
	State              string       `json:"state"`
	HTMLURL            string       `json:"html_url"`
	User               GitHubUser   `json:"user"`
	RequestedReviewers []GitHubUser `json:"requested_reviewers"`
}

type PullRequest struct {
	ID     int    `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Author string `json:"author"`
	State  string `json:"state"`
	PRURL  string `json:"pr_url"`
}

type GitHubComment struct {
	ID        int        `json:"id"`
	Body      string     `json:"body"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	HTMLURL   string     `json:"html_url"`
	User      GitHubUser `json:"user"`
}

type Comment struct {
	ID         int    `json:"id"`
	Author     string `json:"author"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	CommentURL string `json:"comment_url"`
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins; adjust if needed.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getPullsHandler(githubToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, repo := r.URL.Query().Get("owner"), r.URL.Query().Get("repo")
		if owner == "" || repo == "" {
			http.Error(w, `{"message": "Owner and Repository are required!"}`, http.StatusBadRequest)
			return
		}
		urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
		log.Printf("Fetching all PRs from: %s", urlStr)
		req, _ := http.NewRequest("GET", urlStr, nil)
		req.Header.Set("Authorization", "Bearer "+githubToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error fetching PRs: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var prs []GitHubPR
		if err = json.NewDecoder(resp.Body).Decode(&prs); err != nil {
			log.Printf("Error decoding PRs: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(prs) == 0 {
			log.Printf("No open PRs found for %s/%s", owner, repo)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "No open PRs found for this repository."})
			return
		}

		var results []PullRequest
		for _, pr := range prs {
			results = append(results, PullRequest{
				ID:     pr.ID,
				Number: pr.Number,
				Title:  pr.Title,
				Author: pr.User.Login,
				State:  pr.State,
				PRURL:  pr.HTMLURL,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func getCommentsHandler(githubToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		prNumber := vars["prNumber"]
		owner, repo := r.URL.Query().Get("owner"), r.URL.Query().Get("repo")
		if owner == "" || repo == "" {
			http.Error(w, `{"message": "Owner and Repository are required!"}`, http.StatusBadRequest)
			return
		}
		urlStr := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/comments", owner, repo, prNumber)
		log.Printf("Fetching comments from: %s", urlStr)
		req, _ := http.NewRequest("GET", urlStr, nil)
		req.Header.Set("Authorization", "Bearer "+githubToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error fetching comments: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var ghComments []GitHubComment
		if err = json.NewDecoder(resp.Body).Decode(&ghComments); err != nil {
			log.Printf("Error decoding comments: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var results []Comment
		for _, c := range ghComments {
			results = append(results, Comment{
				ID:         c.ID,
				Author:     c.User.Login,
				Body:       c.Body,
				CreatedAt:  c.CreatedAt,
				UpdatedAt:  c.UpdatedAt,
				CommentURL: c.HTMLURL,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// New endpoint: get pull requests where the given reviewer is requested.
func getMyReviewRequestsHandler(githubToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := r.URL.Query().Get("owner")
		repo := r.URL.Query().Get("repo")
		reviewer := r.URL.Query().Get("reviewer")
		if owner == "" || repo == "" || reviewer == "" {
			http.Error(w, `{"message": "Owner, Repository, and Reviewer are required!"}`, http.StatusBadRequest)
			return
		}
		log.Printf("Fetching review requests for %s/%s with reviewer %s", owner, repo, reviewer)
		q := fmt.Sprintf("repo:%s/%s is:pr is:open review-requested:%s", owner, repo, reviewer)
		searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s", url.QueryEscape(q))
		log.Printf("Search URL: %s", searchURL)
		req, _ := http.NewRequest("GET", searchURL, nil)
		req.Header.Set("Authorization", "Bearer "+githubToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error in review request search: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var result struct {
			Items []struct {
				ID      int        `json:"id"`
				Number  int        `json:"number"`
				Title   string     `json:"title"`
				State   string     `json:"state"`
				HTMLURL string     `json:"html_url"`
				User    GitHubUser `json:"user"`
			} `json:"items"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Error decoding search response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Found %d review-requested PRs", len(result.Items))
		var prs []PullRequest
		for _, item := range result.Items {
			prs = append(prs, PullRequest{
				ID:     item.ID,
				Number: item.Number,
				Title:  item.Title,
				Author: item.User.Login,
				State:  item.State,
				PRURL:  item.HTMLURL,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prs)
	}
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN is required")
	}

	r := mux.NewRouter()
	r.Use(enableCors)

	// API endpoints.
	r.HandleFunc("/api/pulls", getPullsHandler(githubToken)).Methods("GET")
	r.HandleFunc("/api/pulls/{prNumber}/comments", getCommentsHandler(githubToken)).Methods("GET")
	r.HandleFunc("/api/my-review-requests", getMyReviewRequestsHandler(githubToken)).Methods("GET")

	// Serve FE.html for all other routes.
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "FE.html")
	})

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
