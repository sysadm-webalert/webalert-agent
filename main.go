package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

const (
	contentType    = "application/json"
	logDir         = "/var/log/webalert-agent/"
	logFileName    = "agent.log"
	configFilePath = "/etc/webalert-agent/config.json"
	version        = "1.0.1"
)

// Struct collection for Config
type Config struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	ApiURI   string   `json:"api_uri"`
	SiteName []string `json:"siteName"`
}

// Struct collection for json token
type LoginResponse struct {
	Token string `json:"token"`
}

// Struct collection for Metric
type Metric struct {
	CPUUsage  float64 `json:"cpu_usage"`
	RAMUsage  float64 `json:"memory_usage"`
	DiskUsage float64 `json:"disk_usage"`
	SiteName  string  `json:"sitename"`
	Version   string  `json:"version"`
}

// Function to load the configuration
func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Function to get metrics from CPU, memory and disk
func getMetrics(siteName string) (*Metric, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &Metric{
		CPUUsage:  cpuPercent[0],
		RAMUsage:  vmStat.UsedPercent,
		DiskUsage: diskStat.UsedPercent,
		SiteName:  siteName,
		Version:   version,
	}, nil
}

// Function to send metrics to the API
func sendMetrics(config *Config, metrics *Metric) error {
	loginURL := fmt.Sprintf("%s/api/login", config.ApiURI)

	// Creating the body to log in
	loginRequest := map[string]string{
		"email":    config.Email,
		"password": config.Password,
	}
	loginRequestBody, err := json.Marshal(loginRequest)
	if err != nil {
		return err
	}

	fmt.Printf("Request Body for Login: {\"email\":\"%s\",\"password\":\"%s\"}\n", obscureString(config.Email), obscureString(config.Password))

	// Send POST to login & get token
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginRequestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Cannot send data to the API: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to login, status code: %d, response body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("API error: %s", string(body))
	}

	var loginResponse LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		log.Printf("Error decoding login response: %v", err)
		return err
	}
	//Get the token from the response
	token := loginResponse.Token

	metricsURL := fmt.Sprintf("%s/api/v1/metrics", config.ApiURI)

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	fmt.Printf("Sending metrics: %s\n", jsonData)

	// Logging the body into the logs
	log.Printf("Sending metrics: %s", jsonData)

	req, err = http.NewRequest("POST", metricsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		log.Printf("Cannot send data to the API: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API error response: %s", string(body))
		return fmt.Errorf("API error: %s", string(body))
	}

	return nil
}

func getCPUUsage() (float64, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	return cpuPercent[0], nil
}

func getDiskUsage() (float64, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return 0, err
	}
	return diskStat.UsedPercent, nil
}

func getMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return vmStat.UsedPercent, nil
}

func obscureString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	visibleChars := 2
	obscured := len(s) - visibleChars
	return s[:visibleChars] + strings.Repeat("*", obscured)
}

var logFile *os.File

func initLog() {
	// If the folder not exist, create
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Error creating log directory: %v", err)
	}
	// Configure the log file
	logFile, err := os.OpenFile(filepath.Join(logDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(logFile)
}

func executeMode(mode string) {
	value, err := fetchMetric(mode)
	if err != nil {
		log.Printf("Error getting value: %v", err)
	} else {
		log.Printf("%s usage: %.2f%%", mode, value)
		fmt.Printf("%s usage: %.2f%%\n", mode, value)
	}
}

func fetchMetric(mode string) (float64, error) {
	switch mode {
	case "cpu":
		return getCPUUsage()
	case "disk":
		return getDiskUsage()
	case "memory":
		return getMemoryUsage()
	default:
		return 0, fmt.Errorf("unknown mode: %s", mode)
	}
}

func runLoopAgent(config *Config) {
	for {
		for _, site := range config.SiteName {
			metrics, err := getMetrics(site)
			if err != nil {
				log.Printf("Error getting metrics for site %s: %v", site, err)
				continue
			}
			log.Printf("Getting metrics for %s: CPU: %.2f%%, RAM: %.2f%%, Disk: %.2f%%", site, metrics.CPUUsage, metrics.RAMUsage, metrics.DiskUsage)
			if err := sendMetrics(config, metrics); err != nil {
				log.Printf("Error sending metrics for site %s: %v", site, err)
			} else {
				log.Printf("Metrics sent successfully for site %s", site)
			}
		}
		time.Sleep(5 * time.Minute) // Sending metrics every 5 minutes
	}
}

func main() {
	initLog()
	defer logFile.Close()

	// Arguments definitions
	mode := flag.String("m", "", "operation mode: cpu, disk, memory")
	flag.Parse()

	// Execute binary as standalone
	if *mode != "" {
		executeMode(*mode)
		return
	}

	// Load the config file
	config, err := loadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Error loading json config: %v", err)
	}
	runLoopAgent(config)
}
