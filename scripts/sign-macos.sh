#!/bin/bash
#
# macOS 应用签名和公证脚本
#
# 用法: ./scripts/sign-macos.sh [选项]
#   -b, --bundle-id <id>      应用 Bundle ID (默认: io.openclaw.updater)
#   -a, --app-path <path>     应用路径 (默认: dist/OpenClaw.app)
#   -c, --cert <name>         开发者证书名称
#   -u, --username <email>    Apple ID 邮箱
#   -p, --password <password> App 专用密码
#   -t, --team-id <id>        团队 ID
#   -n, --notarize            执行公证
#   -s, --staple              将票证附加到应用
#   -h, --help                显示帮助
#

set -e

# 默认配置
BUNDLE_ID="io.openclaw.updater"
APP_PATH="dist/OpenClaw.app"
ENTITLEMENTS="build/entitlements.plist"
CERT_NAME=""
APPLE_ID=""
APP_PASSWORD=""
TEAM_ID=""
DO_NOTARIZE=false
DO_STAPLE=false
VERBOSE=false

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 显示帮助
show_help() {
    head -n 15 "$0" | tail -n 14
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--bundle-id)
            BUNDLE_ID="$2"
            shift 2
            ;;
        -a|--app-path)
            APP_PATH="$2"
            shift 2
            ;;
        -c|--cert)
            CERT_NAME="$2"
            shift 2
            ;;
        -u|--username)
            APPLE_ID="$2"
            shift 2
            ;;
        -p|--password)
            APP_PASSWORD="$2"
            shift 2
            ;;
        -t|--team-id)
            TEAM_ID="$2"
            shift 2
            ;;
        -n|--notarize)
            DO_NOTARIZE=true
            shift
            ;;
        -s|--staple)
            DO_STAPLE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查是否在 macOS 上运行
check_macos() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "此脚本必须在 macOS 上运行"
        exit 1
    fi
}

# 检查必要工具
check_tools() {
    local tools=("codesign" "security" "xcrun" "spctl")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "未找到必要工具: $tool"
            exit 1
        fi
    done
}

# 检查应用路径
check_app() {
    if [[ ! -d "$APP_PATH" ]]; then
        log_error "应用不存在: $APP_PATH"
        exit 1
    fi

    if [[ ! -f "$APP_PATH/Contents/Info.plist" ]]; then
        log_error "无效的 .app 包: $APP_PATH"
        exit 1
    fi
}

# 检查证书
check_certificate() {
    if [[ -z "$CERT_NAME" ]]; then
        log_info "可用证书列表:"
        security find-identity -v -p codesigning | grep "Developer ID Application"
        log_error "请使用 -c 参数指定证书名称"
        exit 1
    fi

    # 验证证书是否存在
    if ! security find-identity -v -p codesigning | grep -q "$CERT_NAME"; then
        log_error "证书未找到: $CERT_NAME"
        log_info "可用证书:"
        security find-identity -v -p codesigning
        exit 1
    fi
}

# 检查 entitlements 文件
check_entitlements() {
    if [[ ! -f "$ENTITLEMENTS" ]]; then
        log_warn "Entitlements 文件不存在: $ENTITLEMENTS"
        log_info "将使用默认 entitlements"
        create_default_entitlements
    fi
}

# 创建默认 entitlements
create_default_entitlements() {
    mkdir -p "$(dirname "$ENTITLEMENTS")"
    cat > "$ENTITLEMENTS" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.network.server</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
</dict>
</plist>
EOF
    log_info "已创建默认 entitlements 文件"
}

# 清理旧的签名
cleanup_signatures() {
    log_info "清理旧的签名..."
    find "$APP_PATH" -name "*.cstemp" -delete 2>/dev/null || true
    find "$APP_PATH" -name "CodeResources" -delete 2>/dev/null || true
}

# 签名单个文件
sign_file() {
    local file="$1"
    local opts="--force --timestamp --options=runtime"

    if [[ "$VERBOSE" == true ]]; then
        opts="$opts --verbose"
    fi

    if ! codesign $opts --sign "$CERT_NAME" --entitlements "$ENTITLEMENTS" "$file" 2>&1; then
        log_error "签名失败: $file"
        return 1
    fi
}

# 签名框架
sign_framework() {
    local framework="$1"
    log_info "签名框架: $(basename "$framework")"

    # 签名框架内的所有二进制文件
    find "$framework" -type f -perm +111 | while read -r binary; do
        sign_file "$binary" || true
    done

    # 签名框架本身
    sign_file "$framework"
}

# 签名应用包
sign_app() {
    log_info "开始签名应用..."
    log_info "Bundle ID: $BUNDLE_ID"
    log_info "证书: $CERT_NAME"
    log_info "Entitlements: $ENTITLEMENTS"

    # 1. 签名所有动态库
    log_info "签名动态库..."
    find "$APP_PATH" -name "*.dylib" -type f | while read -r lib; do
        sign_file "$lib" || log_warn "无法签名: $lib"
    done

    # 2. 签名所有框架
    log_info "签名框架..."
    find "$APP_PATH" -name "*.framework" -type d | while read -r framework; do
        sign_framework "$framework"
    done

    # 3. 签名所有辅助应用和插件
    log_info "签名辅助应用和插件..."
    find "$APP_PATH" -path "*/Contents/MacOS/*" -type f -perm +111 | while read -r binary; do
        sign_file "$binary" || log_warn "无法签名: $binary"
    done

    find "$APP_PATH" -name "*.appex" -type d | while read -r appex; do
        sign_file "$appex" || log_warn "无法签名: $appex"
    done

    # 4. 签名主应用包
    log_info "签名主应用包..."
    sign_file "$APP_PATH"

    log_info "签名完成"
}

# 验证签名
verify_signature() {
    log_info "验证签名..."

    # 检查签名
    if ! codesign --verify --deep --strict "$APP_PATH" 2>&1; then
        log_error "签名验证失败"
        return 1
    fi

    # 显示签名信息
    log_info "签名信息:"
    codesign -d --deep -vv "$APP_PATH" 2>&1 | head -20

    # 检查 Gatekeeper
    log_info "检查 Gatekeeper 兼容性..."
    if spctl --assess --type exec "$APP_PATH" 2>&1; then
        log_info "Gatekeeper 检查通过"
    else
        log_warn "Gatekeeper 检查失败（这在未公证前是正常的）"
    fi

    return 0
}

# 创建 DMG 包（用于公证）
create_dmg() {
    local dmg_path="${APP_PATH%.app}.dmg"
    log_info "创建 DMG: $dmg_path"

    # 创建临时挂载点
    local temp_dir=$(mktemp -d)
    local volume_name="OpenClaw"

    # 计算大小
    local app_size=$(du -sm "$APP_PATH" | cut -f1)
    local dmg_size=$((app_size + 50))

    # 创建 DMG
    hdiutil create -size "${dmg_size}m" -volname "$volume_name" -srcfolder "$APP_PATH" -ov -format UDZO "$dmg_path" 2>&1

    log_info "DMG 创建完成: $dmg_path"
    echo "$dmg_path"
}

# 提交公证
notarize_app() {
    if [[ "$DO_NOTARIZE" != true ]]; then
        log_info "跳过公证（使用 -n 参数启用）"
        return 0
    fi

    if [[ -z "$APPLE_ID" || -z "$APP_PASSWORD" ]]; then
        log_error "公证需要 Apple ID 和 App 专用密码"
        log_info "使用 -u <email> -p <password> 参数"
        exit 1
    fi

    log_info "开始公证流程..."

    # 创建 DMG 进行公证
    local dmg_path=$(create_dmg)

    # 提交公证
    log_info "提交到 Apple Notary Service..."
    local output
    if [[ -n "$TEAM_ID" ]]; then
        output=$(xcrun notarytool submit "$dmg_path" \
            --apple-id "$APPLE_ID" \
            --password "$APP_PASSWORD" \
            --team-id "$TEAM_ID" \
            --wait 2>&1)
    else
        output=$(xcrun notarytool submit "$dmg_path" \
            --apple-id "$APPLE_ID" \
            --password "$APP_PASSWORD" \
            --wait 2>&1)
    fi

    log_info "公证结果:"
    echo "$output"

    # 检查是否成功
    if echo "$output" | grep -q "status: Accepted"; then
        log_info "公证成功！"

        # 获取 submission ID
        local submission_id=$(echo "$output" | grep "id:" | head -1 | awk '{print $2}')

        if [[ "$DO_STAPLE" == true ]]; then
            staple_ticket "$dmg_path"
        fi

        # 清理 DMG
        rm -f "$dmg_path"

        return 0
    else
        log_error "公证失败"

        # 获取日志
        if echo "$output" | grep -q "id:"; then
            local submission_id=$(echo "$output" | grep "id:" | head -1 | awk '{print $2}')
            log_info "获取公证日志..."
            xcrun notarytool log "$submission_id" \
                --apple-id "$APPLE_ID" \
                --password "$APP_PASSWORD" \
                ${TEAM_ID:+--team-id "$TEAM_ID"} \
                /dev/stdout 2>&1 || true
        fi

        # 清理 DMG
        rm -f "$dmg_path"

        return 1
    fi
}

# 附加票证到应用
staple_ticket() {
    local path="$1"
    log_info "附加公证票证到: $path"

    if xcrun stapler staple "$path" 2>&1; then
        log_info "票证附加成功"

        # 验证票证
        if xcrun stapler validate "$path" 2>&1; then
            log_info "票证验证成功"
        else
            log_warn "票证验证失败"
        fi
    else
        log_error "票证附加失败"
        return 1
    fi
}

# 最终验证
final_verification() {
    log_info "执行最终验证..."

    # 签名验证
    if codesign --verify --deep --strict "$APP_PATH" 2>&1; then
        log_info "签名验证通过"
    else
        log_error "签名验证失败"
        return 1
    fi

    # 显示最终签名信息
    log_info "最终签名信息:"
    codesign -vv -d "$APP_PATH" 2>&1

    # 检查是否可运行
    log_info "检查应用可执行性..."
    if [[ -x "$APP_PATH/Contents/MacOS/"* ]]; then
        log_info "应用可执行权限正确"
    else
        log_warn "应用可能缺少可执行权限"
    fi

    log_info "验证完成！"
}

# 主函数
main() {
    log_info "OpenClaw macOS 签名和公证工具"
    log_info "================================"

    check_macos
    check_tools
    check_app
    check_entitlements
    check_certificate

    cleanup_signatures
    sign_app
    verify_signature
    notarize_app
    final_verification

    log_info "所有步骤完成！"
}

# 运行主函数
main
