# WebAlert Agent

## Overview
The WebAlert Agent is a lightweight monitoring tool designed to collect system metrics. It provides essential insights to ensure optimal system reliability and website uptime.

## Features
- **System Metrics Monitoring**: Collects CPU usage, memory consumption, and disk utilization.
- **Configurable**: Easily set up using a `/etc/webalert-agent/config.json` file.
- **Secure Communication**: Sends data to your centralized instance API over HTTPS.
- **Systemd Integration**: Runs as a service on Linux-based systems (Ubuntu/Debian).
- **Flexible Logging**: Stores logs in `/var/log/webalert-agent/` to easy debug.

## Installation

### Prerequisites
- Linux-based system (Ubuntu/Debian)
- Go ^1.23.2 (for building from source)

### Using Prebuilt Package
1. Download the latest `.deb` package from the [Releases](https://github.com/your-repo/webalert-agent/releases).
2. Install the package:
   ```bash
   sudo dpkg -i webalert-agent-latest.deb
   ```
3. Verify installation:
   ```bash
   systemctl status webalert-agent
   ```

### Building from Source
1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/webalert-agent.git
   cd webalert-agent
   ```
2. Init modules:
   ```bash
   go mod init webalert-agent
3. Install dependencies:
   ```bash
   go get github.com/shirou/gopsutil/cpu
   go get github.com/shirou/gopsutil/mem
   go get github.com/shirou/gopsutil/disk
   go mod tidy
   ```
4. Build the binary:
   ```bash
   go build -o ./usr/local/bin/webalert-agent main.go
   ```
5. Copy the binary to `/usr/local/bin`:
   ```bash
   sudo cp webalert-agent /usr/local/bin/
   ```
6. Set up the service:
   ```bash
   sudo cp webalert-agent.service /etc/systemd/system/
   sudo systemctl enable webalert-agent
   sudo systemctl start webalert-agent
   ```
7. Create the directory for logging:
   ```bash
   mkdir /var/log/webalert-agent
   ```

## Configuration
Create a `config.json` file in `/etc/webalert-agent/` with the following structure, you can add one or multiple sites:
1. One site example:
```json
{
  "email": "your_user_app",
  "password": "your_password_app",
  "api_uri": "https://api.webalert.digital",
  "siteName": ["https://yoursite.com"]
}
```
2. Two or more sites example:
```json
{
  "email": "your_user_app",
  "password": "your_password_app",
  "api_uri": "https://api.webalert.digital",
  "siteName": ["https://yoursite.com","https://yoursecond.com"]
}
```
## Usage

### Start the Agent
```bash
sudo systemctl start webalert-agent
```

### Stop the Agent
```bash
sudo systemctl stop webalert-agent
```

### Status the Agent
```bash
sudo systemctl status webalert-agent
```
### Restart the Agent
```bash
sudo systemctl restart webalert-agent
```
### View Logs
```bash
sudo tail -f /var/log/webalert-agent/agent.log
```

### Test Metrics Collection
Run the binary directly to view metrics:
```bash
webalert-agent --test
```

## Contributing
We welcome contributions! Please follow these steps:
1. Fork the repository.
2. Create a feature branch.
3. Commit your changes.
4. Open a pull request.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support
For issues or feature requests, please open an issue in the [GitHub repository](https://github.com/your-repo/webalert-agent/issues).

---
**WebAlert Agent** © 2024

