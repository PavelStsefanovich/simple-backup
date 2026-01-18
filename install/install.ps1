param (
    [string] $ExecutablePath = $(throw "Required parameter not specified: '-ExecutablePath'"),
    [string] $ShortcutName = "SimpleBackup",
    [string] $Arguments
)

$ErrorActionPreference = "Stop"

$ScriptDir = $PSScriptRoot
$ExecutablePath = (Resolve-Path $ExecutablePath).Path
$StartMenuPath = Join-Path $env:USERPROFILE "Start Menu\Programs"
$ShortcutPath = Join-Path $StartMenuPath "$ShortcutName`.lnk"

# Icon
$IconPath = Join-Path $ScriptDir 'simple-backup.ico'
if (! (Test-Path $IconPath)) {
    $IconPath = $null
}

if ( $IconPath ) {
    $IconPath = (cp $IconPath (Split-Path $ExecutablePath) -PassThru).FullName
}

# Create Shortcut
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($ShortcutPath)
$Shortcut.TargetPath = $ExecutablePath
if ($Arguments) { $Shortcut.Arguments = $Arguments }
if ($IconPath) { $Shortcut.IconLocation = $IconPath }
$Shortcut.Save()

Write-Host "Created shortcut `"$ShortcutPath`"" -ForegroundColor Green
