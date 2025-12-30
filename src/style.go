package main

import (
    "fmt"
    "github.com/fatih/color"
)

var (
    _signature  = color.RGB(242, 103, 18).SprintFunc()
    _plain      = color.New().SprintFunc()
    _bold       = color.New(color.Bold).SprintFunc()
    _subMsg     = color.RGB(150, 150, 150).SprintFunc()
    _info       = color.New(color.FgCyan).SprintFunc()
    _warn       = color.New(color.FgYellow, color.Bold).SprintFunc()
    _warnMsg    = color.New(color.FgYellow).SprintFunc()
    _error      = color.New(color.FgRed, color.Bold).SprintFunc()
    _errorMsg   = color.New(color.FgRed).SprintFunc()
    _success    = color.New(color.FgGreen, color.Bold).SprintFunc()
    _okMsg      = color.New(color.FgGreen).SprintFunc()
)

// Print message in app signature color
func Signature(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_signature(msg))
}

// Print message without styling and without new line
func Plain(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Print(_plain(msg))
}

// Print message without styling
func PlainLn(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_plain(msg))
}

// Pring message in bold
func Bold(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_bold(msg))
}

// Print message in soft gray
func Sub(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_subMsg(msg))
}

// Print user prompt message
func Prompt(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println()
    fmt.Print(_info(msg))
    fmt.Print(_info(":\n  "))
}

// Print info message with partial styling
func InfoLite(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_info("[INFO]"), msg)
}

// Print info message with full styling
func Info(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_info("[INFO]"), _info(msg))
}

// Print warning message with partial styling
func WarnLite(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_warn("[WARN]"), msg)
}

// Print warning message with full styling
func Warn(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_warn("[WARN]"), _warnMsg(msg))
}

// Print error message with partial styling
func ErrLite(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_error("[ERROR]"), msg)
}

// Print error message with full styling
func Err(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_error("[ERROR]"), _errorMsg(msg))
}

// Print success message with partial styling
func Ok(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_okMsg("[OK]"), msg)
}

// Print success message with full styling
func Success(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_success("[SUCCESS]"), _okMsg(msg))
}
