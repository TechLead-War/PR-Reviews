# GitHub PR Review Tracker

As your team grows, the number of pull requests requiring review can quickly become overwhelming. The GitHub PR Review Tracker is a simple, must-have tool that helps you keep track of which pull requests require your review—so you never miss an important update.

## Features

- **Fetch All PRs:** View all open pull requests for a repository.
- **Filter by Review Request:** Quickly filter and display only the PRs where your review is requested.
- **View PR Details & Comments:** Easily check PR details and view associated comments.
- **Single-Page Interface:** A clean, self-contained frontend served directly from the Go backend.

## Getting Started

### Prerequisites

- [Go](https://golang.org/) installed on your machine.
- A GitHub Personal Access Token (PAT) with at least read access.

### Installation

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/yourusername/github-pr-review-tracker.git
   cd github-pr-review-tracker

2. **Create a .env File**

   PORT=5000

   GITHUB_TOKEN=your_github_token


3. **Install Dependencies & Run the Server:**
```bash
go get github.com/gorilla/mux
go get github.com/joho/godotenv

go run main.go