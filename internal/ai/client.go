package ai

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Client struct {
	Provider   string
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model     string            `json:"model"`
	Messages  []Message         `json:"messages"`
	MaxTokens int               `json:"max_tokens,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (c Client) Complete(ctx context.Context, model string, messages []Message) (string, error) {
	switch strings.ToLower(strings.TrimSpace(c.Provider)) {
	case "", "openrouter":
		return c.completeOpenAICompatible(ctx, model, messages, true)
	case "openai":
		return c.completeOpenAICompatible(ctx, model, messages, false)
	case "anthropic":
		return c.completeAnthropic(ctx, model, messages)
	default:
		return "", fmt.Errorf("unsupported AI provider %q", c.Provider)
	}
}

func (c Client) completeOpenAICompatible(ctx context.Context, model string, messages []Message, addOpenRouterHeaders bool) (string, error) {
	if c.BaseURL == "" {
		return "", fmt.Errorf("AI base URL is empty")
	}
	if c.APIKey == "" {
		return "", fmt.Errorf("AI API key is empty")
	}
	httpClient := c.httpClient()

	body, err := json.Marshal(openAIChatRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 9000,
		Metadata: map[string]string{
			"tool": "term-paper",
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	if addOpenRouterHeaders {
		req.Header.Set("HTTP-Referer", "https://github.com/vasyukov1/hse-coursework-docs")
		req.Header.Set("X-Title", "term-paper")
	}

	respBody, err := doRequest(httpClient, req)
	if err != nil {
		return "", err
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("AI response contains no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func (c Client) completeAnthropic(ctx context.Context, model string, messages []Message) (string, error) {
	if c.BaseURL == "" {
		return "", fmt.Errorf("AI base URL is empty")
	}
	if c.APIKey == "" {
		return "", fmt.Errorf("AI API key is empty")
	}
	httpClient := c.httpClient()

	systemPrompt, anthropicMessages := splitSystemPrompt(messages)
	if len(anthropicMessages) == 0 {
		return "", fmt.Errorf("anthropic request requires at least one non-system message")
	}

	body, err := json.Marshal(anthropicRequest{
		Model:     model,
		MaxTokens: 9000,
		System:    systemPrompt,
		Messages:  anthropicMessages,
	})
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(c.BaseURL, "/")
	if !strings.HasSuffix(endpoint, "/messages") {
		endpoint += "/messages"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	respBody, err := doRequest(httpClient, req)
	if err != nil {
		return "", err
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	var parts []string
	for _, block := range parsed.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			parts = append(parts, block.Text)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("anthropic response contains no text blocks")
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n")), nil
}

func (c Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 2 * time.Minute}
}

func doRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AI request failed with status %s: %s", resp.Status, string(respBody))
	}
	return respBody, nil
}

func splitSystemPrompt(messages []Message) (string, []anthropicMessage) {
	var systemParts []string
	var out []anthropicMessage
	for _, message := range messages {
		if message.Role == "system" {
			if strings.TrimSpace(message.Content) != "" {
				systemParts = append(systemParts, message.Content)
			}
			continue
		}
		out = append(out, anthropicMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return strings.Join(systemParts, "\n\n"), out
}

func LoadReferenceExamples(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	client := &http.Client{Timeout: 20 * time.Second}
	var blocks []string
	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 24000))
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode >= 300 {
			continue
		}
		blocks = append(blocks, fmt.Sprintf("# Reference example: %s\n%s", url, truncate(string(body), 16000)))
	}
	return strings.Join(blocks, "\n\n")
}

func CollectProjectContext(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return collectFromDir(path)
	}
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		return collectFromZip(path)
	}
	return "", fmt.Errorf("unsupported project context source %q", path)
}

func CollectProjectContexts(paths []string) (string, error) {
	var sections []string
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		context, err := CollectProjectContext(path)
		if err != nil {
			return "", err
		}
		if context == "" {
			continue
		}
		sections = append(sections, fmt.Sprintf("# Context from %s\n\n%s", path, context))
	}
	return strings.Join(sections, "\n\n"), nil
}

func LoadSourceText(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return loadSourceDir(path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func loadSourceDir(root string) (string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".typ", ".md", ".txt":
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return "", fmt.Errorf("no .typ, .md or .txt files found in %s", root)
	}

	var sections []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			rel = file
		}
		sections = append(sections, fmt.Sprintf("# File: %s\n%s", rel, string(data)))
	}
	return strings.Join(sections, "\n\n"), nil
}

func collectFromDir(root string) (string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" || base == "build" || base == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)

	var sections []string
	sections = append(sections, "## Project files\n"+strings.Join(limitStrings(files, 500), "\n"))
	for _, file := range pickImportantFiles(files) {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			continue
		}
		sections = append(sections, fmt.Sprintf("## %s\n%s", file, truncate(string(data), 8000)))
	}
	return strings.Join(sections, "\n\n"), nil
}

func collectFromZip(zipPath string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var files []string
	content := map[string]string{}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || shouldSkipFile(file.Name) {
			continue
		}
		files = append(files, file.Name)
		if isImportantFile(file.Name) {
			rc, err := file.Open()
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(io.LimitReader(rc, 8000))
			_ = rc.Close()
			content[file.Name] = string(data)
		}
	}
	sort.Strings(files)

	var sections []string
	sections = append(sections, "## Project files\n"+strings.Join(limitStrings(files, 500), "\n"))
	for _, file := range pickImportantFiles(files) {
		if text, ok := content[file]; ok {
			sections = append(sections, fmt.Sprintf("## %s\n%s", file, truncate(text, 8000)))
		}
	}
	return strings.Join(sections, "\n\n"), nil
}

func pickImportantFiles(files []string) []string {
	var picked []string
	for _, file := range files {
		if isImportantFile(file) {
			picked = append(picked, file)
		}
		if len(picked) >= 24 {
			break
		}
	}
	return picked
}

func isImportantFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "readme.md", "go.mod", "package.json", "pyproject.toml", "cargo.toml", "requirements.txt", "pom.xml", "build.gradle", "dockerfile", "docker-compose.yml", "compose.yaml":
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".java", ".kt", ".swift", ".rs", ".sql":
		return strings.Contains(base, "main") || strings.Contains(base, "app") || strings.Contains(base, "server") || strings.Contains(base, "api") || strings.Contains(base, "handler") || strings.Contains(base, "controller")
	}
	return false
}

func shouldSkipFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return true
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".pdf", ".mp4", ".mov", ".zip", ".jar", ".exe", ".dll", ".so", ".ttf", ".woff", ".woff2", ".ico":
		return true
	}
	return false
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n...<truncated>"
}

func limitStrings(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	return append(items[:limit], fmt.Sprintf("... and %d more files", len(items)-limit))
}
