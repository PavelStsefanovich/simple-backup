package main

import (
	// "bufio"
	"flag"
	"fmt"
	// "github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
	// "log"
	"os"
	"math/rand"
	"regexp"
	"path/filepath"
	"runtime"
	"simple-backup/style"
	"strconv"
	"strings"
	"time"

	// debug
	// "reflect"
	// "simple-backup/helpers"
)


// Limits and Defaults
const (
	BackupDestDirDefault string  	= "smbkp"
	ConfigFileDefault string		= ".smbkp.yaml"
	LimitMinBackupsToKeep int		= 1
	LimitMinFreeSpace string		= "10mb"
	LimitMinFreeSpaceParsed int64	= 10485760
	MinFreeSpacePattern	string		= `^\d+(mb|gb)$`
	Version string					= "0.1.0"
)


//////////////  STRUCTS  //////////////////////////////////////////////////////

// Backup config object
type Config struct {
	BkpDestDir 		string `yaml:"bkp_dest_dir"`
	Schedule   *struct {
		Frequency	string	`yaml:"frequency"`
		DayOfMonth	int		`yaml:"day_of_the_month,omitempty"`
		DyOfWeek	string	`yaml:"day_of_the_week,omitempty"`
		Time      	int		`yaml:"time_of_the_day"`
	} `yaml:"schedule,omitempty"`
	Retention struct {
		BackupsToKeep 		int    `yaml:"backups_to_keep"`
		MinFreeSpace  		string `yaml:"min_free_space"`
		minFreeSpaceParsed	int64	// set implicitly by parsing MinFreeSpace
	} `yaml:"retention"`
	BkpItems []BackupItem `yaml:"bkp_items"`
}


// Object for each entry under 'bkp_items'
type BackupItem struct {
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
	Include     []string `yaml:"include,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
}


// Backup outcome tracking object
type BackupResult struct {
	Item    BackupItem
	Success bool
	Error   error
	Elapsed time.Duration
}


// Main application object
type BackupApp struct {
	configFile		string
	BkpConfig       Config
	bkpDest         string
	bkpDestFullPath	string
	exitOnError     bool
	// logFilePath		string //TODO To be implemented
	nonInteractive  bool
	runOnce			bool
}


//////////////  INIT FUNCTIONS  ///////////////////////////////////////////////

// ENTRY POINT
func main() {
	fmt.Println()

	// Command-line args
	var (
		configFile      = flag.String("config", "", "Path to configuration file.")
		bkpDest         = flag.String("bkp-dest", "", "Backup destination drive or mount. Required if -config is specified.")
		exitOnError     = flag.Bool("exit-on-error", false, "Exit immediately on any copy operation failure.")
		nonInteractive  = flag.Bool("non-interactive", false, "Skip all user prompts.")
		runOnce         = flag.Bool("run-once", true, "Run backup once and exit (ignores schedule).")
		showHelp        = flag.Bool("help", false, "Show help.")
		showVersion     = flag.Bool("version", false, "Show version info.")
	)
	flag.Parse()

	// Show help
	if *showHelp {
		printHelp()
		return
	}

	// Show version
	if *showVersion {
		printVersion()
		return
	}

	// (debug) Show Backup struct
	// helpers.PrintYAMLKeysForType(reflect.TypeOf(BackupApp{}))

	// Initiate main app
	app, err := NewBackupApp(*bkpDest, *configFile, *exitOnError, *nonInteractive, *runOnce)
	if err != nil {
		style.Err("Failed to initialize application: %v", err)
		os.Exit(1)
		// log.Fatalf("Failed to initialize application: %v", err)
		// return
	}

	// Review backup configuration before proceeding
	reviewBackupConfig(app)

	//TODO Validate against the empty bkp_items[] list


	// DELETE (debug) current end
	style.Info("This is the end (currently)")
	fmt.Println(app)
	return

	// Run once
	// if *runOnce || app.BkpConfig.Schedule == nil {
	// 	if err := app.runBackup(); err != nil {
	// 		log.Fatalf("Backup failed: %v", err)
	// 	}
	// 	return
	// }

	// Run scheduled backup
	// app.runScheduledBackup()
}


// PRINT HELP
func printHelp() {
	fmt.Println()
	style.Signature("===============  Simple Backup  ===============")
	fmt.Println()
	style.Plain("Usage:")
	fmt.Println("  smbkp [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	style.Sub("If -bkp-dest is not provided, the app will search for the first drive/mount")
	style.Sub("that contains '" + ConfigFileDefault + "' file in its root directory.")
	fmt.Println()
}


// PRINT VERSION
func printVersion() {
	style.Signature("Simple Backup")
	style.Plain("v%s", Version)
	fmt.Println()
}


// MAIN APP INIT
func NewBackupApp(bkpDest, configFile string, exitOnError, nonInteractive, runOnce bool) (*BackupApp, error) {
	app := &BackupApp{
		BkpConfig:		*NewConfig(), // Set defaults first
		bkpDest:        bkpDest,
		exitOnError:    exitOnError,
		nonInteractive: nonInteractive,
		runOnce:  		runOnce,
	}

	// Case: Backup Destination explicitly specified by user
	if bkpDest != "" {
		style.Plain("Trying to access specified backup destination %q... ", bkpDest)
		_, err := os.Stat(bkpDest)
		if err != nil {
			style.PlainLn("")
			if perr, ok := err.(*os.PathError); ok { // this wrapper code allows to parse the error and enquote file path
				return nil, fmt.Errorf("%q: %v", perr.Path, perr.Err)
			}
			return nil, fmt.Errorf("accessing backup destination: %w", err)
		}
		style.Ok("")
		app.bkpDest = bkpDest
	}

	// Case: Config File explicitly specified by user
	if configFile != "" {
		// Case: Config File explicitly specified by User, but Backup Destination is NOT
		if app.bkpDest == "" {
			return nil, fmt.Errorf("%q is not provided, but it is required when %q is specified", "-bkp-dest", "-config")
		}
		// Case: Both Config File and Backup Destination explicitly specified by user
		style.Plain("Reading specified config file %q... ", configFile)
		if err := app.loadConfig(configFile); err != nil {
			return nil, err
		}
	}

	// Case: Backup Destination is NOT specified
	// (this means that Config File is NOT specified ether)
	if app.bkpDest == "" {
		// Get available drives and mount points
		style.InfoLite("%q is not specified.", "-bkp-dest")
		style.Plain("Retrieving available drives and common mount points... ")
		drives, err := getAvailableDrives()
		if err != nil {
			style.PlainLn("")
			return nil, fmt.Errorf("getting available drives: %w", err)
		}
		style.Ok("")

		// Print found destinations
		for _, drive := range drives {
			style.Sub("  %s", drive)
		}

		// Search for the first destination with default backup config file in it's root
		style.Plain("Searching for %q file in the root of available drives and mount points... ", ConfigFileDefault)
		for _, drive := range drives {
			configFile := filepath.Join(drive, ConfigFileDefault)
			if _, err := os.Stat(configFile); err == nil {
				// Found a backup destination candidate
				style.Ok("")
				style.Plain("Reading config file %q... ", configFile)
				if err := app.loadConfig(configFile); err != nil {
					return nil, err
				}
				app.bkpDest = drive
				break
			}
		}

		if app.bkpDest == "" {
			style.PlainLn("")
			return nil, fmt.Errorf("no backup destination found. Place '.smbkp.yaml' in the root of the destination drive or use the -bkp-dest flag")
		}
	}

	// Case: Backup Destination is explicitly specified by user, but Config File is NOT
	if app.configFile == "" {
		configFile := filepath.Join(app.bkpDest, ConfigFileDefault)
		style.InfoLite("%q is not specified. Assuming default config file in the root of backup destination.", "-config")
		style.Plain("Reading assumed config file %q... ", configFile)
		if err := app.loadConfig(configFile); err != nil {
			return nil, err
		}
	}

	// Creating full backup destination path (bkpDest/bkp_dest_dir/<unique_dir>)
	fullPath, _ := generateUniquePath(app.bkpDest, app.BkpConfig.BkpDestDir)
	app.bkpDestFullPath = fullPath

	return app, nil
}


// INIT CONFIG STRUCT WITH DEFAULT VALUES
func NewConfig() *Config {
	return &Config{
		BkpDestDir: BackupDestDirDefault,
		Retention: struct {
			BackupsToKeep int    `yaml:"backups_to_keep"`
			MinFreeSpace  string `yaml:"min_free_space"`
			minFreeSpaceParsed int64
		}{
			BackupsToKeep: 		LimitMinBackupsToKeep,
			MinFreeSpace:  		LimitMinFreeSpace,
			minFreeSpaceParsed: LimitMinFreeSpaceParsed,
		},
		BkpItems: []BackupItem{},
	}
}


// LOAD MAIN CONFIG FROM FILE
func (app *BackupApp) loadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)

	if err != nil {
		style.PlainLn("")
		if perr, ok := err.(*os.PathError); ok { // this wrapper code allows to parse the error and enquote file path
			return fmt.Errorf("%q: %v", perr.Path, perr.Err)
		}
		return fmt.Errorf("reading config file: %w", err)
	}
	style.Ok("")

	if err := yaml.Unmarshal(data, &app.BkpConfig); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	if err := app.BkpConfig.validate(); err != nil {
		style.PlainLn("")
		return fmt.Errorf("invalid configuration: %w", err)
	}

	app.configFile = configFile
	return nil
}


// VALIDATE MAIN APP CONFIG
func (c *Config) validate() error {
	// Validate backups_to_keep
	if c.Retention.BackupsToKeep < LimitMinBackupsToKeep {
		msg := fmt.Sprintf("%q value increased from '%d' to '%d', which is allowed minimum.", "backups_to_keep", c.Retention.BackupsToKeep, LimitMinBackupsToKeep)
		style.WarnLite(msg)
		c.Retention.BackupsToKeep = LimitMinBackupsToKeep
	}

	// Validate min_free_space format. This will fail if the user explicitly specifies it with an empty string value.
	re := regexp.MustCompile(MinFreeSpacePattern)
	if !re.MatchString(strings.ToLower(c.Retention.MinFreeSpace)) {
		return fmt.Errorf(
			"%q value %q has invalid format. Expected format is a number followed by 'mb' or 'gb' (e.g., '100mb', '10gb')",
			"min_free_space",
			c.Retention.MinFreeSpace,
		)
	}

	// Valiedate min_free_space value
	minFreeSpaceParsed, err := parseDiskSize(c.Retention.MinFreeSpace)
	if err != nil {
		return err
	}

	if minFreeSpaceParsed < LimitMinFreeSpaceParsed {
		msg := fmt.Sprintf("%q value increased from '%s' to '%s', which is allowed minimum.", "min_free_space", c.Retention.MinFreeSpace, LimitMinFreeSpace)
		style.WarnLite(msg)
		c.Retention.MinFreeSpace = LimitMinFreeSpace
		c.Retention.minFreeSpaceParsed = LimitMinFreeSpaceParsed
	}
	c.Retention.minFreeSpaceParsed = minFreeSpaceParsed

	// Future validation for schedule type, etc., can be added here.
	return nil
}


// PARSE DISK SIZE STRING
func parseDiskSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))

	var multiplier int64
	var valueStr string

	switch {
	case strings.HasSuffix(sizeStr, "mb"):
		multiplier = 1024 * 1024
		valueStr = strings.TrimSuffix(sizeStr, "mb")
	case strings.HasSuffix(sizeStr, "gb"):
		multiplier = 1024 * 1024 * 1024
		valueStr = strings.TrimSuffix(sizeStr, "gb")
	default:
		return 0, fmt.Errorf("invalid format: must end with 'mb' or 'gb'")
	}

	num, err := strconv.ParseInt(strings.TrimSpace(valueStr), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number value: %w", err)
	}

	return num * multiplier, nil
}


// PROVIDE OS-SPECIFIC COMMON DRIVES OR MOUNT POINTS
func getAvailableDrives() ([]string, error) {
	var drives []string

	switch runtime.GOOS {
	case "windows":
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			path := string(drive) + ":\\"
			if _, err := os.Stat(path); err == nil {
				drives = append(drives, path)
			}
		}
	case "darwin", "linux":
		// Check common mount points
		mountPoints := []string{"/mnt", "/media", "/Volumes"}
		for _, mountPoint := range mountPoints {
			if entries, err := os.ReadDir(mountPoint); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						fullPath := filepath.Join(mountPoint, entry.Name())
						drives = append(drives, fullPath)
					}
				}
			}
		}
	}

	return drives, nil
}


// CREATE UNIQUE DIRECTORY FOR THE BACKUP RUN
func generateUniquePath(baseDir, subDir string) (string, error) {
	// Join parent directories
	parentDirPath := filepath.Join(baseDir, subDir)

	// Generate a unique directory name
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	uniqueName := fmt.Sprintf("smbkp-%s", string(b))

	// Combine parent directories with unique directory into full path
	fullPath := filepath.Join(parentDirPath, uniqueName)

	return fullPath, nil
}


// REVIEW BACKUP CONFIGURATION BEFORE PROCEEDING
func reviewBackupConfig(app *BackupApp) {
	fmt.Println()
    style.Signature("========  Backup Configuration Review  ========")
	fmt.Println()
    style.Plain("Config file: %s\n", app.configFile)
	style.Plain("Backup destination: %s\n", app.bkpDestFullPath)
	style.Plain("Min free space: %s\n", app.BkpConfig.Retention.MinFreeSpace)
	style.Plain("Backups to keep: %d\n", app.BkpConfig.Retention.BackupsToKeep)
    style.Plain("Run once: %t\n", app.runOnce)    
    style.Plain("Non-interactive: %t\n", app.nonInteractive)
	style.Plain("Exit on error: %t\n", app.exitOnError)
	fmt.Println()

	style.Plain("Items to backup: %d\n", len(app.BkpConfig.BkpItems))
	if len(app.BkpConfig.BkpItems) == 0 {
		style.Warn("No items listed under 'bkp_items' in the config file, nothing to backup. Exiting.")
		fmt.Println()
		os.Exit(0)
	}

    for i, item := range app.BkpConfig.BkpItems {
        fmt.Printf("  [%d] Source: %s\n", i+1, item.Source)
        fmt.Printf("      Destination: %s\n", item.Destination)
        if len(item.Include) > 0 {
            fmt.Printf("      Include: %v\n", strings.Join(item.Include, ", "))
        }
        if len(item.Exclude) > 0 {
            fmt.Printf("      Exclude: %v\n", strings.Join(item.Exclude, ", "))
        }
    }

    if app.nonInteractive {
        return
    }

    style.Prompt("Proceed with backup? (only \"yes\" will be accepted)")
    var response string
    fmt.Scanln(&response)
	fmt.Println()
    response = strings.TrimSpace(strings.ToLower(response))
    if response != "yes" {
        style.WarnLite("Backup cancelled by user.")
		fmt.Println()
        os.Exit(0)
    }
}


//////////////  BACKUP FUNCTIONS  /////////////////////////////////////////////

// func (app *BackupApp) runScheduledBackup() {
// 	c := cron.New()

// 	cronExpr := app.buildCronExpression()
// 	fmt.Printf("Scheduling backup with cron expression: %s\n", cronExpr)

// 	_, err := c.AddFunc(cronExpr, func() {
// 		fmt.Printf("\n=== Scheduled backup started at %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
// 		if err := app.runBackup(); err != nil {
// 			log.Printf("Scheduled backup failed: %v", err)
// 		}
// 	})

// 	if err != nil {
// 		log.Fatalf("Failed to schedule backup: %v", err)
// 	}

// 	c.Start()
// 	fmt.Println("Backup scheduler started. Press Ctrl+C to exit.")

// 	// Keep the program running
// 	select {}
// }

// func (app *BackupApp) buildCronExpression() string {
// 	schedule := app.BkpConfig.Schedule
// 	timeParts := strings.Split(schedule.time, ":")
// 	hour := timeParts[0]
// 	minute := timeParts[1]

// 	switch schedule.frequency {
// 	case "daily":
// 		return fmt.Sprintf("%s %s * * *", minute, hour)
// 	case "weekly":
// 		dayMap := map[string]string{
// 			"Sunday": "0", "Monday": "1", "Tuesday": "2", "Wednesday": "3",
// 			"Thursday": "4", "Friday": "5", "Saturday": "6",
// 		}
// 		dayNum := dayMap[schedule.dayOfWeek]
// 		return fmt.Sprintf("%s %s * * %s", minute, hour, dayNum)
// 	default:
// 		return fmt.Sprintf("%s %s * * *", minute, hour) // Default to daily
// 	}
// }

// func (app *BackupApp) runBackup() error {
// 	startTime := time.Now()

// 	// Create backup directory
// 	timestamp := time.Now().Format("20060102-150405")
// 	app.BkpConfig.bkpDestDir = filepath.Join(app.bkpDest, app.BkpConfig.bkpDestDir, fmt.Sprintf("psbkp-%s", timestamp))

// 	fmt.Printf("\n=== Backup Configuration ===\n")
// 	fmt.Printf("Backup destination: %s\n", app.bkpDest)
// 	fmt.Printf("Backup directory: %s\n", app.BkpConfig.bkpDestDir)
// 	fmt.Printf("Items to backup: %d\n", len(app.BkpConfig.BkpItems))
// 	fmt.Printf("Exit on error: %t\n", app.exitOnError)

// 	if !app.nonInteractive {
// 		fmt.Print("\nProceed with backup? (y/N): ")
// 		reader := bufio.NewReader(os.Stdin)
// 		response, _ := reader.ReadString('\n')
// 		response = strings.TrimSpace(strings.ToLower(response))
// 		if response != "y" && response != "yes" {
// 			fmt.Println("Backup cancelled.")
// 			return nil
// 		}
// 	}

// 	// Create backup directory
// 	if err := os.MkdirAll(app.BkpConfig.bkpDestDir, 0755); err != nil {
// 		return fmt.Errorf("creating backup directory: %w", err)
// 	}

// 	fmt.Printf("\n=== Starting Backup ===\n")

// 	var results []BackupResult
// 	var failedCount int

// 	for i, item := range app.BkpConfig.BkpItems {
// 		fmt.Printf("\n[%d/%d] Backing up: %s\n", i+1, len(app.BkpConfig.BkpItems), item.Source)

// 		itemStart := time.Now()
// 		err := app.backupItem(item)
// 		elapsed := time.Since(itemStart)

// 		result := BackupResult{
// 			Item:    item,
// 			Success: err == nil,
// 			Error:   err,
// 			Elapsed: elapsed,
// 		}
// 		results = append(results, result)

// 		if err != nil {
// 			failedCount++
// 			fmt.Printf("❌ FAILED (%v): %v\n", elapsed, err)

// 			if app.exitOnError {
// 				if !app.nonInteractive {
// 					fmt.Print("Exit due to error? (Y/n): ")
// 					reader := bufio.NewReader(os.Stdin)
// 					response, _ := reader.ReadString('\n')
// 					response = strings.TrimSpace(strings.ToLower(response))
// 					if response != "n" && response != "no" {
// 						return fmt.Errorf("backup stopped due to error: %w", err)
// 					}
// 				} else {
// 					return fmt.Errorf("backup stopped due to error: %w", err)
// 				}
// 			}
// 		} else {
// 			fmt.Printf("✅ SUCCESS (%v)\n", elapsed)
// 		}
// 	}

// 	totalElapsed := time.Since(startTime)

// 	// Print summary
// 	fmt.Printf("\n=== Backup Summary ===\n")
// 	fmt.Printf("Total time: %v\n", totalElapsed)
// 	fmt.Printf("Total items: %d\n", len(results))
// 	fmt.Printf("Successful: %d\n", len(results)-failedCount)
// 	fmt.Printf("Failed: %d\n", failedCount)

// 	fmt.Printf("\n=== Detailed Results ===\n")
// 	for i, result := range results {
// 		status := "✅"
// 		if !result.Success {
// 			status = "❌"
// 		}
// 		fmt.Printf("[%d] %s %s (%v)\n", i+1, status, result.Item.Source, result.Elapsed)
// 		if result.Error != nil {
// 			fmt.Printf("    Error: %v\n", result.Error)
// 		}
// 	}

// 	// Cleanup old backups
// 	if err := app.cleanupOldBackups(); err != nil {
// 		fmt.Printf("Warning: Failed to cleanup old backups: %v\n", err)
// 	}

// 	if failedCount > 0 {
// 		return fmt.Errorf("backup completed with %d failures", failedCount)
// 	}

// 	fmt.Println("\n🎉 Backup completed successfully!")
// 	return nil
// }

// func (app *BackupApp) backupItem(item BackupItem) error {
// 	srcPath := item.Source
// 	destPath := filepath.Join(app.BkpConfig.bkpDestDir, item.Destination)

// 	// Ensure destination directory exists
// 	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
// 		return fmt.Errorf("creating destination directory: %w", err)
// 	}

// 	// Check if source is a file or directory
// 	srcInfo, err := os.Stat(srcPath)
// 	if err != nil {
// 		return fmt.Errorf("accessing source path: %w", err)
// 	}

// 	if srcInfo.IsDir() {
// 		return app.copyDirectory(srcPath, destPath, item.Include, item.Exclude)
// 	} else {
// 		return app.copyFile(srcPath, destPath)
// 	}
// }

// func (app *BackupApp) copyDirectory(src, dest string, include, exclude []string) error {
// 	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		// Calculate relative path
// 		relPath, err := filepath.Rel(src, path)
// 		if err != nil {
// 			return err
// 		}

// 		// Skip root directory
// 		if relPath == "." {
// 			return nil
// 		}

// 		// Check include/exclude patterns
// 		if !app.shouldInclude(relPath, include, exclude) {
// 			if info.IsDir() {
// 				return filepath.SkipDir
// 			}
// 			return nil
// 		}

// 		destPath := filepath.Join(dest, relPath)

// 		if info.IsDir() {
// 			return os.MkdirAll(destPath, info.Mode())
// 		} else {
// 			return app.copyFile(path, destPath)
// 		}
// 	})
// }

// func (app *BackupApp) shouldInclude(path string, include, exclude []string) bool {
// 	// If there are include patterns, check if path matches any
// 	if len(include) > 0 {
// 		included := false
// 		for _, pattern := range include {
// 			if matched, _ := filepath.Match(pattern, path); matched {
// 				included = true
// 				break
// 			}
// 			// Also check if it's a subdirectory of an included directory
// 			if strings.HasPrefix(path, pattern+string(filepath.Separator)) {
// 				included = true
// 				break
// 			}
// 		}
// 		if !included {
// 			return false
// 		}
// 	}

// 	// Check exclude patterns (exclude takes priority)
// 	for _, pattern := range exclude {
// 		if matched, _ := filepath.Match(pattern, path); matched {
// 			return false
// 		}
// 		// Also check if it's a subdirectory of an excluded directory
// 		if strings.HasPrefix(path, pattern+string(filepath.Separator)) {
// 			return false
// 		}
// 	}

// 	return true
// }

// func (app *BackupApp) copyFile(src, dest string) error {
// 	// Ensure destination directory exists
// 	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
// 		return err
// 	}

// 	srcFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer srcFile.Close()

// 	destFile, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
// 	defer destFile.Close()

// 	_, err = destFile.ReadFrom(srcFile)
// 	if err != nil {
// 		return err
// 	}

// 	// Copy file permissions
// 	srcInfo, err := srcFile.Stat()
// 	if err != nil {
// 		return err
// 	}

// 	return os.Chmod(dest, srcInfo.Mode())
// }

// func (app *BackupApp) cleanupOldBackups() error {
// 	backupRoot := filepath.Join(app.bkpDest, app.BkpConfig.bkpDestDir)

// 	entries, err := os.ReadDir(backupRoot)
// 	if err != nil {
// 		return err
// 	}

// 	var backupDirs []os.DirEntry
// 	for _, entry := range entries {
// 		if entry.IsDir() && strings.HasPrefix(entry.Name(), "psbkp-") {
// 			backupDirs = append(backupDirs, entry)
// 		}
// 	}

// 	if len(backupDirs) <= app.BkpConfig.Retention.BackupsToKeep {
// 		return nil
// 	}

// 	// Sort by name (which includes timestamp) and remove oldest
// 	// Note: This is a simplified approach. For production, you might want more sophisticated sorting
// 	toDelete := len(backupDirs) - app.BkpConfig.Retention.BackupsToKeep
// 	for i := 0; i < toDelete; i++ {
// 		dirPath := filepath.Join(backupRoot, backupDirs[i].Name())
// 		fmt.Printf("Removing old backup: %s\n", dirPath)
// 		if err := os.RemoveAll(dirPath); err != nil {
// 			return fmt.Errorf("removing old backup %s: %w", dirPath, err)
// 		}
// 	}

// 	return nil
// }
