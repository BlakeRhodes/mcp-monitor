# MCP System Monitor

A system monitoring service that exposes hardware and OS metrics via the Model Context Protocol (MCP). This enables LLMs or MCP-compatible clients to retrieve real-time system and process information in a structured manner.

![Screenshot](./doc/snapshot-1.png)

---

## Features

MCP System Monitor provides these monitoring tools (each as an MCP tool):

- **CPU:** Usage %, core count, detailed hardware info
- **Memory:** Virtual & swap memory stats
- **Disk:** Usage, partitions, and I/O statistics (with filtering)
- **Network:** Interfaces, connections, and traffic
- **NVIDIA GPU:** Info and usage statistics (via NVML, if available)
- **Host:** General system details, uptime, boot time, current users
- **Processes:** List, sort, and details for system processes
- **Logs:** Read, filter, and tail log files (with substring/regex search)

---

## Tool Reference

### 1. CPU Information

**Tool:** `get_cpu_info`

- Description: Get CPU usage and information.
- Parameters:
  - `per_cpu` (boolean, default: false): Return metrics for each core.

---

### 2. Memory Information

**Tool:** `get_memory_info`

- Description: Get total, used, free, and swap memory info.
- Parameters: _none_

---

### 3. Disk Information

**Tool:** `get_disk_info`

- Description: Get disk usage, partition, and I/O statistics.
- Parameters:
  - `path` (string, default: "/"): Path to check usage for.
  - `all_partitions` (boolean, default: false): If true, include all partitions.

---

### 4. Network Information

**Tool:** `get_network_info`

- Description: Get summary of network interfaces, traffic, and connections.
- Parameters:
  - `interface` (string, optional): If set, reports only that interface.

---

### 5. NVIDIA GPU Information

**Tool:** `get_nvidia_gpu_info`

- Description: Returns NVIDIA GPU hardware and usage details (requires NVIDIA drivers/NVML).
- Parameters: _none_

---

### 6. Host/System Information

**Tool:** `get_host_info`

- Description: General system info, uptime, boot time, users.
- Parameters: _none_

---

### 7. Process Information

**Tool:** `get_process_info`

- Description: Summary or detailed information on system processes.
- Parameters:
  - `pid` (number, optional): Return detail for a specific PID.
  - `limit` (number, default: 10): Limit number of returned processes.
  - `sort_by` (string, default: "cpu"): Sort by "cpu", "memory", "pid", or "name".

---

### 8. Log File Reader

**Tool:** `get_log_info`

- Description: Read and filter lines from a log file (e.g. /var/log/syslog).
- Parameters:
  - `file` (string, required): Log file path.
  - `lines` (number, default: 20): How many lines from the end.
  - `filter` (string, optional): Filter lines containing this substring.
  - `regex` (string, optional): Filter lines matching this Go regex.

On systems using systemd, if the specified log file (e.g. /var/log/syslog) is not found, get_log_info will automatically fallback to reading logs from the systemd journal (journalctl), providing equivalent output.

This works transparently for tools and scripts expecting traditional syslog files.

---

## Installation

```bash
git clone https://github.com/seekrays/mcp-monitor.git
cd mcp-monitor
make build
UsageRun the binary in stdio mode (MCP server):./mcp-monitor
The service starts in stdio mode and will communicate over MCP with clients.LibreChat IntegrationTo connect MCP System Monitor to LibreChat or any compatible AI orchestration tool that supports SSE-based Model Context Protocol (MCP) endpoints, add the following configuration to your config.yaml or settings file:monitor:
    type: sse
    url: http://host.docker.internal:3001/sse

type: sse tells LibreChat to use Server-Sent Events for streaming data.
url: should point to the MCP monitoring server.

If running in Docker, host.docker.internal targets the host’s network.
Adjust the port (3001 here) if you run the monitor on a different port.

After applying this config, LibreChat will be able to call MCP monitor tools to view and query live system information.Note:
If running everything on localhost outside Docker, you can use http://localhost:3001/sse.
If deploying to a server, replace the URL with your MCP monitor’s public endpoint.

## Security Considerations

MCP System Monitor provides low-level telemetry on the host system. Keep the following in mind for secure deployments:

### Restricted File Reads

- The log file tool (`get_log_info`) allows users to specify a file path. By default, it can read any file your process user has access to. To avoid disclosure of sensitive files, restrict allowed log file locations to a known safe directory (e.g., `/var/log`) and reject paths containing `..` or not matching your allowlist.
- Limit the maximum number of lines and regex size for log requests (e.g., 1000 lines, modest regex).
- Be aware that complex user-supplied regex patterns may cause excessive CPU use (ReDoS risk) or cause unresponsiveness.

### Privileged Information

- Disk, process, and network tools can reveal system-level information and the list of running processes, users, and interfaces. Restrict access to the MCP endpoint so only trusted clients can connect.
- Run mcp-monitor as a non-root user when possible. This minimizes the harm if a tool is abused or a vulnerability is discovered.

### Endpoint Security

- Do not expose the MCP Monitor endpoint to the general internet without strict authentication and access controls. All tools are intended for integration with trusted, local LLM or orchestration frameworks.

### Input Validation

- Be cautious when expanding monitoring or log-access tooling for non-local or multi-user scenarios. Sanitize and validate all incoming parameters that access files or process critical resources.

### Updates & Maintenance

- Keep dependencies (such as gopsutil, go-nvml, and MCP libraries) up to date for the latest security fixes.
- Regularly audit new contributions and features for exposure points or misuse of system calls.

## Contributing

Contributions are welcome! Please file an issue or submit a Pull Request.Maintainer: seekrays
