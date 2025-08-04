package style

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

func Signature(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_signature(msg))
}

func Plain(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_plain(msg))
}

func Bold(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_bold(msg))
}

func Sub(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_subMsg(msg))
}

func Prompt(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Print(_info("\n%s", msg))
}

func InfoLite(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_info("[INFO]"), msg)
}

func Info(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_info("[INFO]"), _info(msg))
}

func WarnLite(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_warn("[WARN]"), msg)
}

func Warn(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_warn("[WARN]"), _warnMsg(msg))
}

func ErrLite(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_error("[ERROR]"), msg)
}

func Err(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_error("[ERROR]"), _errorMsg(msg))
}

func Ok(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_okMsg("[OK]"), msg)
}

func Success(format string, a ...any) {
    msg := fmt.Sprintf(format, a...)
    fmt.Println(_success("[SUCCESS]"), _okMsg(msg))
}
