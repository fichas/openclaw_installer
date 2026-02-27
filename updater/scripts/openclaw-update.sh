#!/bin/bash
#
# OpenClaw Update Script
# Wrapper script for the OpenClaw updater
# Supports Linux and macOS
#

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${INSTALL_DIR:-/usr/local}"
UPDATER_BIN="${UPDATER_BIN:-${SCRIPT_DIR}/openclaw-updater}"
LOG_DIR="${LOG_DIR:-/var/log/openclaw}"
LOCK_FILE="/var/run/openclaw-update.lock"
CONFIG_FILE="${CONFIG_FILE:-/etc/openclaw/updater.json}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Usage information
usage() {
    cat << EOF
OpenClaw Update Script

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    check           Check for available updates only
    update          Perform update (default)
    rollback        Rollback to previous version
    list-backups    List available backups
    install-cron    Install cron job for automatic updates
    install-systemd Install systemd timer for automatic updates
    status          Show update status

Options:
    -y, --yes               Auto-confirm updates
    -f, --force             Force update even if versions match
    -d, --dry-run           Simulate update without making changes
    -a, --adapter NAME      Update specific adapter (wecom, dingtalk, feishu, all)
    -c, --config PATH       Path to configuration file
    -v, --verbose           Verbose output
    -h, --help              Show this help message

Environment Variables:
    INSTALL_DIR             Installation directory (default: /usr/local)
    UPDATER_BIN             Path to updater binary
    LOG_DIR                 Log directory
    CONFIG_FILE             Configuration file path

Examples:
    $0 check                    # Check for updates
    $0 update --yes             # Update with auto-confirm
    $0 update --adapter=wecom   # Update only wecom adapter
    $0 rollback                 # Rollback to previous version
    $0 install-systemd          # Setup automatic updates via systemd
EOF
}

# Check if running as root (required for system-wide updates)
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_warn "Not running as root. Some operations may fail."
        return 1
    fi
    return 0
}

# Acquire lock to prevent concurrent updates
acquire_lock() {
    if [[ -f "$LOCK_FILE" ]]; then
        local pid
        pid=$(cat "$LOCK_FILE" 2>/dev/null)
        if kill -0 "$pid" 2>/dev/null; then
            log_error "Another update process is running (PID: $pid)"
            exit 1
        else
            log_warn "Removing stale lock file"
            rm -f "$LOCK_FILE"
        fi
    fi

    echo $$ > "$LOCK_FILE"
}

# Release lock
release_lock() {
    rm -f "$LOCK_FILE"
}

# Ensure log directory exists
ensure_log_dir() {
    if [[ ! -d "$LOG_DIR" ]]; then
        mkdir -p "$LOG_DIR" 2>/dev/null || {
            log_warn "Cannot create log directory: $LOG_DIR"
            LOG_DIR="/tmp"
        }
    fi
}

# Check if updater binary exists
check_updater() {
    if [[ ! -x "$UPDATER_BIN" ]]; then
        # Try to find updater in PATH
        if command -v openclaw-updater &> /dev/null; then
            UPDATER_BIN="$(command -v openclaw-updater)"
        else
            log_error "Updater binary not found: $UPDATER_BIN"
            log_info "Please ensure openclaw-updater is installed and executable"
            exit 1
        fi
    fi
}

# Run updater with given arguments
run_updater() {
    local args=("$@")
    local log_file="${LOG_DIR}/update-$(date +%Y%m%d-%H%M%S).log"

    log_info "Running updater with args: ${args[*]}"
    log_info "Log file: $log_file"

    if "$UPDATER_BIN" "${args[@]}" 2>&1 | tee "$log_file"; then
        log_success "Updater completed successfully"
        return 0
    else
        local exit_code=$?
        log_error "Updater failed with exit code: $exit_code"
        log_info "Check log file for details: $log_file"
        return $exit_code
    fi
}

# Install cron job for automatic updates
install_cron() {
    log_info "Installing cron job for automatic updates..."

    local cron_file="/etc/cron.d/openclaw-update"
    local script_path
    script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"

    cat > "$cron_file" << EOF
# OpenClaw automatic update cron job
# Generated on $(date)
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
MAILTO=""

# Run update check daily at 2 AM
0 2 * * * root $script_path update --yes >> /var/log/openclaw/cron.log 2>&1
EOF

    chmod 644 "$cron_file"
    log_success "Cron job installed: $cron_file"
    log_info "Updates will run daily at 2:00 AM"
}

# Remove cron job
remove_cron() {
    log_info "Removing cron job..."
    rm -f /etc/cron.d/openclaw-update
    log_success "Cron job removed"
}

# Install systemd timer for automatic updates
install_systemd() {
    log_info "Installing systemd timer for automatic updates..."

    local script_path
    script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"

    # Create systemd service
    cat > /etc/systemd/system/openclaw-update.service << EOF
[Unit]
Description=OpenClaw Automatic Update
After=network.target

[Service]
Type=oneshot
ExecStart=$script_path update --yes
StandardOutput=journal
StandardError=journal
EOF

    # Create systemd timer
    cat > /etc/systemd/system/openclaw-update.timer << EOF
[Unit]
Description=Run OpenClaw update daily

[Timer]
OnCalendar=daily
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
EOF

    # Reload systemd and enable timer
    systemctl daemon-reload
    systemctl enable openclaw-update.timer
    systemctl start openclaw-update.timer

    log_success "Systemd timer installed and started"
    log_info "Check status with: systemctl status openclaw-update.timer"
}

# Remove systemd timer
remove_systemd() {
    log_info "Removing systemd timer..."
    systemctl stop openclaw-update.timer 2>/dev/null || true
    systemctl disable openclaw-update.timer 2>/dev/null || true
    rm -f /etc/systemd/system/openclaw-update.service
    rm -f /etc/systemd/system/openclaw-update.timer
    systemctl daemon-reload
    log_success "Systemd timer removed"
}

# Show update status
show_status() {
    log_info "OpenClaw Update Status"
    echo "========================"

    # Check if updater exists
    if [[ -x "$UPDATER_BIN" ]]; then
        log_success "Updater binary: $UPDATER_BIN"
    else
        log_error "Updater binary not found: $UPDATER_BIN"
    fi

    # Check installation directory
    if [[ -d "$INSTALL_DIR" ]]; then
        log_success "Install directory: $INSTALL_DIR"
    else
        log_warn "Install directory not found: $INSTALL_DIR"
    fi

    # Check log directory
    if [[ -d "$LOG_DIR" ]]; then
        log_success "Log directory: $LOG_DIR"
    else
        log_warn "Log directory not found: $LOG_DIR"
    fi

    # Check configuration
    if [[ -f "$CONFIG_FILE" ]]; then
        log_success "Configuration file: $CONFIG_FILE"
    else
        log_warn "Configuration file not found: $CONFIG_FILE"
    fi

    # Check cron job
    if [[ -f /etc/cron.d/openclaw-update ]]; then
        log_info "Cron job: installed"
    else
        log_info "Cron job: not installed"
    fi

    # Check systemd timer
    if systemctl is-enabled openclaw-update.timer &>/dev/null; then
        log_info "Systemd timer: installed and enabled"
        systemctl status openclaw-update.timer --no-pager 2>/dev/null || true
    else
        log_info "Systemd timer: not installed"
    fi

    # Show recent logs
    echo ""
    log_info "Recent update logs:"
    if [[ -d "$LOG_DIR" ]]; then
        ls -lt "$LOG_DIR" | head -5
    fi
}

# Cleanup old log files (keep last 30 days)
cleanup_logs() {
    log_info "Cleaning up old log files..."
    if [[ -d "$LOG_DIR" ]]; then
        find "$LOG_DIR" -name "update-*.log" -type f -mtime +30 -delete
        log_success "Old log files cleaned up"
    fi
}

# Main function
main() {
    local command="update"
    local updater_args=()

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            check|update|rollback|list-backups|install-cron|install-systemd|status)
                command="$1"
                shift
                ;;
            -y|--yes)
                updater_args+=("-yes")
                shift
                ;;
            -f|--force)
                updater_args+=("-force")
                shift
                ;;
            -d|--dry-run)
                updater_args+=("-dry-run")
                shift
                ;;
            -a|--adapter)
                updater_args+=("-adapter" "$2")
                shift 2
                ;;
            -c|--config)
                CONFIG_FILE="$2"
                updater_args+=("-config" "$2")
                shift 2
                ;;
            -v|--verbose)
                updater_args+=("-v")
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # Execute command
    case $command in
        check)
            check_updater
            ensure_log_dir
            acquire_lock
            trap release_lock EXIT
            run_updater -check "${updater_args[@]}"
            ;;
        update)
            check_updater
            ensure_log_dir
            acquire_lock
            trap release_lock EXIT
            cleanup_logs
            run_updater "${updater_args[@]}"
            ;;
        rollback)
            check_updater
            ensure_log_dir
            acquire_lock
            trap release_lock EXIT
            run_updater -rollback "${updater_args[@]}"
            ;;
        list-backups)
            check_updater
            run_updater -list-backups "${updater_args[@]}"
            ;;
        install-cron)
            check_root
            install_cron
            ;;
        install-systemd)
            check_root
            install_systemd
            ;;
        status)
            show_status
            ;;
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
