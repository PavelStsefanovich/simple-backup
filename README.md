# Simple Backup

A cross-platform backup application written in Go that executes on-demand file/directory backups with include/exclude patterns.


## Features

- **Cross-platform**: Works on Windows, macOS, and Linux
- **Pattern matching**: Include/exclude patterns for fine-grained control
- **Auto-discovery**: Automatically finds backup destinations
- **Retention management**: Automatic cleanup of old backups
- **Interactive mode**: User prompts for confirmation
- **Detailed logging**: Progress tracking and summary reports, support for logging to a file


## Build From Source

1. Make sure you have Go 1.21 or later installed and discoverable
2. Clone or download the source code
3. Run the following commands in the project root directory:

```bash
go mod tidy
go build -o simple-backup ./src
```

### Build on Linux for For different platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o simple-backup.exe ./src

# macOS
GOOS=darwin GOARCH=amd64 go build -o simple-backup ./src

# Linux
GOOS=linux GOARCH=amd64 go build -o simple-backup ./src
```


## Configuration

### Main Configuration File (.smbkp.yaml)

A YAML configuration file with the following structure is required for Simple Backup app:

#### Full Template (Linux/MacOS)
```yaml
# `Drive Info` block is optional. If provided, it will be displayed during Backup Review stage,
# (before backup starts) so that the user can visually confirm the destination media.
drive_info:
    name: Backup Drive Name
    description: Backup drive description
		# path: /mnt/backup #TODO @PS Implement using thing entry.

retention:
  # Number of previous backups to keep (min 1)
  backups_to_keep: 3
  # Minimum free space that is required to be available on the destination media (min 10mb)
  # Accepted format: XXmb or XXgb
  min_free_space: 10gb

# Root directory on the destination media, where backups will be stored.
# Each backup will create it's own unique folder under this path.
bkp_dest_dir: SimpleBackups

# List of the items to be backed up. Each item must specify `source` and `destination`,
# where `source` is the path to a file or folder to be backed up,
# and `destination` is path on the destination media relative to `bkp_dest_dir/bkp_unique_folder`.
bkp_items:
  - source: '/home/MyUser/Documents/'
    destination: 'MyUser/files'
    # `Include` is optional. Allows to filter the child items
    # to be included into backup if the `source` is a directory.
    # Supports wildcards '*'. Defaults to all items.
    include:
      - '*.pdf'
      - 'important*'
    # `Exclude` is optional. Allows to filter out the child items
    # that are included from the `source` if it's a directory.
    # Supports wildcards '*'. Takes priority over `include`.
    exclude:
      - 'temp*'
      - '.cache'
```

#### Example of Backup Items config for Windows
```yaml
bkp_items:
  - source: 'C:\Users\MyUser\'
    destination: 'C\users\MyUser'
    include:
      - Documents
      - '.*'
    exclude:
      - 'Documents\My *'
      - .virtualenvs
```

### How It Works
1. **Loading Configuration**:
  + By default, the app looks for the config file named `.smbkp.yaml` in the root of the available drives and known mount points.
    + The first found file is used. The order is not guaranteed.
    + If config file is found, the parent drive/mount will be used as the backup destination media.
    + If config file is not found, the app will exit with error.
  + User can specify the config file explicitly using `-c`/`-config` command line argument.
    + When specified explicitly, the config file does not have to be located in the root
      of the backup destination media, and can have any name.
    + If config file is specified explicitly,
      **the backup destination media must also be specified** using `-b`/`-bkp-dest` command line argument.

2. **Backup Destination Media**:
  + If the backup destination is specified using `-b`/`-bkp-dest` command line argument, uses that path.
  + If config file is specified explicitly, uses that file,
    Otherwise, looks for `.smbkp.yaml` in the root of `bkp-dest`.
    If config file is not found, the app will exit with error.

3. **Backup Execution**:
  + The app validates the provided config and prints the details for user review and confirmation.
    In non-interactive mode (`-n`/`-non-interactive`) it will proceed with backup immediately.
  + The app creates `bkp_dest_dir` directory on the destination media if it does not exist.
    Inside of it, the current run's timestamped backup directory `smbkp-YYYYMMDD-HHMMSS` is created.
  + During backup, processes each backup item with include/exclude patterns.
  + Tracks timing and success/failure for each item.

4. **Cleanup**:
  + If backup completed successfully, the app will delete the oldest timestamped backup directories
    under `bkp_dest_dir`, if the number of directories is greater than `retention.backups_to_keep`.
  + If backup finished with errors, the app will promt the user whether to delete the old backups.
    In non-interactive mode (`-n`/`-non-interactive`) it will skip the deletion.

5. **Logging**:
  + Use `-l/-log-dir` command line argument to to enable logging to file
    and to specify the directory where the timestamped log file will be stored.


## Usage

`simple-backup(.exe) [options]`

### Command Line Options

| Option | Type | Required? | Details |
| ------ | ---- | --------- | ------- |
| `-c`, `-config` | string | no | Explicit path/name of backup configuration file. |
| `-b`, `-bkp-dest` | string | no | Explicit path to backup destination drive or mount. |
| `-l`, `-log-dir` | string | no | Path to a directory to store log file. Also enables logging to file. |
| `-e`, `-exit-on-error` | bool | no | Exit immediately on any copy operation failure. |
| `-n`, `-non-interactive` | bool |no | Skip all user prompts. |
| `-h`, `-help` | bool |no | Show help message. |
| `-v`, `-version` | bool |no | Show version info. |


### Examples

```bash
# Run backup with auto-descovery of the backup config/destination.
./simple-backup

# Run backup to specific destination.
# Requires that the destination media has valid `.smbkp.yaml` file in it's root.
./simple-backup -bkp-dest /mnt/backup

# Run backup with custom backup configuration file.
# Configuration file must specify the destination media path.
./simple-backup -config configs/bkp-config-01.yaml

# Run in non-interactive mode
./simple-backup -non-interactive

# Run with logging to file
./simple-backup -log-dir logs
```

## License

This project is provided as-is for educational and personal use.
