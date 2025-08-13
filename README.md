# Squid Proxy Whitelist Editor

**Repository**: https://github.com/coledykstra7/docker-whitelist

## Project Purpose
A comprehensive web-based editor for managing Squid proxy whitelist and blacklist files with real-time monitoring, domain management, and integrated log analysis. Features a modern interface with live updates, smart domain sorting, and automatic squid configuration reloading.

## Features
- **Interactive Domain Management**: Add, remove, and move domains between whitelist/blacklist with one-click actions
- **Real-time Log Monitoring**: Live tail of categorized access logs (WL/BL/RG) with color-coded display
- **Smart Domain Organization**: Automatic sorting by note categories, then reverse domain parts for logical grouping
- **Column-aligned Domain Lists**: Clean formatting with padded spacing for easy readability
- **Note Management**: Add contextual notes to domains (e.g., "marketing team", "alaska project")
- **Interactive Summary Dashboard**: Filterable domain statistics with emoji status indicators (✅🚫❓)
- **Auto-refresh Interface**: 5-second auto-refresh for live monitoring with toggle control
- **Automatic File Creation**: Creates missing whitelist/blacklist files on startup
- **API-driven Architecture**: Modern REST API with instant feedback (no manual save required)
- **Dockerized Deployment**: Complete containerized setup with optimized Squid configuration

## API Endpoints
- `GET /` — Main web interface with domain management and monitoring
- `GET /summary-data` — JSON summary data for filtering and dashboard
- `GET /log` — Recent access log entries (last 50 lines) with embedded tags
- `GET /lists` — Current whitelist/blacklist content as JSON
- `POST /move-domain` — Move domains between whitelist/blacklist/unknown status with notes
- `POST /clear-all-logs` — Clear all categorized access logs
- `GET /static/*` — Static assets (CSS, JS, templates)

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)

### Docker Compose (Recommended)
```bash
git clone https://github.com/coledykstra7/docker-whitelist
cd docker-whitelist
docker-compose up --build
```

**Access Points:**
- Web UI: http://localhost:8080
- Squid Proxy: localhost:3128

### Local Development
```bash
cd src
go mod download
go build -o ../squid-editor .
cd ..
./squid-editor
```

## Architecture

### Backend (Go)
- **Framework**: Gin with middleware for no-cache headers
- **File Management**: Centralized domain list writing with automatic sorting
- **Log Processing**: Smart merging of categorized logs with timestamp ordering
- **Domain Operations**: API-driven CRUD operations with automatic squid reloading
- **Auto-initialization**: Creates required files and directories on startup

### Frontend (Vanilla JavaScript)
- **Real-time Updates**: Live refresh of logs and statistics every 5 seconds
- **Interactive Tables**: Sortable domain lists with action buttons
- **Local Storage**: Persistent note field across sessions
- **Responsive Design**: Grid layout adapting to screen size
- **Color-coded Logs**: Visual distinction between WL/BL/RG entries

### Proxy (Squid)
- **Whitelist-first**: Allows whitelisted domains, blocks blacklisted, logs unknown
- **Categorized Logging**: Separate logs for whitelist, blacklist, and regular traffic
- **Performance Optimized**: Workers, DNS caching, and connection tuning
- **No-cache Headers**: Prevents caching for consistent filtering

## Domain List Format
```
# Automatically formatted with column alignment
example.com          # Work project
google.com           # Search engine
stackoverflow.com    # Development
```

## Configuration Files
```
data/
├── whitelist.txt    # Allowed domains (auto-created)
├── blacklist.txt    # Blocked domains (auto-created)  
├── access-whitelist.log
├── access-blacklist.log
└── access-regular.log
```

## Testing
```bash
cd src
go test -v
```

## Directory Structure
```
├── src/                     # Go application source
│   ├── main.go             # Entry point with file initialization
│   ├── handlers.go         # HTTP request handlers and routing
│   ├── logs.go             # Log processing and merging
│   ├── squid.go            # Squid control and status checking
│   ├── utils.go            # Domain sorting and file operations
│   ├── types.go            # Data structures and constants
│   ├── files.go            # File I/O utilities
│   ├── html.go             # HTML generation for tables
│   └── main_test.go        # Comprehensive unit tests
├── html/                   # Frontend assets
│   ├── template.html       # Main UI template
│   ├── template.js         # Interactive JavaScript
│   └── template.css        # Responsive styling
├── squid/                  # Proxy configuration
│   └── squid.conf          # Optimized whitelist proxy config
├── data/                   # Runtime data (gitignored)
├── docker-compose.yml      # Multi-service orchestration
├── Dockerfile              # Squid container
└── Dockerfile.editor       # Go application container
```

## Key Features Deep Dive

### Smart Domain Sorting
- **Primary**: Sort by note (alphabetically)
- **Secondary**: Sort by reverse domain parts (www.example.com → com.example.www)
- **Consistent**: Applied automatically on all saves

### Real-time Interface
- **No Manual Save**: All changes applied instantly via API
- **Live Updates**: Logs and statistics refresh every 5 seconds
- **Immediate Feedback**: Success/error messages for all operations

### Column Formatting
- **Aligned Notes**: Consistent spacing for easy reading
- **Preserved on Edit**: Notes maintained when moving between lists
- **Backward Compatible**: Handles both old and new formats

## License
MIT
