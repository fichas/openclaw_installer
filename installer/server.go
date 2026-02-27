package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

//go:embed static/templates/*
var templatesFS embed.FS

//go:embed static/assets/*
var assetsFS embed.FS

// Server represents the web server
type Server struct {
	installer *Installer
	server    *http.Server
	mux       *http.ServeMux
}

// InstallRequest represents the installation request from the web form
type InstallRequest struct {
	SourceDir    string            `json:"source_dir"`
	InstallDir   string            `json:"install_dir"`
	ServerHost   string            `json:"server_host"`
	ServerPort   int               `json:"server_port"`
	EnableTLS    bool              `json:"enable_tls"`
	AdapterType  string            `json:"adapter_type"`
	AdapterName  string            `json:"adapter_name"`
	Version      string            `json:"version"`
	Settings     map[string]string `json:"settings,omitempty"`
}

// InstallResponse represents the installation response
type InstallResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// NewServer creates a new web server instance
func NewServer(installer *Installer) *Server {
	s := &Server{
		installer: installer,
		mux:       http.NewServeMux(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() {
	// Static assets
	s.mux.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	// API endpoints
	s.mux.HandleFunc("/api/platform", s.handlePlatform)
	s.mux.HandleFunc("/api/install", s.handleInstall)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/verify", s.handleVerify)

	// Main page
	s.mux.HandleFunc("/", s.handleIndex)
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleIndex serves the main installation page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(templatesFS, "static/templates/index.html")
	if err != nil {
		// Fallback to embedded HTML if template not found
		s.serveEmbeddedHTML(w, r)
		return
	}

	data := map[string]interface{}{
		"Platform": s.installer.platform,
		"Version":  "1.0.0",
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// serveEmbeddedHTML serves a basic HTML page when templates are not available
func (s *Server) serveEmbeddedHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenClaw Installer</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 500px;
            width: 100%;
        }
        h1 { color: #333; margin-bottom: 10px; }
        .subtitle { color: #666; margin-bottom: 30px; }
        .platform-info {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; color: #555; font-weight: 500; }
        input[type="text"], input[type="number"], select {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus, select:focus {
            outline: none;
            border-color: #667eea;
        }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        input[type="checkbox"] {
            width: 20px;
            height: 20px;
        }
        button {
            width: 100%;
            padding: 15px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
        }
        button:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        .status {
            margin-top: 20px;
            padding: 15px;
            border-radius: 6px;
            display: none;
        }
        .status.success { background: #d4edda; color: #155724; display: block; }
        .status.error { background: #f8d7da; color: #721c24; display: block; }
        .status.loading { background: #d1ecf1; color: #0c5460; display: block; }
    </style>
</head>
<body>
    <div class="container">
        <h1>OpenClaw Installer</h1>
        <p class="subtitle">Configure and install OpenClaw on your system</p>

        <div class="platform-info">
            <strong>Detected Platform:</strong> ` + s.installer.platform.OS + `/` + s.installer.platform.Arch + `
        </div>

        <form id="installForm">
            <div class="form-group">
                <label for="sourceDir">Source Directory (USB Path)</label>
                <input type="text" id="sourceDir" name="source_dir" placeholder="/media/usb or D:\">
            </div>

            <div class="form-group">
                <label for="installDir">Installation Directory</label>
                <input type="text" id="installDir" name="install_dir" value="` + s.installer.platform.GetInstallDir() + `">
            </div>

            <div class="form-group">
                <label for="serverHost">Server Host</label>
                <input type="text" id="serverHost" name="server_host" value="0.0.0.0">
            </div>

            <div class="form-group">
                <label for="serverPort">Server Port</label>
                <input type="number" id="serverPort" name="server_port" value="8080">
            </div>

            <div class="form-group">
                <label for="adapterType">Adapter Type</label>
                <select id="adapterType" name="adapter_type">
                    <option value="usb">USB</option>
                    <option value="bluetooth">Bluetooth</option>
                    <option value="network">Network</option>
                </select>
            </div>

            <div class="form-group checkbox-group">
                <input type="checkbox" id="enableTLS" name="enable_tls">
                <label for="enableTLS">Enable TLS</label>
            </div>

            <button type="submit" id="installBtn">Install OpenClaw</button>
        </form>

        <div id="status" class="status"></div>
    </div>

    <script>
        document.getElementById('installForm').addEventListener('submit', async (e) => {
            e.preventDefault();

            const status = document.getElementById('status');
            const btn = document.getElementById('installBtn');

            status.className = 'status loading';
            status.textContent = 'Installing...';
            btn.disabled = true;

            const formData = {
                source_dir: document.getElementById('sourceDir').value,
                install_dir: document.getElementById('installDir').value,
                server_host: document.getElementById('serverHost').value,
                server_port: parseInt(document.getElementById('serverPort').value),
                enable_tls: document.getElementById('enableTLS').checked,
                adapter_type: document.getElementById('adapterType').value,
                adapter_name: 'default',
                version: '1.0.0'
            };

            try {
                const response = await fetch('/api/install', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formData)
                });

                const result = await response.json();

                if (result.success) {
                    status.className = 'status success';
                    status.textContent = result.message;
                } else {
                    status.className = 'status error';
                    status.textContent = result.error || 'Installation failed';
                }
            } catch (error) {
                status.className = 'status error';
                status.textContent = 'Network error: ' + error.message;
            }

            btn.disabled = false;
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// handlePlatform returns platform information
func (s *Server) handlePlatform(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"os":           s.installer.platform.OS,
		"arch":         s.installer.platform.Arch,
		"install_dir":  s.installer.platform.GetInstallDir(),
		"config_dir":   s.installer.platform.GetConfigDir(),
	}

	s.sendJSON(w, response)
}

// handleInstall processes installation requests
func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendJSON(w, InstallResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Use defaults if not provided
	if req.InstallDir == "" {
		req.InstallDir = s.installer.platform.GetInstallDir()
	}
	configDir := s.installer.platform.GetConfigDir()

	// Install binaries
	installOpts := InstallOptions{
		SourceDir:  req.SourceDir,
		InstallDir: req.InstallDir,
		ConfigDir:  configDir,
		BinaryName: "openclaw",
	}

	if err := s.installer.InstallWithOptions(installOpts); err != nil {
		s.sendJSON(w, InstallResponse{
			Success: false,
			Error:   "Installation failed: " + err.Error(),
		})
		return
	}

	// Generate configuration
	configOpts := ConfigOptions{
		Version:        req.Version,
		ServerHost:     req.ServerHost,
		ServerPort:     req.ServerPort,
		EnableTLS:      req.EnableTLS,
		AdapterType:    req.AdapterType,
		AdapterName:    req.AdapterName,
		CustomSettings: req.Settings,
	}

	config, err := GenerateConfig(configOpts)
	if err != nil {
		s.sendJSON(w, InstallResponse{
			Success: false,
			Error:   "Config generation failed: " + err.Error(),
		})
		return
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		s.sendJSON(w, InstallResponse{
			Success: false,
			Error:   "Invalid configuration: " + err.Error(),
		})
		return
	}

	// Save configuration
	configPath := GetDefaultConfigPath()
	if err := config.Save(configPath); err != nil {
		s.sendJSON(w, InstallResponse{
			Success: false,
			Error:   "Failed to save config: " + err.Error(),
		})
		return
	}

	s.sendJSON(w, InstallResponse{
		Success: true,
		Message: fmt.Sprintf("OpenClaw installed successfully! Configuration saved to %s", configPath),
	})
}

// handleConfig handles configuration API
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Load existing config
		configPath := GetDefaultConfigPath()
		config, err := LoadConfig(configPath)
		if err != nil {
			s.sendJSON(w, InstallResponse{
				Success: false,
				Error:   "Failed to load config: " + err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
		// Save new config
		var config Config
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			s.sendJSON(w, InstallResponse{
				Success: false,
				Error:   "Invalid config: " + err.Error(),
			})
			return
		}

		if err := config.Validate(); err != nil {
			s.sendJSON(w, InstallResponse{
				Success: false,
				Error:   "Invalid configuration: " + err.Error(),
			})
			return
		}

		configPath := GetDefaultConfigPath()
		if err := config.Save(configPath); err != nil {
			s.sendJSON(w, InstallResponse{
				Success: false,
				Error:   "Failed to save config: " + err.Error(),
			})
			return
		}

		s.sendJSON(w, InstallResponse{
			Success: true,
			Message: "Configuration saved successfully",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleStatus returns installation status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.sendJSON(w, s.installer.GetStatus())
}

// handleVerify verifies the installation
func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	installDir := s.installer.platform.GetInstallDir()
	files := []string{
		s.installer.platform.GetBinaryName("openclaw"),
	}

	if err := s.installer.VerifyInstallation(installDir, files); err != nil {
		s.sendJSON(w, InstallResponse{
			Success: false,
			Error:   "Verification failed: " + err.Error(),
		})
		return
	}

	s.sendJSON(w, InstallResponse{
		Success: true,
		Message: "Installation verified successfully",
	})
}

// sendJSON sends a JSON response
func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// OpenBrowser opens the default browser to the specified URL
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // linux and others
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
