# Binary Installation Instructions

> You must have built or downloaded the SimpleBackup executable for your platform prior to running installation script.


## Windows

Run install.ps1 to create a shortcut for the executable in "Start Menu\Programs" in order for it to appear in the apps list in the Start Menu.

Examples
```powershell
# Install executable only (no command-line arguments)
.\install.ps1 -ExecutablePath path\to\simple-backup.exe

# Install for specific destination drive
.\install.ps1 -ExecutablePath path\to\simple-backup.exe -ShortcutName "SimpleBackup-BkpDrive_E" -Arguments "-bkp-dest E:\"

# Install to always run as Admin
.\install.ps1 -ExecutablePath path\to\simple-backup.exe -Admin
```
