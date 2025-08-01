# Simple Backup

A cross-platform backup application written in Go that supports scheduled and on-demand file/directory backups with include/exclude patterns.

## Features

- **Cross-platform**: Works on Windows, macOS, and Linux
- **Scheduled backups**: Support for daily and weekly schedules
- **Pattern matching**: Include/exclude patterns for fine-grained control
- **Auto-discovery**: Automatically finds backup destinations
- **Retention management**: Automatic cleanup of old backups
- **Interactive mode**: User prompts for confirmation
- **Detailed logging**: Progress tracking and summary reports

## Installation

1. Make sure you have Go 1.21 or later installed
2. Clone or download the source code
3. Run the following commands:

```bash
go mod tidy
go build -o smbkp main.go
```

### For different platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o smbkp.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o smbkp-darwin main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o smbkp-linux main.go
```

## Configuration

### Main Configuration File (config.yaml)

```yaml
bkp_root_dir: PS.Backup

# Comment out 'schedule' key to disable scheduled backups
schedule:
  type: weekly  # daily or weekly
  day_of_the_week: Tuesday  # Only for weekly
  interval: 1  # Every N days/weeks
  time: "21:00"  # 24-hour format

retention:
  backups_to_keep: 4
  min_free_space: 100gb

bkp_items:
  - source: '/home/user/Documents/'
    destination: 'home/user/Documents'
    include:
      - '*.pdf'
      - 'important*'
    exclude:
      - 'temp*'
      - '.cache'
```

### Backup Drive Configuration (.smbkp.yaml)

Place this file in the root of your backup drive or mount point:

```yaml
backup_drive: true
version: "1.0"
description: "Simple Backup Drive Configuration"
```

## Usage

### Command Line Options

- `-bkp-dest string`: Backup destination drive or mount
- `-config string`: Path to configuration file (default: config.yaml)
- `-exit-on-error`: Exit immediately on any copy operation failure
- `-non-interactive`: Skip all user prompts
- `-run-once`: Run backup once and exit (ignores schedule)
- `-help`: Show help message

### Examples

```bash
# Run backup once with auto-discovery
./smbkp -run-once

# Run backup to specific destination
./smbkp -run-once -bkp-dest /mnt/backup

# Run in non-interactive mode
./smbkp -run-once -non-interactive

# Run scheduled backup (daemon mode)
./smbkp

# Use custom config file
./smbkp -config /path/to/my-config.yaml -run-once
```

## How It Works

1. **Configuration Loading**: Reads the YAML configuration file
2. **Destination Discovery**: 
   - If `-bkp-dest` is provided, uses that path
   - Otherwise, searches all available drives/mounts for `.smbkp.yaml`
3. **Backup Execution**:
   - Creates timestamped backup directory: `psbkp-YYYYMMDD-HHMMSS`
   - Processes each backup item with include/exclude patterns
   - Tracks timing and success/failure for each item
4. **Cleanup**: Removes old backups based on retention settings

## Directory Structure

The backup creates the following structure:

```
<backup-drive>/
├── .smbkp.yaml
└── PS.Backup/
    ├── psbkp-20240101-210000/
    │   ├── home/user/Documents/
    │   └── etc/config/
    ├── psbkp-20240108-210000/
    └── psbkp-20240115-210000/
```

## Include/Exclude Patterns

- **Include patterns**: If specified, only matching files/directories are backed up
- **Exclude patterns**: Take priority over include patterns
- **Pattern matching**: Uses Go's `filepath.Match` (supports `*` and `?` wildcards)
- **Directory handling**: Patterns apply to directory names and affect subdirectories

### Pattern Examples

```yaml
include:
  - '*.pdf'        # All PDF files
  - 'Documents'    # Documents directory and contents
  - '.*'          # All hidden files/directories

exclude:
  - 'temp*'       # Anything starting with 'temp'
  - '.cache'      # Specific directory
  - '*.tmp'       # All temporary files
```

## Scheduling

The application supports two scheduling modes:

### Daily Backups
```yaml
schedule:
  type: daily
  interval: 1    # Every day
  time: "02:00"  # 2:00 AM
```

### Weekly Backups
```yaml
schedule:
  type: weekly
  day_of_the_week: Sunday
  interval: 1    # Every week
  time: "21:00"  # 9:00 PM
```

## Error Handling

- **Interactive Mode**: Prompts user on errors if `exit-on-error` is enabled
- **Non-Interactive Mode**: Exits immediately on errors if `exit-on-error` is enabled
- **Continue Mode**: Logs errors and continues with remaining items (default)

## Platform-Specific Notes

### Windows
- Drive letters: `C:\`, `D:\`, etc.
- Path separators: Backslashes
- Example source: `C:\Users\MyUser\`

### macOS/Linux
- Mount points: `/mnt`, `/media`, `/Volumes`
- Path separators: Forward slashes
- Example source: `/home/user/Documents/`

## Troubleshooting

### Common Issues

1. **"No backup destination found"**
   - Ensure `.smbkp.yaml` exists in the root of your backup drive
   - Check drive mounting and permissions

2. **"Permission denied"**
   - Run with appropriate permissions (sudo on Unix systems if needed)
   - Check source and destination directory permissions

3. **"Failed to create backup directory"**
   - Verify backup destination has write permissions
   - Check available disk space

### Debug Tips

- Use `-run-once` for testing without scheduling
- Check file paths are correct for your operating system
- Verify include/exclude patterns work as expected

## License

This project is provided as-is for educational and personal use.