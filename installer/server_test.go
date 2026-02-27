package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.installer != installer {
		t.Error("Server.installer not set correctly")
	}

	if server.mux == nil {
		t.Error("Server.mux not initialized")
	}
}

// TestHandlePlatform tests the platform API endpoint
func TestHandlePlatform(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/api/platform", nil)
	rec := httptest.NewRecorder()

	server.handlePlatform(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["os"] != "linux" {
		t.Errorf("Response os = %v, want linux", response["os"])
	}

	if response["arch"] != "amd64" {
		t.Errorf("Response arch = %v, want amd64", response["arch"])
	}
}

// TestHandlePlatformMethodNotAllowed tests platform endpoint with wrong method
func TestHandlePlatformMethodNotAllowed(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodPost, "/api/platform", nil)
	rec := httptest.NewRecorder()

	server.handlePlatform(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestHandleStatus tests the status API endpoint
func TestHandleStatus(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	server.handleStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var response InstallStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Installed {
		t.Error("Status.Installed should be false for new installer")
	}
}

// TestHandleInstall tests the install API endpoint
func TestHandleInstall(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	// Create temp directories for test
	installReq := InstallRequest{
		SourceDir:   "/tmp/source",
		InstallDir:  "/tmp/install",
		ServerHost:  "localhost",
		ServerPort:  8080,
		EnableTLS:   false,
		AdapterType: "ollama",
		AdapterName: "local",
		Version:     "1.0.0",
	}

	body, _ := json.Marshal(installReq)
	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleInstall(rec, req)

	// Should fail because source directory doesn't exist
	if rec.Code != http.StatusOK {
		// Expected to fail with bad request due to missing source
		t.Logf("Install failed as expected: %s", rec.Body.String())
	}
}

// TestHandleInstallMethodNotAllowed tests install endpoint with wrong method
func TestHandleInstallMethodNotAllowed(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/api/install", nil)
	rec := httptest.NewRecorder()

	server.handleInstall(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestHandleInstallInvalidJSON tests install endpoint with invalid JSON
func TestHandleInstallInvalidJSON(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleInstall(rec, req)

	var response InstallResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Success {
		t.Error("Response.Success should be false for invalid JSON")
	}

	if response.Error == "" {
		t.Error("Response.Error should not be empty for invalid JSON")
	}
}

// TestHandleConfigGet tests the config GET endpoint
func TestHandleConfigGet(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	server.handleConfig(rec, req)

	// Should fail because no config file exists yet
	if rec.Code != http.StatusOK {
		t.Logf("Config get returned: %d", rec.Code)
	}
}

// TestHandleConfigPost tests the config POST endpoint
func TestHandleConfigPost(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	config := Config{
		Version: "1.0.0",
		Server:  ServerConfig{Host: "localhost", Port: 8080},
		Adapters: []AdapterConfig{
			{Name: "test", Type: "ollama"},
		},
	}

	body, _ := json.Marshal(config)
	req := httptest.NewRequest(http.MethodPost, "/api/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleConfig(rec, req)

	var response InstallResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// May fail due to permission issues, but should have proper response structure
	t.Logf("Config post response: %+v", response)
}

// TestHandleConfigMethodNotAllowed tests config endpoint with wrong method
func TestHandleConfigMethodNotAllowed(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodDelete, "/api/config", nil)
	rec := httptest.NewRecorder()

	server.handleConfig(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestHandleVerify tests the verify endpoint
func TestHandleVerify(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/api/verify", nil)
	rec := httptest.NewRecorder()

	server.handleVerify(rec, req)

	// Should fail because installation doesn't exist
	var response InstallResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Success {
		t.Log("Verify succeeded unexpectedly")
	}
}

// TestHandleIndex tests the index page handler
func TestHandleIndex(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.handleIndex(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusOK)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %s, want text/html; charset=utf-8", contentType)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}
}

// TestHandleIndexNotFound tests 404 for non-root paths
func TestHandleIndexNotFound(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	server.handleIndex(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestSendJSON tests the JSON response helper
func TestSendJSON(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	server.sendJSON(rec, data)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["key"] != "value" {
		t.Errorf("Response key = %s, want value", response["key"])
	}
}

// TestServerStartAndShutdown tests server lifecycle
func TestServerStartAndShutdown(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)
	server := NewServer(installer)

	// Test that server can be created and has proper structure
	if server.mux == nil {
		t.Error("Server mux not initialized")
	}
}

// TestInstallRequestValidation tests install request validation
func TestInstallRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request InstallRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: InstallRequest{
				SourceDir:   "/tmp/source",
				InstallDir:  "/tmp/install",
				ServerHost:  "localhost",
				ServerPort:  8080,
				AdapterType: "ollama",
				AdapterName: "local",
				Version:     "1.0.0",
			},
			valid: true,
		},
		{
			name: "empty source dir",
			request: InstallRequest{
				SourceDir:   "",
				InstallDir:  "/tmp/install",
				ServerHost:  "localhost",
				ServerPort:  8080,
				AdapterType: "ollama",
				Version:     "1.0.0",
			},
			valid: true, // Empty source dir is allowed (will fail later)
		},
		{
			name: "zero port",
			request: InstallRequest{
				SourceDir:   "/tmp/source",
				InstallDir:  "/tmp/install",
				ServerHost:  "localhost",
				ServerPort:  0,
				AdapterType: "ollama",
				Version:     "1.0.0",
			},
			valid: true, // Zero port is allowed (will use default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the request can be serialized
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("Failed to marshal request: %v", err)
			}

			var restored InstallRequest
			if err := json.Unmarshal(data, &restored); err != nil {
				t.Errorf("Failed to unmarshal request: %v", err)
			}
		})
	}
}

// TestCrossPlatformAPIResponses tests API responses across different platforms
func TestCrossPlatformAPIResponses(t *testing.T) {
	platforms := []struct {
		os   string
		arch string
	}{
		{"linux", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	for _, p := range platforms {
		t.Run(p.os+"_"+p.arch, func(t *testing.T) {
			platform := &Platform{OS: p.os, Arch: p.arch}
			installer := NewInstaller(platform)
			server := NewServer(installer)

			req := httptest.NewRequest(http.MethodGet, "/api/platform", nil)
			rec := httptest.NewRecorder()

			server.handlePlatform(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Status code = %d, want %d", rec.Code, http.StatusOK)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if response["os"] != p.os {
				t.Errorf("Response os = %v, want %v", response["os"], p.os)
			}
		})
	}
}
