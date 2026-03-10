# Shell Setup Guide

This guide explains how to set up jem in your shell environment.

## Quick Setup

The easiest way to set up jem is to run:

```bash
jem setup
```

This command will:
- Detect your shell
- Add the appropriate initialization code to your shell profile
- Create necessary directories and configuration files

## Manual Setup

If you prefer to manually configure your shell, follow the instructions below for your specific shell.

### Bash

Add the following to your `~/.bashrc` file:

```bash
# jem initialization
eval "$(jem init)"
```

After adding, reload your shell:

```bash
source ~/.bashrc
```

### Zsh

Add the following to your `~/.zshrc` file:

```zsh
# jem initialization
eval "$(jem init)"
```

After adding, reload your shell:

```zsh
source ~/.zshrc
```

### PowerShell

Add the following to your PowerShell profile (run `notepad $PROFILE` on Windows or edit `~/.config/powershell/Microsoft.PowerShell_profile.ps1` on Linux/macOS):

```powershell
# jem initialization
jem init | Invoke-Expression
```

To reload your profile:

```powershell
. $PROFILE
```

**Note:** If you haven't created a profile yet, run:

```powershell
New-Item -Path $PROFILE -Type File -Force
```

### Fish

Fish shell is not directly supported by `jem init` due to syntax differences. You have two options:

#### Option 1: Use Bash Compatibility Mode

Add to `~/.config/fish/config.fish`:

```fish
# jem initialization (using bash compatibility)
if status is-interactive
    bash -c 'jem init fish' | source
end
```

#### Option 2: Manual Configuration

Add the following to `~/.config/fish/config.fish`:

```fish
# jem initialization - manual setup
set -x JAVA_HOME $HOME/.jem/current/java
set -x GRADLE_HOME $HOME/.jem/current/gradle
set -x PATH $JAVA_HOME/bin $PATH
```

**Note:** With manual configuration, you need to run `jem init` before starting a new shell to update symlinks, or manually update `JAVA_HOME` and `GRADLE_HOME` when changing versions.

## How It Works

The `jem init` command:

1. **Updates symlinks** in `~/.jem/current/` based on your `[defaults]` configuration
2. **Outputs shell-specific initialization code** that sets environment variables
3. **Sets `JAVA_HOME` and `GRADLE_HOME`** to point to the default versions
4. **Updates `PATH`** to include the default JDK's bin directory

## Session vs Default Mode

Understanding the difference between session and default modes is important:

### Default Mode (Persistent)

When you use `jem use jdk 21.0.1` (or `jem use jdk 21.0.1 --default`):

- Updates the `[defaults]` section in `~/.jem/config.toml`
- Updates symlinks in `~/.jem/current/`
- Changes persist across shell sessions
- The next time you open a shell (after `jem init`), you'll have the new version

### Session Mode (Temporary)

When you use `jem use jdk 21.0.1 --session`:

- Outputs environment variable exports to stdout
- Does NOT update config or symlinks
- Changes are only for the current shell session
- Must be used with `eval`: `eval "$(jem use jdk 21.0.1 --session)"`

**Example workflow:**

```bash
# Set default JDK (persistent)
jem use jdk 21.0.1

# In a new shell, use a different JDK temporarily
eval "$(jem use jdk 17.0.7 --session)"
java -version  # Shows JDK 17

# Open another shell - it will have JDK 21 (the default)
```

## Verifying Setup

To verify your setup is working:

```bash
# Check jem can find your shell
jem init  # Should output environment setup code

# Check current versions
jem current

# Check environment variables
echo $JAVA_HOME
echo $GRADLE_HOME
echo $PATH

# Verify Java and Gradle work
java -version
gradle --version
```

## Troubleshooting

### "jem: command not found"

Make sure jem is in your PATH. If you installed with `go install`, add `$GOPATH/bin` or `$HOME/go/bin` to your PATH.

### Environment variables not set

1. Make sure you ran `jem setup` or manually added the init code
2. Reload your shell: `source ~/.bashrc` (or equivalent)
3. Check if `jem init` outputs anything: `jem init`

### Symlinks not updating

1. Check the `[defaults]` section in `~/.jem/config.toml`
2. Verify the JDK/Gradle directories exist: `ls ~/.jem/jdks/`
3. Run `jem init` manually and check for errors

### PowerShell execution policy

If you get an execution policy error in PowerShell:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Fish shell not working

Fish is not natively supported. Use the manual configuration option above, or consider switching to Bash/Zsh for full compatibility.

## Uninstalling

To remove jem from your shell:

1. Remove the `eval "$(jem init)"` line from your shell profile
2. Optionally, delete the `~/.jem/` directory
3. Reload your shell
