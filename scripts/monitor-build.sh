#!/bin/bash
#
# GitHub Actions 构建实时监控脚本
# 自动轮询构建状态，完成后主动通知
#

REPO="fichas/openclaw_installer"
POLL_INTERVAL=30  # 轮询间隔（秒）
MAX_WAIT=1800     # 最大等待时间（30分钟）

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')] INFO: $1${NC}"
}

log_success() {
    echo -e "${GREEN}[$(date '+%H:%M:%S')] SUCCESS: $1${NC}"
}

log_error() {
    echo -e "${RED}[$(date '+%H:%M:%S')] ERROR: $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}[$(date '+%H:%M:%S')] WARN: $1${NC}"
}

# 获取最新构建信息
get_latest_run() {
    curl -s "https://api.github.com/repos/$REPO/actions/runs?per_page=1" 2>/dev/null
}

# 解析构建信息
parse_run_info() {
    local json="$1"
    echo "$json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('workflow_runs'):
    run = data['workflow_runs'][0]
    print(f\"RUN_ID={run['id']}\")
    print(f\"RUN_NUMBER={run['run_number']}\")
    print(f\"STATUS={run['status']}\")
    print(f\"CONCLUSION={run.get('conclusion', 'null')}\")
    print(f\"MESSAGE={run['head_commit']['message'][:50]}\")
    print(f\"BRANCH={run['head_branch']}\")
    print(f\"URL={run['html_url']}\")
" 2>/dev/null || echo "PARSE_ERROR=true"
}

# 获取构建日志
get_build_log() {
    local run_id="$1"
    curl -s "https://api.github.com/repos/$REPO/actions/runs/$run_id/logs" 2>/dev/null
}

# 分析失败原因
analyze_failure() {
    local run_id="$1"
    log_error "构建失败，正在分析原因..."

    # 获取失败的 jobs
    local jobs_json=$(curl -s "https://api.github.com/repos/$REPO/actions/runs/$run_id/jobs" 2>/dev/null)

    # 提取失败信息
    echo "$jobs_json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for job in data.get('jobs', []):
    if job.get('conclusion') == 'failure':
        print(f\"\\n失败的 Job: {job['name']}\")
        for step in job.get('steps', []):
            if step.get('conclusion') == 'failure':
                print(f\"  失败的步骤: {step['name']}\")
                if step.get('completed_at'):
                    print(f\"  时间: {step['completed_at']}\")
" 2>/dev/null

    log_info "查看完整日志: https://github.com/$REPO/actions/runs/$run_id"
}

# 主监控循环
main() {
    log_info "开始监控 $REPO 的构建状态..."
    log_info "轮询间隔: ${POLL_INTERVAL}秒"

    local start_time=$(date +%s)
    local last_run_id=""
    local notified_start=false

    while true; do
        # 检查是否超时
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        if [ $elapsed -gt $MAX_WAIT ]; then
            log_error "监控超时（${MAX_WAIT}秒）"
            exit 1
        fi

        # 获取最新构建
        local run_json=$(get_latest_run)

        if [ -z "$run_json" ] || echo "$run_json" | grep -q "PARSE_ERROR"; then
            log_warn "无法获取构建信息，等待重试..."
            sleep $POLL_INTERVAL
            continue
        fi

        # 解析信息
        eval $(parse_run_info "$run_json")

        if [ -z "$RUN_ID" ]; then
            log_warn "解析失败，等待重试..."
            sleep $POLL_INTERVAL
            continue
        fi

        # 新构建开始
        if [ "$RUN_ID" != "$last_run_id" ] && [ "$notified_start" = false ]; then
            log_info "========================================"
            log_info "检测到新构建 #${RUN_NUMBER}"
            log_info "提交: ${MESSAGE}"
            log_info "分支: ${BRANCH}"
            log_info "状态: ${STATUS}"
            log_info "========================================"
            notified_start=true
            last_run_id="$RUN_ID"
        fi

        # 构建完成
        if [ "$STATUS" = "completed" ]; then
            echo ""
            log_info "========================================"
            log_info "构建 #${RUN_NUMBER} 完成!"
            log_info "结论: ${CONCLUSION}"
            log_info "========================================"

            if [ "$CONCLUSION" = "success" ]; then
                log_success "🎉 构建成功！"
                log_info ""
                log_info "下载地址:"
                log_info "  https://github.com/$REPO/releases"
                log_info ""
                log_info "构建产物:"

                # 获取 artifact 列表
                local artifacts=$(curl -s "https://api.github.com/repos/$REPO/actions/runs/$RUN_ID/artifacts" 2>/dev/null | \
                    python3 -c "import sys,json; d=json.load(sys.stdin); [print(f'  - {a[\"name\"]}') for a in d.get('artifacts', [])]" 2>/dev/null)
                if [ -n "$artifacts" ]; then
                    echo "$artifacts"
                else
                    log_info "  (暂无 artifact 信息)"
                fi

                # 尝试发送桌面通知
                if command -v notify-send &> /dev/null; then
                    notify-send "OpenClaw 构建成功" "构建 #${RUN_NUMBER} 已完成" 2>/dev/null || true
                fi

                exit 0
            else
                log_error "❌ 构建失败！"
                analyze_failure "$RUN_ID"

                # 尝试发送桌面通知
                if command -v notify-send &> /dev/null; then
                    notify-send "OpenClaw 构建失败" "构建 #${RUN_NUMBER} 失败，请检查日志" 2>/dev/null || true
                fi

                exit 1
            fi
        fi

        # 显示进度
        local mins=$((elapsed / 60))
        local secs=$((elapsed % 60))
        echo -ne "\r${YELLOW}等待中... ${mins}m ${secs}s | 状态: ${STATUS} | 构建 #${RUN_NUMBER}${NC}"

        sleep $POLL_INTERVAL
    done
}

# 处理中断
trap 'echo -e "\n${YELLOW}监控已停止${NC}"; exit 0' INT TERM

# 运行
main "$@"
