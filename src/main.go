package main

import (
	// "bufio"
	// "errors"
	"fmt"
	"github.com/spf13/pflag"
	// "gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	// "regexp"
	"path/filepath"
	"simple-backup/src/style"
	// "strings"
	"time"

	// debug
	// "reflect"
)


// LIMITS AND DEFAULTS
const (
	BackupDestDirDefault string  	= "smbkp"
	ConfigFileDefault string		= ".smbkp.yaml"
	LimitMinBackupsToKeep uint16	= 1
	LimitMinFreeSpace string		= "10mb"
	LimitMinFreeSpaceParsed uint64	= 10485760
	MinFreeSpacePattern	string		= `^\d+(mb|gb)$`
	Prefix string					= "smbkp"
	Version string					= "0.1.0"
)



//////////////  STRUCTS  //////////////////////////////////////////////////////

// BACKUP CONFIG OBJECT
type Config struct {
	BkpDestDir		string `yaml:"bkp_dest_dir"`
	Retention struct {
		BackupsToKeep 		uint16 `yaml:"backups_to_keep"`
		MinFreeSpace  		string `yaml:"min_free_space"`
		minFreeSpaceParsed	uint64	// set implicitly by parsing MinFreeSpace
	} `yaml:"retention"`
	BkpItems []BackupItem `yaml:"bkp_items"`
}


// OBJECT FOR EACH ENTRY UNDER 'BKP_ITEMS'
type BackupItem struct {
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
	Include     []string `yaml:"include,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
}


// BACKUP OUTCOME TRACKING OBJECT
type BackupResult struct {
	Item    BackupItem
	Success bool
	Error   error
	Elapsed time.Duration
}


// MAIN APPLICATION OBJECT
type BackupApp struct {
	configFile		string
	BkpConfig       Config
	bkpDest         string
	bkpDestFullPath	string
	exitOnError     bool
	// logDir		string //TODO To be implemented
	nonInteractive  bool
}



//////////////  INIT FUNCTIONS  ///////////////////////////////////////////////

// ENTRY POINT
func main() {
	// Command-line args
	var (
		// configFile     = pflag.StringP("config", "c", "", "Path to configuration file.")
		// bkpDest        = pflag.StringP("bkp-dest", "b", "", "Backup destination drive or mount. Required if -config is specified.")
		// exitOnError    = pflag.BoolP("exit-on-error", "e", false, "Exit immediately on any copy operation failure.")
		logDir         = pflag.StringP("log-dir", "l", "", "Path to a directory to store log file.")
		// nonInteractive = pflag.BoolP("non-interactive", "n", false, "Skip all user prompts.")
		showHelp       = pflag.BoolP("help", "h", false, "Show help.")
		showVersion    = pflag.BoolP("version", "v", false, "Show version info.")
	)
	pflag.Parse()

	// Set up logging
	if *logDir != "" {
		logStartTime := time.Now()
		logFileName := fmt.Sprintf("smbkp-%s.log", logStartTime.Format("20060102-150405"))
		logFilePath := filepath.Join(*logDir, logFileName)

		if err := os.MkdirAll(*logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %v", err)
			os.Exit(1)			
		}

		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Failed to open log file: %v", err)
			os.Exit(1)
		}
		defer logFile.Close()

		logObj := log.New(logFile, "", log.LstdFlags)
		st := style.New(logObj)
		st.Info("Logging initialized.", style.Log(), style.Bold())

	} else {
		logObj := log.New(io.Discard, "", log.LstdFlags)
		st := style.New(logObj)
		st.Info("Log directory not specified, writing to console only.")
	}

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

	// (debug) Show Backup App object
	// helpers.PrintYAMLKeysForType(reflect.TypeOf(BackupApp{}))
	// printYAMLKeysForType(reflect.TypeOf(BackupApp{}))

	// // Initiate main app
	// app, err := NewBackupApp(*bkpDest, *configFile, *exitOnError, *nonInteractive)
	// if err != nil {
	// 	style.Err("Failed to initialize application: %v", err)
	// 	os.Exit(1)
	// 	// log.Fatalf("Failed to initialize application: %v", err)
	// 	// return
	// }

	// // Review backup configuration before proceeding
	// if err = reviewBackupConfig(app); err != nil {
	// 	style.Err("Review failed: %v", err)
	// 	return
	// }

	// // Run backup
	// if err := app.runBackup(); err != nil {
	// 	style.Err("Backup failed: %v", err)
	// }
}


// PRINT HELP
func printHelp() {
	fmt.Println()
	// style.Signature("===============  Simple Backup  ===============")
	fmt.Println()
	// style.Plain("Usage:")
	fmt.Println("  smbkp [options]")
	fmt.Println()
	fmt.Println("Options:")
	pflag.PrintDefaults()
	fmt.Println()
	// style.Sub("If -bkp-dest is not provided, the app will search for the first drive/mount")
	// style.Sub("that contains '" + ConfigFileDefault + "' file in its root directory.")
	fmt.Println()
}


// PRINT VERSION
func printVersion() {
	// style.Signature("Simple Backup")
	// style.Plain("v%s", Version)
	fmt.Println()
}


// // MAIN APP INIT
// func NewBackupApp(bkpDest, configFile string, exitOnError, nonInteractive bool) (*BackupApp, error) {
// 	app := &BackupApp{
// 		BkpConfig:		*NewConfig(), // Set defaults first
// 		bkpDest:        bkpDest,
// 		exitOnError:    exitOnError,
// 		nonInteractive: nonInteractive,
// 	}

// 	// Case: Backup Destination explicitly specified by user
// 	if bkpDest != "" {
// 		style.Plain("Trying to access specified backup destination %q... ", bkpDest)
// 		_, err := os.Stat(bkpDest)
// 		if err != nil {
// 			style.PlainLn("")
// 			if perr, ok := err.(*os.PathError); ok { // this wrapper code allows to parse the error and enquote file path
// 				return nil, fmt.Errorf("%q: %v", perr.Path, perr.Err)
// 			}
// 			return nil, fmt.Errorf("accessing backup destination: %w", err)
// 		}
// 		style.Ok("")
// 		app.bkpDest = bkpDest
// 	}

// 	// Case: Config File explicitly specified by user
// 	if configFile != "" {
// 		// Case: Config File explicitly specified by User, but Backup Destination is NOT
// 		if app.bkpDest == "" {
// 			return nil, fmt.Errorf("%q is not provided, but it is required when %q is specified", "-bkp-dest", "-config")
// 		}
// 		// Case: Both Config File and Backup Destination explicitly specified by user
// 		style.Plain("Reading specified config file %q... ", configFile)
// 		if err := app.loadConfig(configFile); err != nil {
// 			return nil, err
// 		}
// 	}

// 	// Case: Backup Destination is NOT specified
// 	// (this means that Config File is NOT specified ether)
// 	if app.bkpDest == "" {
// 		// Get available drives and mount points
// 		style.InfoLite("%q is not specified.", "-bkp-dest")
// 		style.Plain("Retrieving available drives and common mount points... ")
// 		drives, err := getAvailableDrives()
// 		if err != nil {
// 			style.PlainLn("")
// 			return nil, fmt.Errorf("getting available drives: %w", err)
// 		}
// 		style.Ok("")

// 		// Print found destinations
// 		for _, drive := range drives {
// 			style.Sub("  %s", drive)
// 		}

// 		// Search for the first destination with default backup config file in it's root
// 		style.Plain("Searching for %q file in the root of available drives and mount points... ", ConfigFileDefault)
// 		for _, drive := range drives {
// 			configFile := filepath.Join(drive, ConfigFileDefault)
// 			if _, err := os.Stat(configFile); err == nil {
// 				// Found a backup destination candidate
// 				style.Ok("")
// 				style.Plain("Reading config file %q... ", configFile)
// 				if err := app.loadConfig(configFile); err != nil {
// 					return nil, err
// 				}
// 				app.bkpDest = drive
// 				break
// 			}
// 		}

// 		if app.bkpDest == "" {
// 			style.PlainLn("")
// 			return nil, fmt.Errorf("no backup destination found. Place '.smbkp.yaml' in the root of the destination drive or use the -bkp-dest flag")
// 		}
// 	}

// 	// Case: Backup Destination is explicitly specified by user, but Config File is NOT
// 	if app.configFile == "" {
// 		configFile := filepath.Join(app.bkpDest, ConfigFileDefault)
// 		style.InfoLite("%q is not specified. Assuming default config file in the root of backup destination.", "-config")
// 		style.Plain("Reading assumed config file %q... ", configFile)
// 		if err := app.loadConfig(configFile); err != nil {
// 			return nil, err
// 		}
// 	}

// 	// Creating full backup destination path (bkpDest/bkp_dest_dir/<unique_dir>) #REVIEW The usnique path can be constructed using the timestamp wnen backup starts
// 	// fullPath, _ := generateUniquePath(app.bkpDest, app.BkpConfig.BkpDestDir)
// 	app.bkpDestFullPath = filepath.Join(app.bkpDest, app.BkpConfig.BkpDestDir)

// 	return app, nil
// }


// // INIT CONFIG STRUCT WITH DEFAULT VALUES
// func NewConfig() *Config {
// 	return &Config{
// 		BkpDestDir: BackupDestDirDefault,
// 		Retention: struct {
// 			BackupsToKeep uint16    `yaml:"backups_to_keep"`
// 			MinFreeSpace  string `yaml:"min_free_space"`
// 			minFreeSpaceParsed uint64
// 		}{
// 			BackupsToKeep: 		LimitMinBackupsToKeep,
// 			MinFreeSpace:  		LimitMinFreeSpace,
// 			minFreeSpaceParsed:	LimitMinFreeSpaceParsed,
// 		},
// 		BkpItems: []BackupItem{},
// 	}
// }


// // LOAD MAIN CONFIG FROM FILE
// func (app *BackupApp) loadConfig(configFile string) error {
// 	data, err := os.ReadFile(configFile)

// 	if err != nil {
// 		style.PlainLn("")
// 		if perr, ok := err.(*os.PathError); ok { // this wrapper code allows to parse the error and enquote file path
// 			return fmt.Errorf("%q: %v", perr.Path, perr.Err)
// 		}
// 		return fmt.Errorf("reading config file: %w", err)
// 	}
// 	style.Ok("")

// 	if err := yaml.Unmarshal(data, &app.BkpConfig); err != nil {
// 		return fmt.Errorf("parsing config file: %w", err)
// 	}

// 	if err := app.BkpConfig.validate(); err != nil {
// 		style.PlainLn("")
// 		return fmt.Errorf("invalid configuration: %w", err)
// 	}

// 	app.configFile = configFile
// 	return nil
// }


// // VALIDATE MAIN APP CONFIG
// func (c *Config) validate() error {
// 	// Validate backups_to_keep
// 	if c.Retention.BackupsToKeep < LimitMinBackupsToKeep {
// 		msg := fmt.Sprintf("%q value increased from '%d' to '%d', which is allowed minimum.", "backups_to_keep", c.Retention.BackupsToKeep, LimitMinBackupsToKeep)
// 		style.WarnLite(msg)
// 		c.Retention.BackupsToKeep = LimitMinBackupsToKeep
// 	}

// 	// Validate min_free_space format. This will fail if the user explicitly specifies it with an empty string value.
// 	re := regexp.MustCompile(MinFreeSpacePattern)
// 	if !re.MatchString(strings.ToLower(c.Retention.MinFreeSpace)) {
// 		return fmt.Errorf(
// 			"%q value %q has invalid format. Expected format is a number followed by 'mb' or 'gb' (e.g., '100mb', '10gb')",
// 			"min_free_space",
// 			c.Retention.MinFreeSpace,
// 		)
// 	}

// 	// Valiedate min_free_space value
// 	minFreeSpaceParsed, err := parseDiskSize(c.Retention.MinFreeSpace)
// 	if err != nil {
// 		return err
// 	}

// 	if minFreeSpaceParsed < LimitMinFreeSpaceParsed {
// 		msg := fmt.Sprintf("%q value increased from '%s' to '%s', which is allowed minimum.", "min_free_space", c.Retention.MinFreeSpace, LimitMinFreeSpace)
// 		style.WarnLite(msg)
// 		c.Retention.MinFreeSpace = LimitMinFreeSpace
// 		c.Retention.minFreeSpaceParsed = LimitMinFreeSpaceParsed
// 	}
// 	c.Retention.minFreeSpaceParsed = minFreeSpaceParsed

// 	// Set destination attribute of each item under bkp_items to item's source leaf, if destination is not specified
// 	for i := range c.BkpItems {
// 		if c.BkpItems[i].Destination == "" {
// 			c.BkpItems[i].Destination = filepath.Base(c.BkpItems[i].Source)
// 		}
// 	}


// 	// Future validation for schedule type, etc., can be added here.
// 	return nil
// }


// // REVIEW BACKUP CONFIGURATION BEFORE PROCEEDING
// func reviewBackupConfig(app *BackupApp) error {
// 	fmt.Println()
//     style.Signature("========  Backup Configuration Review  ========")
//     style.PlainLn("Config file: %s", app.configFile)
// 	style.Plain("Backup destination: ")
// 	style.Signature(app.bkpDestFullPath)


// 	// Validate min_free_space
// 	style.PlainLn("Minimum required free space: %s ", app.BkpConfig.Retention.MinFreeSpace)

// 	availableFreeSpace, availableFreeSpaceFormatted, err := getFreeSpace(app.bkpDest)
// 	if err != nil {
// 		return fmt.Errorf("reading free space: %w", err)
// 	}

// 	style.PlainLn("Available free space: %s", availableFreeSpaceFormatted) // Check space on the root of the backup destination

// 	if availableFreeSpace < app.BkpConfig.Retention.minFreeSpaceParsed {
// 		return fmt.Errorf("available free space (%s) is less than required minimum (%s)", availableFreeSpaceFormatted, app.BkpConfig.Retention.MinFreeSpace)
// 	}

// 	style.PlainLn("Backups to keep: %d", app.BkpConfig.Retention.BackupsToKeep)
//     style.PlainLn("Non-interactive: %t", app.nonInteractive)
// 	style.PlainLn("Exit on error: %t", app.exitOnError)
// 	fmt.Println()

// 	// Validate bkp_items
// 	style.PlainLn("Items to backup: %d", len(app.BkpConfig.BkpItems))
// 	if len(app.BkpConfig.BkpItems) == 0 {
// 		style.Warn("No items listed under 'bkp_items' in the config file, nothing to backup. Exiting.")
// 		fmt.Println()
// 		os.Exit(0)
// 	}

//     for i, item := range app.BkpConfig.BkpItems {
//         style.PlainLn("\n  [%d] Source: %s", i+1, item.Source)
//         style.PlainLn("      Destination: %s", item.Destination)
//         if len(item.Include) > 0 {
//             style.PlainLn("      Include: %v", strings.Join(item.Include, ", "))
//         }
//         if len(item.Exclude) > 0 {
//             style.PlainLn("      Exclude: %v", strings.Join(item.Exclude, ", "))
//         }
//     }

//     if app.nonInteractive {
//         return nil
//     }

//     style.Prompt("Proceed with backup? (only \"yes\" will be accepted)")
//     var response string
//     fmt.Scanln(&response)
// 	fmt.Println()
//     response = strings.TrimSpace(strings.ToLower(response))
//     if response != "yes" {
//         style.WarnLite("Backup cancelled by user.")
// 		fmt.Println()
//         os.Exit(0)
//     }
// 	return nil
// }



// //////////////  BACKUP FUNCTIONS  /////////////////////////////////////////////

// // EXECUTE BACKUP
// func (app *BackupApp) runBackup() error {
// 	startTime := time.Now()
// 	timestamp := startTime.Format("20060102-150405")
// 	// log.Println("[INFO] Backup process started.") //DELETE @PS tryout

// 	style.Signature("Backup stated on: %s", startTime.Format(time.RFC822))

// 	// Create backup directory
// 	app.bkpDestFullPath = filepath.Join(app.bkpDestFullPath, fmt.Sprintf("%s-%s", Prefix, timestamp))
// 	style.Plain("Creating backup directory %q... ", app.bkpDestFullPath)
// 	if err := os.MkdirAll(app.bkpDestFullPath, 0755); err != nil {
// 		style.PlainLn("")
// 		return fmt.Errorf("creating backup directory: %w", err)
// 	}
// 	style.Ok("")

// 	// Copy backup items
// 	var results []BackupResult
// 	var failedCount int

// 	for i, item := range app.BkpConfig.BkpItems {
// 		style.PlainLn("\n[%d/%d] Backing up: %s", i+1, len(app.BkpConfig.BkpItems), item.Source)

// 		totalItems, err := app.countTotalItems(item)
// 		if err != nil {
// 			style.Err("Failed to count items for backup: %v", err)
// 			continue
// 		}

// 		var processedItems int
// 		lastUpdate := -1

// 		progressCb := func() {
// 			processedItems++
// 			if totalItems > 0 {
// 				percentage := int(float64(processedItems) * 100 / float64(totalItems))
// 				if percentage > lastUpdate {
// 					progressBarLength := 50
// 					completed := int(float64(percentage) / 100.0 * float64(progressBarLength))
// 					remaining := progressBarLength - completed
// 					if remaining < 0 {
// 						remaining = 0
// 					}
// 					progressBar := strings.Repeat("■", completed) + strings.Repeat(".", remaining)
// 					style.Plain("\r[%s]", progressBar)
// 					lastUpdate = percentage
// 				}
// 			}
// 		}

// 		itemStart := time.Now()

// 		err = app.backupItem(item, progressCb)
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
// 			if errors.Is(err, os.ErrNotExist) {
// 				style.PlainLn("\n❌ %v", err)
// 			} else {
// 				style.PlainLn("\n❌ (%v): %v", elapsed, err)
// 			}

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
// 			progressBarLength := 50
// 			progressBar := strings.Repeat("■", progressBarLength)
// 			style.Plain("\r[%s] ", progressBar)
// 			style.Ok(" (%s)", result.Elapsed)
// 		}
// 	}

// 	// Cleanup old backups
// 	if err := app.cleanupOldBackups(); err != nil {
// 		style.Warn("Failed to cleanup old backups: %v\n", err)
// 	}

// 	totalElapsed := time.Since(startTime)

// 	// Print summary
// 	style.PlainLn("")
// 	style.Signature("==============  Backup  Summary  ==============")
// 	style.PlainLn("Backup destination: %v", app.bkpDestFullPath)
// 	style.PlainLn("Total time: %v", totalElapsed)
// 	style.PlainLn("Total items: %d", len(results))
// 	style.PlainLn("Successful: %d", len(results)-failedCount)
// 	style.PlainLn("Failed: %d", failedCount)

// 	style.PlainLn("")
// 	style.Signature("Detailed Results")
// 	for i, result := range results {
// 		status := "✅"
// 		if !result.Success {
// 			status = "❌"
// 		}
// 		style.PlainLn("[%d] %s %s (%v)", i+1, status, result.Item.Source, result.Elapsed)
// 		if result.Error != nil {
// 			style.Err("%v", result.Error)
// 		}
// 	}

// 	if failedCount > 0 {
// 		return fmt.Errorf("backup completed with %d failures", failedCount)
// 	}

// 	style.PlainLn("")
// 	style.Success("Backup completed successfully!")
// 	style.PlainLn("")
// 	return nil
// }


// // BACKUP EACH INDIVIDUAL ITEM
// func (app *BackupApp) backupItem(item BackupItem, progressCb func()) error {
// 	srcPath := item.Source
// 	destPath := filepath.Join(app.bkpDestFullPath, item.Destination)

// 	// Check if source is a file or directory
// 	srcInfo, err := os.Stat(srcPath)
// 	if err != nil {
// 		return fmt.Errorf("accessing source path: %w", err)
// 	}

// 	if srcInfo.IsDir() {
// 		if err := os.MkdirAll(destPath, srcInfo.Mode()); err != nil {
// 			return fmt.Errorf("creating destination directory: %w", err)
// 		}
// 		return app.copyDirectory(srcPath, destPath, item.Include, item.Exclude, progressCb)
// 	} else {
// 		return app.copyFile(srcPath, destPath, progressCb)
// 	}
// }


// // COUNT TOTAL NUMBER OF ITEMS TO BACKUP
// func (app *BackupApp) countTotalItems(item BackupItem) (int, error) {
// 	var totalItems int
// 	srcInfo, err := os.Stat(item.Source)
// 	if err != nil {
// 		return 0, err
// 	}

// 	if !srcInfo.IsDir() {
// 		return 1, nil // A single file
// 	}

// 	err = filepath.Walk(item.Source, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		relPath, err := filepath.Rel(item.Source, path)
// 		if err != nil {
// 			return err
// 		}

// 		if relPath == "." {
// 			return nil
// 		}

// 		if !app.shouldInclude(relPath, item.Include, item.Exclude) {
// 			if info.IsDir() {
// 				return filepath.SkipDir
// 			}
// 			return nil
// 		}

// 		totalItems++
// 		return nil
// 	})

// 	return totalItems, err
// }


// // COPY DIRECTORY
// func (app *BackupApp) copyDirectory(src, dest string, include, exclude []string, progressCb func()) error {
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

// 		// If it's a directory, create it
// 		if info.IsDir() {
// 			err := os.MkdirAll(destPath, info.Mode())
// 			if err == nil {
// 				progressCb()
// 			}
// 			return err
// 		}

// 		// Handle symlinks
// 		if info.Mode()&os.ModeSymlink != 0 {
// 			// Check what the symlink points to
// 			stat, err := os.Stat(path) // This follows the symlink
// 			if err != nil {
// 				return err
// 			}
// 			if stat.IsDir() {
// 				// It's a symlink to a directory. Recreate the symlink.
// 				target, err := os.Readlink(path)
// 				if err != nil {
// 					return err
// 				}
// 				return os.Symlink(target, destPath)
// 			}
// 			// It's a symlink to a file, fall through to copyFile
// 		}

// 		// It's a regular file or a symlink to a file
// 		return app.copyFile(path, destPath, progressCb)
// 	})
// }


// // COPY FILE
// func (app *BackupApp) copyFile(src, dest string, progressCb func()) error {
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

// 	progressCb()

// 	// Copy file permissions
// 	srcInfo, err := srcFile.Stat()
// 	if err != nil {
// 		return err
// 	}

// 	return os.Chmod(dest, srcInfo.Mode())
// }


// // EVALUATE INCLUDE/EXCLUDE PATTERNS
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


// // REMOVE OLDEST BACKUP(S)
// func (app *BackupApp) cleanupOldBackups() error {
// 	backupRoot := filepath.Dir(app.bkpDestFullPath)

// 	entries, err := os.ReadDir(backupRoot)
// 	if err != nil {
// 		return err
// 	}

// 	var backupDirs []os.DirEntry
// 	for _, entry := range entries {
// 		if entry.IsDir() && strings.HasPrefix(entry.Name(), fmt.Sprintf("%s-", Prefix)) {
// 			backupDirs = append(backupDirs, entry)
// 		}
// 	}

// 	if len(backupDirs) <= int(app.BkpConfig.Retention.BackupsToKeep) {
// 		return nil
// 	}

// 	// Sort by name (which includes timestamp) and remove oldest
// 	toDelete := len(backupDirs) - int(app.BkpConfig.Retention.BackupsToKeep)

// 	if toDelete > 0 {
// 		style.PlainLn("")
// 		style.Signature("Cleanup")
// 	}

// 	for i := 0; i < toDelete; i++ {
// 		dirPath := filepath.Join(backupRoot, backupDirs[i].Name())
// 		style.Sub("removing old backup: %s", dirPath)
// 		if err := os.RemoveAll(dirPath); err != nil {
// 			return fmt.Errorf("removing old backup %s: %w", dirPath, err)
// 		}
// 	}

// 	return nil
// }
