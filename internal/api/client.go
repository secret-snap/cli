package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type LoginRequest struct {
	LicenseKey string `json:"license_key"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type BundlePushRequest struct {
	ProjectID string `json:"project_id"`
	SizeBytes int    `json:"size_bytes"`
}

type BundlePushResponse struct {
	UploadURL string `json:"upload_url"`
	BundleID  string `json:"bundle_id"`
	S3Key     string `json:"s3_key"`
	Version   int    `json:"version"`
}

type BundleFinalizeRequest struct {
	BundleID   string `json:"bundle_id"`
	S3Key      string `json:"s3_key"`
	WrappedKey string `json:"wrapped_key"`
}

type BundlePullResponse struct {
	DownloadURL string `json:"download_url"`
	DataKey     string `json:"data_key"`
	Version     int    `json:"version"`
}

type ShareRequest struct {
	ProjectID string `json:"project_id"`
	UserEmail string `json:"user_email"`
	Role      string `json:"role"`
}

type AuditLog struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details"`
	CreatedAt string                 `json:"created_at"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

func (c *Client) Login(licenseKey string) (*LoginResponse, error) {
	req := LoginRequest{LicenseKey: licenseKey}
	resp, err := c.post("/v1/auth/login", req)
	if err != nil {
		return nil, err
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(resp, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %v", err)
	}

	return &loginResp, nil
}

func (c *Client) CreateProject(name string) (*Project, error) {
	req := CreateProjectRequest{Name: name}
	resp, err := c.post("/v1/projects", req)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project response: %v", err)
	}

	return &project, nil
}

func (c *Client) BundlePush(projectID string, sizeBytes int) (*BundlePushResponse, error) {
	req := BundlePushRequest{
		ProjectID: projectID,
		SizeBytes: sizeBytes,
	}
	resp, err := c.post("/v1/bundles/push", req)
	if err != nil {
		return nil, err
	}

	var pushResp BundlePushResponse
	if err := json.Unmarshal(resp, &pushResp); err != nil {
		return nil, fmt.Errorf("failed to parse bundle push response: %v", err)
	}

	return &pushResp, nil
}

func (c *Client) BundleFinalize(bundleID, s3Key string, wrappedKey []byte) error {
	req := BundleFinalizeRequest{
		BundleID:   bundleID,
		S3Key:      s3Key,
		WrappedKey: base64.StdEncoding.EncodeToString(wrappedKey),
	}
	_, err := c.post("/v1/bundles/finalize", req)
	return err
}

func (c *Client) BundlePull(projectID string) (*BundlePullResponse, error) {
	return c.BundlePullVersion(projectID, 0) // 0 means latest
}

func (c *Client) BundlePullVersion(projectID string, version int) (*BundlePullResponse, error) {
	var url string
	if version > 0 {
		url = fmt.Sprintf("%s/v1/bundles/pull?project_id=%s&version=%d", c.baseURL, projectID, version)
	} else {
		url = fmt.Sprintf("%s/v1/bundles/pull?project_id=%s&latest=true", c.baseURL, projectID)
	}

	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var pullResp BundlePullResponse
	if err := json.Unmarshal(resp, &pullResp); err != nil {
		return nil, fmt.Errorf("failed to parse bundle pull response: %v", err)
	}

	return &pullResp, nil
}

func (c *Client) Share(projectID, userEmail, role string) error {
	req := ShareRequest{
		ProjectID: projectID,
		UserEmail: userEmail,
		Role:      role,
	}
	_, err := c.post("/v1/shares", req)
	return err
}

func (c *Client) GetAuditLogs(projectID string, limit int) ([]AuditLog, error) {
	url := fmt.Sprintf("%s/v1/audit?project_id=%s&limit=%d", c.baseURL, projectID, limit)
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var logs []AuditLog
	if err := json.Unmarshal(resp, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse audit logs response: %v", err)
	}

	return logs, nil
}

func (c *Client) UploadToAPI(uploadURL string, data []byte) error {
	// If it's a relative URL, make it absolute
	if strings.HasPrefix(uploadURL, "/") {
		uploadURL = c.baseURL + uploadURL
	}

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) DownloadFromAPI(downloadURL string) ([]byte, error) {
	// If it's a relative URL, make it absolute
	if strings.HasPrefix(downloadURL, "/") {
		downloadURL = c.baseURL + downloadURL
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %v", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API download failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return data, nil
}

func (c *Client) post(path string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
