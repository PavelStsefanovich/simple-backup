package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"simple-backup/style"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

// Backup configuration
type Config struct {
	BkpRootDir string `yaml:"bkp_root_dir"`
	Schedule   *struct {
		Type      string `yaml:"type"`
		DayOfWeek string `yaml:"day_of_the_week,omitempty"`
		Interval  int    `yaml:"interval"`
		Time      string `yaml:"time"`
	} `yaml:"schedule,omitempty"`
	Retention struct {
		BackupsToKeep int    `yaml:"backups_to_keep"`
		MinFreeSpace  string `yaml:"min_free_space"`
	} `yaml:"retention"`
	BkpItems []BackupItem `yaml:"bkp_items"`
}

// Each entry under 'bkp_items'
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

// Main application config
type BackupApp struct {
	config         Config
	bkpDest        string
	exitOnError    bool
	nonInteractive bool
	backupDir      string
}


// ENTRY POINT
func main() {
	// Command-line args
	var (
		bkpDest           = flag.String("bkp-dest", "", "Backup destination drive or mount")
		exitOnError       = flag.Bool("exit-on-error", false, "Exit immediately on any copy operation failure")
		nonInteractive    = flag.Bool("non-interactive", false, "Skip all user prompts")
		configFile        = flag.String("config", "", "Path to configuration file")
		runOnce           = flag.Bool("run-once", true, "Run backup once and exit (ignores schedule)")
		showHelp          = flag.Bool("help", false, "Show help")
		showVersion       = flag.Bool("version", false, "Show version info")
	)
	flag.Parse()

	// Vars
	configFileDefault := ".smbkp.yaml"
	version			  := "v0.1.0"

	// Show help
	if *showHelp {
		printHelp(configFileDefault)
		return
	}

	// Show version
	if *showVersion {
		printVersion(version)
		return
	}

	// Initiate main app
	app, err := NewBackupApp(*bkpDest, *configFile, configFileDefault, *exitOnError, *nonInteractive)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run once
	if *runOnce || app.config.Schedule == nil {
		if err := app.runBackup(); err != nil {
			log.Fatalf("Backup failed: %v", err)
		}
		return
	}

	// Run scheduled backup
	app.runScheduledBackup()
}

func printHelp(configFileDefault string) {
	fmt.Println()
	style.Signature("===============  Simple Backup  ===============")
	fmt.Println()
	style.Plain("Usage:")
	fmt.Println("  smbkp [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	style.Sub("If -bkp-dest is not provided, the app will search for a drive/mount")
	style.Sub("containing '" + configFileDefault + "' in its root directory.")
	fmt.Println()
}

func printVersion(version string) {
	style.Signature("Simple Backup")
	style.Plain(version)
	fmt.Println()
}

func (c *Config) validate() error {
	if c.Retention.BackupsToKeep < 1 {
		return fmt.Errorf("retention.backups_to_keep must be 1 or greater")
	}
	// Future validation for MinFreeSpace format, schedule type, etc., can be added here.
	return nil
}

func NewBackupApp(bkpDest, configFile, configFileDefault string, exitOnError, nonInteractive bool) (*BackupApp, error) {
	app := &BackupApp{
		bkpDest:        bkpDest,
		exitOnError:    exitOnError,
		nonInteractive: nonInteractive,
	}

	if configFile == "" {
		configFile = configFileDefault
	}


	if err := app.loadConfig(configFile); err != nil {
		if configFile != configFileDefault {
			return nil, err
		}
		style.Warn("Default config file '%s' not found in the current directory.", configFileDefault)
	}

	if err := app.findBackupDestination(configFileDefault); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *BackupApp) loadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &app.config); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	// Set defaults
	if app.config.Retention.BackupsToKeep == 0 {
		app.config.Retention.BackupsToKeep = 1
	}
	if app.config.Retention.MinFreeSpace == "" {
		app.config.Retention.MinFreeSpace = "1gb"
	}

	if err := app.config.validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return nil
}

func (app *BackupApp) findBackupDestination(configFileDefault string) error {
	if app.bkpDest != "" {
		// Verify the provided destination
		if _, err := os.Stat(app.bkpDest); os.IsNotExist(err) {
			return fmt.Errorf("backup destination %s does not exist", app.bkpDest)
		}
		return nil
	}

	// Search for backup destination
	drives, err := app.getAvailableDrives()
	if err != nil {
		return fmt.Errorf("getting available drives: %w", err)
	}

	for _, drive := range drives {
		configPath := filepath.Join(drive, configFileDefault)
		if _, err := os.Stat(configPath); err == nil {
			// Found a valid backup destination
			app.bkpDest = drive
			fmt.Printf("Found backup destination with valid config file: %s\n", drive)
			return nil
		}
	}

	return fmt.Errorf("no backup destination found. Place '.smbkp.yaml' in the root of the destination drive or use the -bkp-dest flag")
}

func (app *BackupApp) getAvailableDrives() ([]string, error) {
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
		// Also check root
		drives = append(drives, "/")
	}

	return drives, nil
}

func (app *BackupApp) runScheduledBackup() {
	c := cron.New()

	cronExpr := app.buildCronExpression()
	fmt.Printf("Scheduling backup with cron expression: %s\n", cronExpr)

	_, err := c.AddFunc(cronExpr, func() {
		fmt.Printf("\n=== Scheduled backup started at %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
		if err := app.runBackup(); err != nil {
			log.Printf("Scheduled backup failed: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Failed to schedule backup: %v", err)
	}

	c.Start()
	fmt.Println("Backup scheduler started. Press Ctrl+C to exit.")

	// Keep the program running
	select {}
}

func (app *BackupApp) buildCronExpression() string {
	schedule := app.config.Schedule
	timeParts := strings.Split(schedule.Time, ":")
	hour := timeParts[0]
	minute := timeParts[1]

	switch schedule.Type {
	case "daily":
		if schedule.Interval == 1 {
			return fmt.Sprintf("%s %s * * *", minute, hour)
		}
		// For intervals > 1, we'll use a simple daily check (more complex logic would be needed for true interval support)
		return fmt.Sprintf("%s %s */%d * *", minute, hour, schedule.Interval)
	case "weekly":
		dayMap := map[string]string{
			"Sunday": "0", "Monday": "1", "Tuesday": "2", "Wednesday": "3",
			"Thursday": "4", "Friday": "5", "Saturday": "6",
		}
		dayNum := dayMap[schedule.DayOfWeek]
		return fmt.Sprintf("%s %s * * %s", minute, hour, dayNum)
	default:
		return fmt.Sprintf("%s %s * * *", minute, hour) // Default to daily
	}
}

func (app *BackupApp) runBackup() error {
	startTime := time.Now()

	// Create backup directory
	timestamp := time.Now().Format("20060102-150405")
	app.backupDir = filepath.Join(app.bkpDest, app.config.BkpRootDir, fmt.Sprintf("psbkp-%s", timestamp))

	fmt.Printf("\n=== Backup Configuration ===\n")
	fmt.Printf("Backup destination: %s\n", app.bkpDest)
	fmt.Printf("Backup directory: %s\n", app.backupDir)
	fmt.Printf("Items to backup: %d\n", len(app.config.BkpItems))
	fmt.Printf("Exit on error: %t\n", app.exitOnError)

	if !app.nonInteractive {
		fmt.Print("\nProceed with backup? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Backup cancelled.")
			return nil
		}
	}

	// Create backup directory
	if err := os.MkdirAll(app.backupDir, 0755); err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	fmt.Printf("\n=== Starting Backup ===\n")

	var results []BackupResult
	var failedCount int

	for i, item := range app.config.BkpItems {
		fmt.Printf("\n[%d/%d] Backing up: %s\n", i+1, len(app.config.BkpItems), item.Source)

		itemStart := time.Now()
		err := app.backupItem(item)
		elapsed := time.Since(itemStart)

		result := BackupResult{
			Item:    item,
			Success: err == nil,
			Error:   err,
			Elapsed: elapsed,
		}
		results = append(results, result)

		if err != nil {
			failedCount++
			fmt.Printf("❌ FAILED (%v): %v\n", elapsed, err)

			if app.exitOnError {
				if !app.nonInteractive {
					fmt.Print("Exit due to error? (Y/n): ")
					reader := bufio.NewReader(os.Stdin)
					response, _ := reader.ReadString('\n')
					response = strings.TrimSpace(strings.ToLower(response))
					if response != "n" && response != "no" {
						return fmt.Errorf("backup stopped due to error: %w", err)
					}
				} else {
					return fmt.Errorf("backup stopped due to error: %w", err)
				}
			}
		} else {
			fmt.Printf("✅ SUCCESS (%v)\n", elapsed)
		}
	}

	totalElapsed := time.Since(startTime)

	// Print summary
	fmt.Printf("\n=== Backup Summary ===\n")
	fmt.Printf("Total time: %v\n", totalElapsed)
	fmt.Printf("Total items: %d\n", len(results))
	fmt.Printf("Successful: %d\n", len(results)-failedCount)
	fmt.Printf("Failed: %d\n", failedCount)

	fmt.Printf("\n=== Detailed Results ===\n")
	for i, result := range results {
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		fmt.Printf("[%d] %s %s (%v)\n", i+1, status, result.Item.Source, result.Elapsed)
		if result.Error != nil {
			fmt.Printf("    Error: %v\n", result.Error)
		}
	}

	// Cleanup old backups
	if err := app.cleanupOldBackups(); err != nil {
		fmt.Printf("Warning: Failed to cleanup old backups: %v\n", err)
	}

	if failedCount > 0 {
		return fmt.Errorf("backup completed with %d failures", failedCount)
	}

	fmt.Println("\n🎉 Backup completed successfully!")
	return nil
}

func (app *BackupApp) backupItem(item BackupItem) error {
	srcPath := item.Source
	destPath := filepath.Join(app.backupDir, item.Destination)

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	// Check if source is a file or directory
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("accessing source path: %w", err)
	}

	if srcInfo.IsDir() {
		return app.copyDirectory(srcPath, destPath, item.Include, item.Exclude)
	} else {
		return app.copyFile(srcPath, destPath)
	}
}

func (app *BackupApp) copyDirectory(src, dest string, include, exclude []string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Check include/exclude patterns
		if !app.shouldInclude(relPath, include, exclude) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		} else {
			return app.copyFile(path, destPath)
		}
	})
}

func (app *BackupApp) shouldInclude(path string, include, exclude []string) bool {
	// If there are include patterns, check if path matches any
	if len(include) > 0 {
		included := false
		for _, pattern := range include {
			if matched, _ := filepath.Match(pattern, path); matched {
				included = true
				break
			}
			// Also check if it's a subdirectory of an included directory
			if strings.HasPrefix(path, pattern+string(filepath.Separator)) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check exclude patterns (exclude takes priority)
	for _, pattern := range exclude {
		if matched, _ := filepath.Match(pattern, path); matched {
			return false
		}
		// Also check if it's a subdirectory of an excluded directory
		if strings.HasPrefix(path, pattern+string(filepath.Separator)) {
			return false
		}
	}

	return true
}

func (app *BackupApp) copyFile(src, dest string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(srcFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dest, srcInfo.Mode())
}

func (app *BackupApp) cleanupOldBackups() error {
	backupRoot := filepath.Join(app.bkpDest, app.config.BkpRootDir)

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return err
	}

	var backupDirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "psbkp-") {
			backupDirs = append(backupDirs, entry)
		}
	}

	if len(backupDirs) <= app.config.Retention.BackupsToKeep {
		return nil
	}

	// Sort by name (which includes timestamp) and remove oldest
	// Note: This is a simplified approach. For production, you might want more sophisticated sorting
	toDelete := len(backupDirs) - app.config.Retention.BackupsToKeep
	for i := 0; i < toDelete; i++ {
		dirPath := filepath.Join(backupRoot, backupDirs[i].Name())
		fmt.Printf("Removing old backup: %s\n", dirPath)
		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("removing old backup %s: %w", dirPath, err)
		}
	}

	return nil
}
