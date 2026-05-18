# Ghost Installation Script for Windows
#
# This script automatically downloads and installs the latest version of Ghost
# from the release server. It downloads the appropriate binary for your Windows
# system architecture (x86_64, ARM64, or i386).
#
# Usage:
#   irm https://install.ghost.build/install.ps1 | iex
#
# With custom parameters:
#   $Version="v1.2.3"; irm https://install.ghost.build/install.ps1 | iex
#   $InstallDir="C:\custom\path"; irm https://install.ghost.build/install.ps1 | iex
#
# Or if saved locally:
#   .\install.ps1 -Version "v1.2.3" -InstallDir "C:\custom\path"
#
# Parameters (all optional):
#   Version           - Specific version to install (e.g., "v1.2.3")
#                       Default: installs the latest version
#
#   InstallDir        - Custom installation directory
#                       Default: $env:LOCALAPPDATA\Programs\Ghost
#
# Requirements:
#   - PowerShell 5.1 or higher
#   - Internet connection for downloading

param(
    [string]$Version = $Version,
    [string]$InstallDir = $InstallDir
)

$ErrorActionPreference = "Stop"

# Force TLS 1.2+ for secure downloads
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# Ensure Unicode characters in the progress bar / status render correctly on
# legacy consoles. Modern Windows Terminal handles UTF-8 by default; conhost
# without this still works for ASCII but may show '?' for block glyphs.
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
} catch {}

# System.Net.Http isn't auto-loaded under Windows PowerShell 5.1 (it lives in
# a separate assembly), so HttpClient lookups fail with "Unable to find type".
# Add-Type is a no-op when the assembly is already loaded (e.g. on PS 7+),
# so it's safe to call unconditionally.
Add-Type -AssemblyName System.Net.Http

# Configuration
$RepoName = "ghost"
$BinaryName = "ghost.exe"
$DownloadBaseUrl = "https://install.ghost.build"

# Unicode glyphs used in status output. Built from [char] codes rather than
# literals so the script source stays pure ASCII and parses correctly under
# PowerShell 5.1 even when read without a UTF-8 BOM (PS 5.1 defaults to
# system ANSI for unmarked files, which corrupts multi-byte UTF-8 glyphs).
$Script:GlyphFullBlock  = [string][char]0x2588  # full block
$Script:GlyphLightShade = [string][char]0x2591  # light shade
$Script:GlyphCheck      = [string][char]0x2713  # check mark

# ===========================================================================
# Status line + color helpers
#
# The happy path renders as a single header line followed by a status line
# that gets overwritten in place ("Downloading [bar] 67%" → "Verifying
# integrity..." → "✓ Installed to ..."). Warnings and errors clear the
# status line first so they don't get jumbled with whatever was there.
# ===========================================================================

$Script:LastStatusLength = 0
$Script:UseColor = $false
$Script:IsErrorTty = -not [Console]::IsErrorRedirected

# Try to enable virtual terminal (ANSI escape) processing on the console.
# Windows Terminal and PowerShell 7+ handle this natively; legacy conhost
# under PowerShell 5.1 needs an explicit SetConsoleMode call. If it fails,
# we just disable color — the in-place status line still works since '\r'
# does not require VT.
function Enable-AnsiSupport {
    if (-not $Script:IsErrorTty) { return $false }
    if ($env:WT_SESSION) { return $true }
    if ($PSVersionTable.PSVersion.Major -ge 7) { return $true }

    try {
        if (-not ('GhostInstall.NativeMethods' -as [type])) {
            Add-Type -Namespace 'GhostInstall' -Name 'NativeMethods' -MemberDefinition @'
                [System.Runtime.InteropServices.DllImport("kernel32.dll", SetLastError = true)]
                public static extern System.IntPtr GetStdHandle(int nStdHandle);
                [System.Runtime.InteropServices.DllImport("kernel32.dll", SetLastError = true)]
                public static extern bool GetConsoleMode(System.IntPtr handle, out uint mode);
                [System.Runtime.InteropServices.DllImport("kernel32.dll", SetLastError = true)]
                public static extern bool SetConsoleMode(System.IntPtr handle, uint mode);
'@
        }
        $STD_ERROR_HANDLE = -12
        $ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
        $handle = [GhostInstall.NativeMethods]::GetStdHandle($STD_ERROR_HANDLE)
        $mode = 0
        if ([GhostInstall.NativeMethods]::GetConsoleMode($handle, [ref]$mode)) {
            if ([GhostInstall.NativeMethods]::SetConsoleMode($handle, $mode -bor $ENABLE_VIRTUAL_TERMINAL_PROCESSING)) {
                return $true
            }
        }
    } catch {}
    return $false
}

$Script:UseColor = Enable-AnsiSupport

function Get-Ansi {
    param([string]$Code)
    if ($Script:UseColor) { return "$([char]27)[${Code}m" }
    return ''
}

# Overwrite the current status line with $Text. Uses '\r' + padding rather
# than ANSI clear-line so the in-place update works even when VT processing
# isn't enabled. $VisibleLength is the rendered width of $Text excluding any
# ANSI escape sequences; callers pass it explicitly when $Text is colored,
# otherwise we default to $Text.Length.
function Update-StatusLine {
    param(
        [Parameter(Mandatory)][string]$Text,
        [int]$VisibleLength = -1
    )
    if (-not $Script:IsErrorTty) {
        [Console]::Error.WriteLine($Text)
        return
    }
    if ($VisibleLength -lt 0) { $VisibleLength = $Text.Length }
    $padding = ''
    if ($VisibleLength -lt $Script:LastStatusLength) {
        $padding = ' ' * ($Script:LastStatusLength - $VisibleLength)
    }
    [Console]::Error.Write("`r$Text$padding")
    $Script:LastStatusLength = $VisibleLength
}

# Terminate the in-place status line so subsequent output starts on a fresh
# line. After this, Update-StatusLine starts a brand-new line.
function Complete-StatusLine {
    if ($Script:IsErrorTty -and $Script:LastStatusLength -gt 0) {
        [Console]::Error.WriteLine()
    }
    $Script:LastStatusLength = 0
}

# Warning/error helpers. They clear any in-place status line first so the
# message doesn't collide with partially-overwritten progress output.
function Write-Warn {
    param([string]$Message)
    if ($Script:IsErrorTty -and $Script:LastStatusLength -gt 0) {
        [Console]::Error.Write("`r" + (' ' * $Script:LastStatusLength) + "`r")
        $Script:LastStatusLength = 0
    }
    $yellow = Get-Ansi '33'
    $reset = Get-Ansi '0'
    [Console]::Error.WriteLine("${yellow}[WARN]${reset} $Message")
}

function Write-ErrorMsg {
    param([string]$Message)
    if ($Script:IsErrorTty -and $Script:LastStatusLength -gt 0) {
        [Console]::Error.Write("`r" + (' ' * $Script:LastStatusLength) + "`r")
        $Script:LastStatusLength = 0
    }
    $red = Get-Ansi '31'
    $reset = Get-Ansi '0'
    [Console]::Error.WriteLine("${red}[ERROR]${reset} $Message")
}

# Render a 24-cell progress bar with a trailing percentage label.
function Format-ProgressBar {
    param([int]$Percent)
    if ($Percent -lt 0) { $Percent = 0 }
    if ($Percent -gt 100) { $Percent = 100 }
    $width = 24
    $filled = [int]($Percent * $width / 100)
    $bar = ($Script:GlyphFullBlock * $filled) + ($Script:GlyphLightShade * ($width - $filled))
    return ('{0} {1,3}%' -f $bar, $Percent)
}

function Show-InstallHeader {
    param([string]$Version, [string]$Platform)
    $cyan = Get-Ansi '36'
    $green = Get-Ansi '32'
    $reset = Get-Ansi '0'
    [Console]::Error.WriteLine("${cyan}Ghost${reset} ${green}${Version}${reset} - ${Platform}")
}

# ===========================================================================
# Platform detection
# ===========================================================================

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x86_64" }
        "ARM64" { return "arm64" }
        "x86"   { return "i386" }
        default {
            Write-ErrorMsg "Unsupported architecture: $arch"
            exit 1
        }
    }
}

function Get-PlatformLabel {
    return "windows_$(Get-Architecture)"
}

function Get-ArchiveName {
    $arch = Get-Architecture
    return "${RepoName}_Windows_${arch}.zip"
}

# ===========================================================================
# Downloads
# ===========================================================================

# Simple download for small files (latest.txt, checksum). Quiet on success;
# emits a warning on each retry.
function Get-SmallFileWithRetry {
    param(
        [Parameter(Mandatory)][string]$Url,
        [string]$OutputFile,
        [string]$Description = 'file'
    )

    $maxRetries = 3
    $retryCount = 0
    $backoffSeconds = 1

    while ($retryCount -le $maxRetries) {
        try {
            if ($OutputFile) {
                Invoke-WebRequest -Uri $Url -OutFile $OutputFile -UseBasicParsing | Out-Null
                return
            }
            return (Invoke-WebRequest -Uri $Url -UseBasicParsing).Content
        }
        catch {
            $retryCount++
            if ($retryCount -le $maxRetries) {
                Write-Warn "$Description fetch failed, retrying ($retryCount/$maxRetries)..."
                Start-Sleep -Seconds $backoffSeconds
                $backoffSeconds *= 2
            }
            else {
                Write-ErrorMsg "Failed to fetch $Description after $($maxRetries + 1) attempts"
                Write-ErrorMsg "URL: $Url"
                Write-ErrorMsg "Error: $_"
                exit 1
            }
        }
    }
}

# Download a file with live progress-bar updates on the status line. Uses
# HttpClient so we can read the response stream in chunks and report bytes
# transferred, which Invoke-WebRequest doesn't expose. Status updates are
# throttled to ~10/sec to keep terminal churn low.
function Get-ArchiveWithProgress {
    param(
        [Parameter(Mandatory)][string]$Url,
        [Parameter(Mandatory)][string]$OutputFile
    )

    $maxRetries = 3
    $retryCount = 0
    $backoffSeconds = 1

    while ($retryCount -le $maxRetries) {
        $client = $null
        $response = $null
        $stream = $null
        $fileStream = $null
        try {
            $client = [System.Net.Http.HttpClient]::new()
            $client.Timeout = [TimeSpan]::FromMinutes(10)

            $response = $client.GetAsync($Url, [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()
            $response.EnsureSuccessStatusCode() | Out-Null

            $total = 0L
            if ($response.Content.Headers.ContentLength) {
                $total = [int64]$response.Content.Headers.ContentLength
            }

            $stream = $response.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
            $fileStream = [System.IO.File]::Create($OutputFile)

            $buffer = New-Object byte[] 65536
            $totalRead = 0L
            $lastUpdate = [DateTime]::MinValue

            while ($true) {
                $read = $stream.Read($buffer, 0, $buffer.Length)
                if ($read -le 0) { break }
                $fileStream.Write($buffer, 0, $read)
                $totalRead += $read

                # In non-TTY mode every Update-StatusLine call produces a
                # fresh line, so per-chunk updates would flood CI logs. Skip
                # them entirely — the surrounding flow already prints one
                # "Downloading..." line in that mode.
                if ($Script:IsErrorTty) {
                    $now = [DateTime]::UtcNow
                    if (($now - $lastUpdate).TotalMilliseconds -ge 100) {
                        if ($total -gt 0) {
                            $pct = [int](($totalRead * 100) / $total)
                            Update-StatusLine "Downloading $(Format-ProgressBar $pct)"
                        } else {
                            $kb = [int]($totalRead / 1024)
                            Update-StatusLine "Downloading ${kb} KB"
                        }
                        $lastUpdate = $now
                    }
                }
            }
            # Final paint so the bar settles at 100% if we never hit the
            # throttle window on the last chunk.
            if ($Script:IsErrorTty -and $total -gt 0) {
                Update-StatusLine "Downloading $(Format-ProgressBar 100)"
            }
            return
        }
        catch {
            $retryCount++
            if ($retryCount -le $maxRetries) {
                Write-Warn "Download failed, retrying ($retryCount/$maxRetries)..."
                Start-Sleep -Seconds $backoffSeconds
                $backoffSeconds *= 2
            }
            else {
                Write-ErrorMsg "Failed to download archive after $($maxRetries + 1) attempts"
                Write-ErrorMsg "URL: $Url"
                Write-ErrorMsg "Error: $_"
                exit 1
            }
        }
        finally {
            if ($fileStream) { $fileStream.Dispose() }
            if ($stream) { $stream.Dispose() }
            if ($response) { $response.Dispose() }
            if ($client) { $client.Dispose() }
        }
    }
}

# ===========================================================================
# Version, install dir, checksum, extraction
# ===========================================================================

function Get-Version {
    if ($Version) { return $Version }

    $content = Get-SmallFileWithRetry -Url "$DownloadBaseUrl/latest.txt" -Description 'latest version'
    $latest = "$content".Trim()
    if ([string]::IsNullOrWhiteSpace($latest)) {
        Write-ErrorMsg "latest.txt file is empty"
        exit 1
    }
    return $latest
}

function Test-InPath {
    param([string]$Directory)
    $target = $Directory.TrimEnd('\').ToLower()
    foreach ($dir in ($env:PATH -split ';')) {
        if ($dir.TrimEnd('\').ToLower() -eq $target) { return $true }
    }
    return $false
}

function Get-InstallDir {
    if ($InstallDir) {
        try {
            if (-not (Test-Path $InstallDir)) {
                New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
            }
            return $InstallDir
        }
        catch {
            Write-ErrorMsg "Cannot create user-specified install directory: $InstallDir"
            Write-ErrorMsg "Error: $_"
            exit 1
        }
    }

    $defaultInstallDir = "$env:LOCALAPPDATA\Programs\Ghost"
    try {
        if (-not (Test-Path $defaultInstallDir)) {
            New-Item -ItemType Directory -Path $defaultInstallDir -Force | Out-Null
        }
        return $defaultInstallDir
    }
    catch {
        Write-ErrorMsg "Cannot create install directory: $defaultInstallDir"
        Write-ErrorMsg "Error: $_"
        Write-ErrorMsg "Please specify a custom installation directory with `$InstallDir parameter"
        exit 1
    }
}

function Test-Checksum {
    param(
        [string]$Version,
        [string]$Filename,
        [string]$TmpDir
    )

    $checksumUrl = "$DownloadBaseUrl/releases/$Version/${Filename}.sha256"
    $checksumFile = Join-Path $TmpDir "${Filename}.sha256"
    Get-SmallFileWithRetry -Url $checksumUrl -OutputFile $checksumFile -Description 'checksum file'

    $expectedChecksum = (Get-Content $checksumFile).Trim()
    $actualChecksum = (Get-FileHash -Path (Join-Path $TmpDir $Filename) -Algorithm SHA256).Hash.ToLower()

    if ($actualChecksum -ne $expectedChecksum) {
        Write-ErrorMsg "Checksum validation failed"
        Write-ErrorMsg "Expected: $expectedChecksum"
        Write-ErrorMsg "Actual: $actualChecksum"
        Write-ErrorMsg "For security reasons, installation has been aborted"
        exit 1
    }
}

function Expand-GhostArchive {
    param(
        [string]$ArchiveName,
        [string]$TmpDir
    )

    Expand-Archive -Path (Join-Path $TmpDir $ArchiveName) -DestinationPath $TmpDir -Force
    $binaryPath = Join-Path $TmpDir $BinaryName
    if (-not (Test-Path $binaryPath)) {
        Write-ErrorMsg "Binary not found in archive"
        exit 1
    }
    return $binaryPath
}

# ===========================================================================
# Install + verification + PATH
# ===========================================================================

# Copy the binary into the install directory. Handles the case where the
# previous binary is locked (e.g. ghost is running) by renaming it aside.
function Install-Binary {
    param(
        [string]$BinaryPath,
        [string]$InstallDir
    )

    # Clean up any .old files left behind by previous in-use replacements.
    Get-ChildItem -Path $InstallDir -Filter "${BinaryName}.old*" -ErrorAction SilentlyContinue |
        ForEach-Object {
            try { Remove-Item $_.FullName -Force -ErrorAction Stop } catch { }
        }

    $targetPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $targetPath) {
        try {
            Remove-Item $targetPath -Force -ErrorAction Stop
        }
        catch {
            $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
            $oldPath = "${targetPath}.old.${timestamp}"
            try {
                Move-Item $targetPath $oldPath -Force -ErrorAction Stop
                Write-Warn "Existing binary is in use, renamed to: $(Split-Path $oldPath -Leaf)"
            }
            catch {
                Write-ErrorMsg "Cannot replace binary at $targetPath"
                Write-ErrorMsg "The binary is locked by another process"
                Write-ErrorMsg ""
                Write-ErrorMsg "To fix this:"
                Write-ErrorMsg "  1. Close all Ghost windows/processes"
                Write-ErrorMsg "  2. Run: Stop-Process -Name ghost -Force"
                Write-ErrorMsg "  3. Try the installation again"
                exit 1
            }
        }
    }
    Copy-Item $BinaryPath $targetPath -Force
    return $targetPath
}

# Verify the installed binary runs. Silent on success so it doesn't disturb
# the clean output flow.
function Test-Installation {
    param([string]$BinaryPath)

    if (-not (Test-Path $BinaryPath)) {
        Write-ErrorMsg "Installation verification failed: Binary not found at $BinaryPath"
        exit 1
    }

    try {
        $null = & $BinaryPath version --bare --version-check=false 2>$null
    }
    catch {
        Write-ErrorMsg "Installation verification failed: Binary exists but is not executable"
        Write-ErrorMsg "Error: $_"
        exit 1
    }
}

function Add-DirToUserPath {
    param([string]$Directory)

    if (Test-InPath $Directory) { return }

    try {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        $newPath = if ($userPath -and $userPath.EndsWith(';')) {
            "${userPath}${Directory}"
        } else {
            "${userPath};${Directory}"
        }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')

        $sessionPath = if ($env:PATH.EndsWith(';')) {
            "${env:PATH}${Directory}"
        } else {
            "${env:PATH};${Directory}"
        }
        $env:PATH = $sessionPath

        [Console]::Error.WriteLine("Added $Directory to your PATH")
    }
    catch {
        Write-Warn "Failed to update PATH automatically: $_"
        Write-Warn "You can add it manually with these commands:"
        Write-Warn "  `$path = [Environment]::GetEnvironmentVariable('Path', 'User')"
        Write-Warn "  [Environment]::SetEnvironmentVariable('Path', `"`$path;$Directory`", 'User')"
    }
}

# Run `ghost init` to drive the post-install configuration flow (login,
# MCP server installation, shell completions). PATH setup is handled inline
# by Add-DirToUserPath so this run's current PowerShell session can
# immediately pick up ghost — something a Go subprocess can't do for its
# parent. We pass --skip-if-configured so re-runs of the installer don't
# re-prompt the user unnecessarily.
#
# `ghost init` needs an interactive terminal for its multi-select prompts.
# When running under `irm | iex` with redirected stdin (CI, scripted runs,
# etc.), fall back to printing a hint instead so the install still succeeds.
function Invoke-GhostInit {
    param([string]$BinaryPath)

    if ([Console]::IsInputRedirected -or -not [Environment]::UserInteractive) {
        [Console]::Error.WriteLine("Run '$BinaryPath init' to finish configuring Ghost.")
        return
    }

    try {
        & $BinaryPath --version-check=false init --skip-if-configured
    }
    catch {
        Write-Warn "ghost init failed: $_"
        Write-Warn "Run '$BinaryPath init' manually to finish configuring Ghost."
    }
}

# ===========================================================================
# Main
# ===========================================================================

function Install-Ghost {
    $version = Get-Version
    $installDir = Get-InstallDir
    $platform = Get-PlatformLabel
    $archiveName = Get-ArchiveName

    Show-InstallHeader -Version $version -Platform $platform

    $tmpDir = Join-Path $env:TEMP "ghost-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        $downloadUrl = "$DownloadBaseUrl/releases/$version/$archiveName"
        $archivePath = Join-Path $tmpDir $archiveName

        Update-StatusLine "Downloading..."
        Get-ArchiveWithProgress -Url $downloadUrl -OutputFile $archivePath

        Update-StatusLine "Verifying integrity..."
        Test-Checksum -Version $version -Filename $archiveName -TmpDir $tmpDir

        Update-StatusLine "Extracting archive..."
        $binaryPath = Expand-GhostArchive -ArchiveName $archiveName -TmpDir $tmpDir

        Update-StatusLine "Installing to $installDir..."
        $targetPath = Install-Binary -BinaryPath $binaryPath -InstallDir $installDir

        Test-Installation -BinaryPath $targetPath

        $green = Get-Ansi '32'
        $reset = Get-Ansi '0'
        $successText = "$($Script:GlyphCheck) Installed to $targetPath"
        Update-StatusLine -Text "${green}$($Script:GlyphCheck)${reset} Installed to $targetPath" -VisibleLength $successText.Length
        Complete-StatusLine

        # Blank line between the in-place section and the post-install steps.
        [Console]::Error.WriteLine()

        Add-DirToUserPath -Directory $installDir
        Invoke-GhostInit -BinaryPath $targetPath
    }
    finally {
        if (Test-Path $tmpDir) {
            Remove-Item $tmpDir -Recurse -Force
        }
    }
}

Install-Ghost
