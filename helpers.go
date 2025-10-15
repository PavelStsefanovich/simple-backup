package main

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "reflect"
    "strings"
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
