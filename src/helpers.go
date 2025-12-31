package main

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "os"
    "path/filepath"
    "reflect"
    "runtime"
	"strconv"
    "strings"
)

const (
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

// getYAMLKeysRecursively inspects a Go type and returns a nested map
// representing the YAML keys and subkeys, with empty placeholders for values.
func getYAMLKeysRecursively(t reflect.Type) (interface{}, error) {
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }

    switch t.Kind() {
    case reflect.Struct:
        result := make(map[string]interface{})
        for i := 0; i < t.NumField(); i++ {
            field := t.Field(i)
            tag := field.Tag.Get("yaml")
            var keyName string
            if tag != "" {
                keyName = strings.Split(tag, ",")[0]
            } else {
                keyName = field.Name
            }
            if keyName == "-" {
                continue
            }
            subKeys, err := getYAMLKeysRecursively(field.Type)
            if err != nil {
                return nil, err
            }
            result[keyName] = subKeys
        }
        return result, nil

    case reflect.Slice:
        elemType := t.Elem()
        elem, err := getYAMLKeysRecursively(elemType)
        if err != nil {
            return nil, err
        }
        return []interface{}{elem}, nil

    case reflect.Map:
        keyType := t.Key()
        valType := t.Elem()
        mapKey, err := getYAMLKeysRecursively(keyType)
        if err != nil {
            return nil, err
        }
        mapVal, err := getYAMLKeysRecursively(valType)
        if err != nil {
            return nil, err
        }
        result := make(map[interface{}]interface{})
        result[mapKey] = mapVal
        return result, nil

    case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
        reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
        reflect.Bool, reflect.Float32, reflect.Float64:
        return "", nil

    default:
        return "", fmt.Errorf("unsupported type: %s", t.Kind())
    }
}


// Print YAML structure for debug purposes
func printYAMLKeysForType(t reflect.Type) {
    keyStructure, err := getYAMLKeysRecursively(t)
    if err != nil {
        fmt.Printf("Error generating YAML keys: %v\n", err)
        return
    }
    yamlData, err := yaml.Marshal(keyStructure)
    if err != nil {
        fmt.Printf("Error marshaling to YAML: %v\n", err)
        return
    }
    fmt.Println(string(yamlData))
}


// formatBytes converts a size in bytes to a human-readable string in MB or GB.
func formatBytes(bytes uint64) string {
	if bytes < GB {
		mb := bytes / MB
		return fmt.Sprintf("%dmb", mb)
	}

	gb := float64(bytes) / float64(GB)
	// Format with one decimal place and replace '.' with ','
	return strings.Replace(fmt.Sprintf("%.1fgb", gb), ".", ",", 1)
}


// Parse disk size string that is formatted for human readability
func parseDiskSize(sizeStr string) (uint64, error) {
	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))

	var multiplier uint64
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

	return uint64(num) * multiplier, nil
}


// Provide os-specific common drives or mount points
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
