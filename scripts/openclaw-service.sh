#!/bin/bash
#
# OpenClaw Linux 服务管理脚本
#
# 用法: ./openclaw-service.sh {install|uninstall|start|stop|restart|reload|status|logs|enable|disable}
#

set -e

# 配置
SERVICE_NAME="openclaw"
SERVICE_FILE="scripts/openclaw.service"
EXEC_PATH="/usr/local/bin/openclaw"
CONFIG_PATH="/etc/openclaw/openclaw.json"
WORK_DIR="/var/lib/openclaw"
USER="openclaw"
GROUP="openclaw"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查 root 权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此操作需要 root 权限"
        log_info "请使用: sudo $0 $1"
        exit 1
    fi
}

# 检查 systemd
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        log_error "未找到 systemctl，此系统可能不使用 systemd"
        exit 1
    fi

    if [[ ! -d /etc/systemd/system ]]; then
        log_error "systemd 服务目录不存在"
        exit 1
    fi
}

# 检查服务是否已安装
is_installed() {
    [[ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]]
}

# 检查服务是否正在运行
is_running() {
    systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null
}

# 安装服务
install_service() {
    check_root
    check_systemd

    log_info "安装 OpenClaw 服务..."

    # 检查可执行文件
    if [[ ! -f "$EXEC_PATH" ]]; then
        log_error "可执行文件不存在: $EXEC_PATH"
        log_info "请先安装 OpenClaw"
        exit 1
    fi

    # 创建用户和组
    if ! getent group "$GROUP" > /dev/null 2>&1; then
        log_info "创建组: $GROUP"
        groupadd -r "$GROUP"
    else
        log_info "组已存在: $GROUP"
    fi

    if ! id "$USER" > /dev/null 2>&1; then
        log_info "创建用户: $USER"
        useradd -r -g "$GROUP" -d "$WORK_DIR" -s /bin/false \
            -c "OpenClaw Service User" "$USER"
    else
        log_info "用户已存在: $USER"
    fi

    # 创建目录
    log_info "创建工作目录..."
    mkdir -p "$WORK_DIR"
    mkdir -p /var/log/openclaw
    mkdir -p /etc/openclaw
    mkdir -p /tmp/openclaw

    # 设置权限
    chown -R "$USER:$GROUP" "$WORK_DIR"
    chown -R "$USER:$GROUP" /var/log/openclaw
    chown -R "$USER:$GROUP" /etc/openclaw

    # 复制服务文件
    if [[ -f "$SERVICE_FILE" ]]; then
        log_info "安装服务文件..."
        cp "$SERVICE_FILE" "/etc/systemd/system/${SERVICE_NAME}.service"
        chmod 644 "/etc/systemd/system/${SERVICE_NAME}.service"
    else
        log_error "服务模板文件不存在: $SERVICE_FILE"
        exit 1
    fi

    # 重新加载 systemd
    log_info "重新加载 systemd..."
    systemctl daemon-reload

    # 启用服务
    log_info "启用开机自启..."
    systemctl enable "$SERVICE_NAME"

    log_info "服务安装完成！"
    log_info "使用 '$0 start' 启动服务"
}

# 卸载服务
uninstall_service() {
    check_root

    log_info "卸载 OpenClaw 服务..."

    # 停止服务
    if is_running; then
        log_info "停止服务..."
        systemctl stop "$SERVICE_NAME" || true
    fi

    # 禁用服务
    if is_installed; then
        log_info "禁用开机自启..."
        systemctl disable "$SERVICE_NAME" || true
    fi

    # 删除服务文件
    if [[ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]]; then
        log_info "删除服务文件..."
        rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    fi

    # 重新加载 systemd
    systemctl daemon-reload

    log_info "服务卸载完成"
    log_warn "保留的数据目录: $WORK_DIR, /var/log/openclaw, /etc/openclaw"
}

# 启动服务
start_service() {
    check_root

    if ! is_installed; then
        log_error "服务未安装"
        log_info "请先运行: $0 install"
        exit 1
    fi

    if is_running; then
        log_warn "服务已经在运行"
        return 0
    fi

    log_info "启动服务..."
    systemctl start "$SERVICE_NAME"

    sleep 1

    if is_running; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败"
        log_info "查看日志: $0 logs"
        exit 1
    fi
}

# 停止服务
stop_service() {
    check_root

    if ! is_running; then
        log_warn "服务未在运行"
        return 0
    fi

    log_info "停止服务..."
    systemctl stop "$SERVICE_NAME"

    sleep 1

    if is_running; then
        log_error "服务停止失败"
        exit 1
    else
        log_info "服务已停止"
    fi
}

# 重启服务
restart_service() {
    check_root

    log_info "重启服务..."
    systemctl restart "$SERVICE_NAME"

    sleep 1

    if is_running; then
        log_info "服务重启成功"
    else
        log_error "服务重启失败"
        log_info "查看日志: $0 logs"
        exit 1
    fi
}

# 重新加载服务（平滑重启）
reload_service() {
    check_root

    if ! is_running; then
        log_warn "服务未在运行，尝试启动..."
        start_service
        return 0
    fi

    log_info "重新加载服务配置..."
    if systemctl reload "$SERVICE_NAME" 2>/dev/null; then
        log_info "服务重新加载成功"
    else
        log_warn "服务不支持 reload，执行重启..."
        restart_service
    fi
}

# 查看服务状态
show_status() {
    if ! is_installed; then
        log_error "服务未安装"
        return 1
    fi

    echo ""
    systemctl status "$SERVICE_NAME" --no-pager || true
    echo ""

    if is_running; then
        log_info "服务状态: 运行中"

        # 显示进程信息
        local pid
        pid=$(systemctl show -p MainPID --value "$SERVICE_NAME")
        if [[ "$pid" != "0" && -n "$pid ]]; then
            echo ""
            echo "进程信息:"
            ps -p "$pid" -o pid,ppid,cmd,%cpu,%mem --no-headers 2>/dev/null || true
        fi
    else
        log_warn "服务状态: 未运行"
    fi
}

# 查看日志
show_logs() {
    local lines=${2:-50}
    local follow=false

    # 解析参数
    if [[ "$2" == "-f" || "$2" == "--follow" ]]; then
        follow=true
        lines=${3:-50}
    fi

    log_info "显示最近 $lines 行日志..."

    if [[ "$follow" == true ]]; then
        echo "按 Ctrl+C 退出日志跟踪"
        echo ""
        journalctl -u "$SERVICE_NAME" -f -n "$lines"
    else
        journalctl -u "$SERVICE_NAME" --no-pager -n "$lines"
    fi
}

# 启用开机自启
enable_service() {
    check_root

    if ! is_installed; then
        log_error "服务未安装"
        exit 1
    fi

    log_info "启用开机自启..."
    systemctl enable "$SERVICE_NAME"
    log_info "开机自启已启用"
}

# 禁用开机自启
disable_service() {
    check_root

    if ! is_installed; then
        log_error "服务未安装"
        exit 1
    fi

    log_info "禁用开机自启..."
    systemctl disable "$SERVICE_NAME"
    log_info "开机自启已禁用"
}

# 清理日志
clear_logs() {
    check_root

    log_info "清理日志..."
    journalctl --rotate -u "$SERVICE_NAME" || true
    journalctl --vacuum-time=1s -u "$SERVICE_NAME" || true
    log_info "日志已清理"
}

# 调试模式 - 显示详细信息
debug_info() {
    echo "================================"
    echo "OpenClaw 服务调试信息"
    echo "================================"
    echo ""

    echo "系统信息:"
    echo "  OS: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"' || echo 'Unknown')"
    echo "  Kernel: $(uname -r)"
    echo "  systemd: $(systemctl --version | head -1)"
    echo ""

    echo "服务配置:"
    echo "  Service Name: $SERVICE_NAME"
    echo "  Exec Path: $EXEC_PATH"
    echo "  Config Path: $CONFIG_PATH"
    echo "  Work Dir: $WORK_DIR"
    echo "  User: $USER"
    echo "  Group: $GROUP"
    echo ""

    echo "文件检查:"
    if [[ -f "$EXEC_PATH" ]]; then
        log_info "可执行文件: 存在"
        ls -la "$EXEC_PATH"
    else
        log_error "可执行文件: 不存在 ($EXEC_PATH)"
    fi

    if [[ -f "$CONFIG_PATH" ]]; then
        log_info "配置文件: 存在"
        ls -la "$CONFIG_PATH"
    else
        log_warn "配置文件: 不存在 ($CONFIG_PATH)"
    fi

    if [[ -d "$WORK_DIR" ]]; then
        log_info "工作目录: 存在"
        ls -la "$WORK_DIR"
    else
        log_warn "工作目录: 不存在 ($WORK_DIR)"
    fi
    echo ""

    echo "服务状态:"
    if is_installed; then
        log_info "服务文件: 已安装"
        ls -la "/etc/systemd/system/${SERVICE_NAME}.service"
    else
        log_warn "服务文件: 未安装"
    fi

    if is_running; then
        log_info "服务状态: 运行中"
    else
        log_warn "服务状态: 未运行"
    fi
    echo ""

    echo "进程信息:"
    ps aux | grep -i openclaw | grep -v grep || echo "无 OpenClaw 进程"
    echo ""

    echo "端口监听:"
    ss -tlnp | grep -i openclaw || echo "无监听端口"
    echo ""

    echo "最近日志 (最后 10 行):"
    journalctl -u "$SERVICE_NAME" --no-pager -n 10 2>/dev/null || echo "无法读取日志"
}

# 显示帮助
show_help() {
    cat << EOF
OpenClaw Linux 服务管理脚本

用法: $0 <命令> [选项]

命令:
    install     安装并配置 systemd 服务
    uninstall   卸载服务（保留数据）
    start       启动服务
    stop        停止服务
    restart     重启服务
    reload      重新加载配置（平滑重启）
    status      查看服务状态
    logs        查看服务日志 [行数] [-f]
    enable      启用开机自启
    disable     禁用开机自启
    clear-logs  清理日志
    debug       显示调试信息
    help        显示此帮助

日志选项:
    -f, --follow    跟踪日志输出（实时）

示例:
    sudo $0 install          # 安装服务
    sudo $0 start            # 启动服务
    $0 status                # 查看状态
    $0 logs -f               # 实时跟踪日志
    $0 logs 100              # 查看最后 100 行
    sudo $0 restart          # 重启服务

EOF
}

# 主函数
main() {
    case "${1:-}" in
        install)
            install_service
            ;;
        uninstall)
            uninstall_service
            ;;
        start)
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        reload)
            reload_service
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$@"
            ;;
        enable)
            enable_service
            ;;
        disable)
            disable_service
            ;;
        clear-logs)
            clear_logs
            ;;
        debug)
            debug_info
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: ${1:-}"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
