#!/bin/sh
# 安装、升级或卸载 AlonaBot，保留运行数据并按发行摘要校验下载内容。
set -eu

fail() {
    printf 'error: %s\n' "$*" >&2
    exit 1
}

require() {
    for tool do
        command -v "$tool" >/dev/null 2>&1 || fail "required tool not found: $tool; use a release archive or container instead"
    done
}

version=''
uninstall=false
while [ "$#" -gt 0 ]; do
    case "$1" in
        --version)
            [ "$#" -ge 2 ] && [ -z "$version" ] && [ "$uninstall" = false ] || fail 'usage: install.sh [--version VERSION] | --uninstall'
            version=$2
            [ -n "$version" ] || fail 'version must be a single path component'
            shift 2
            ;;
        --uninstall)
            [ "$#" -eq 1 ] && [ -z "$version" ] && [ "$uninstall" = false ] || fail 'usage: install.sh [--version VERSION] | --uninstall'
            uninstall=true
            shift
            ;;
        *) fail 'usage: install.sh [--version VERSION] | --uninstall' ;;
    esac
done

validate_version() {
    case "$version" in
        ''|.|..|*/*|*\\*|*' '*|*'	'*|*'
'*) fail 'version must be a single path component' ;;
    esac
}

require uname id
case "$(uname -s)" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    *) fail 'unsupported OS; use a release archive or container instead' ;;
esac
case "$(uname -m)" in
    x86_64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) fail 'unsupported architecture; use a release archive or container instead' ;;
esac
require mkdir rm mv chmod cat
if [ "$os" = linux ]; then
    [ "$(id -u)" -eq 0 ] || fail 'Linux installation requires root; use a release archive or container instead'
    require systemctl useradd runuser chown
    [ -d /run/systemd/system ] || fail 'systemd is required; use a release archive or container instead'
    systemctl show-environment >/dev/null
    program_dir=/opt/alonabot
    data_dir=/var/lib/alonabot
    service_file=/etc/systemd/system/alonabot.service
    service=alonabot.service
else
    [ "$(id -u)" -ne 0 ] || fail 'run macOS installation as the logged-in user; use a release archive or container instead'
    [ -n "${HOME:-}" ] && [ -d "$HOME" ] && [ -w "$HOME" ] || fail 'a writable home directory is required; use a release archive or container instead'
    require launchctl sed
    domain=gui/$(id -u)
    launchctl print "$domain" >/dev/null
    program_dir=$HOME/Library/Application\ Support/AlonaBot
    data_dir=$program_dir/data
    service_file=$HOME/Library/LaunchAgents/io.github.alonabot.plist
    service=io.github.alonabot
fi

stop_service() {
    if [ "$os" = linux ]; then
        if [ -f "$service_file" ]; then
            systemctl stop "$service"
        fi
    elif launchctl print "$domain/$service" >/dev/null 2>&1; then
        launchctl bootout "$domain/$service"
    fi
}

if [ "$uninstall" = true ]; then
    stop_service
    if [ "$os" = linux ] && [ -f "$service_file" ]; then
        systemctl disable "$service"
    fi
    rm -f "$service_file" "$program_dir/alonabot"
    rm -rf "$program_dir/packages"
    if [ "$os" = linux ]; then
        systemctl daemon-reload
    fi
    printf 'Uninstalled AlonaBot. Data preserved at: %s\n' "$data_dir"
    exit 0
fi

require curl tar mktemp awk od
if [ "$os" = linux ]; then
    require sha256sum
else
    require shasum
fi
release_base=${ALONABOT_RELEASE_BASE_URL-https://github.com/AlonaBot/AlonaBot/releases}
release_base=${release_base%/}
case "$release_base" in
    https://?*) ;;
    *) fail 'release base URL must use HTTPS' ;;
esac
case "$release_base" in
    *\?*|*\#*) fail 'release base URL must not contain a query or fragment' ;;
esac

# curl 对初始地址和每次重定向都强制 HTTPS，并保留下载错误。
fetch() {
    curl --fail --silent --show-error --location --globoff --proto '=https' --proto-redir '=https' "$@"
}

# 对整个路径组件编码，版本沿用 dist 的单路径组件约定。
url_component() {
    printf '%s' "$1" | od -An -v -tx1 | awk '{ for (i = 1; i <= NF; i++) printf "%%%s", toupper($i) }'
}

if [ -z "$version" ]; then
    resolved=$(fetch --output /dev/null --write-out '%{url_effective}' "$release_base/latest")
    case "$resolved" in
        "$release_base"/tag/v?*) version=${resolved#"$release_base/tag/v"} ;;
        *) fail "latest release did not resolve to a version tag: $resolved" ;;
    esac
    # curl 的最终 URL 可能已经转义；只解码 %HH，原有反斜线保持原样。
    version=$(printf '%s\n' "$version" | awk '
        function hex(c) { return index("0123456789ABCDEF", toupper(c)) - 1 }
        { for (i = 1; i <= length($0); i++) {
            c = substr($0, i, 1)
            if (c == "%" && i + 2 <= length($0) && hex(substr($0,i+1,1)) >= 0 && hex(substr($0,i+2,1)) >= 0) {
                n = hex(substr($0,i+1,1)) * 16 + hex(substr($0,i+2,1))
                if (n == 0) exit 1
                printf "%c", n; i += 2
            } else printf "%s", c
        }}
        END { printf "." }') || fail 'version tag contains an invalid path component'
    version=${version%.}
fi
validate_version
archive_root=alonabot-$version-$os-$arch
archive=$archive_root.tar.gz
release_url=$release_base/download/v$(url_component "$version")

umask 077
staging=$(mktemp -d "${TMPDIR:-/tmp}/alonabot-install.XXXXXX")
cleanup() {
    result=$?
    trap - 0
    if rm -rf "$staging"; then :; else
        [ "$result" -ne 0 ] || result=1
    fi
    exit "$result"
}
trap cleanup 0
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
fetch --output "$staging/$archive" "$release_url/$(url_component "$archive")"
fetch --output "$staging/SHA256SUMS" "$release_url/SHA256SUMS"
expected=$(awk -v name="$archive" '
    substr($0, 67) == name && substr($0, 65, 2) == "  " {
        count++; digest = substr($0, 1, 64)
    }
    END {
        if (count != 1 || length(digest) != 64 || digest ~ /[^0-9a-fA-F]/) exit 1
        print tolower(digest)
    }' "$staging/SHA256SUMS") || fail "missing or ambiguous checksum for $archive"
if [ "$os" = linux ]; then
    actual=$(sha256sum "$staging/$archive")
else
    actual=$(shasum -a 256 "$staging/$archive")
fi
actual=${actual%% *}
[ "$actual" = "$expected" ] || fail "checksum mismatch for $archive"

# 归档边界在临时目录内验证，拒绝绝对路径、越界路径与链接后再解压。
tar -tzf "$staging/$archive" > "$staging/entries"
awk -v root="$archive_root" '
    { if ($0 != root && index($0, root "/") != 1) exit 1
      n = split($0, parts, "/"); for (i = 1; i <= n; i++) if (parts[i] == "..") exit 1 }
' "$staging/entries" || fail 'archive contains a path outside its release directory'
tar -tvzf "$staging/$archive" > "$staging/modes"
awk 'substr($0,1,1) != "d" && substr($0,1,1) != "-" { exit 1 }' "$staging/modes" || fail 'archive contains an unsupported entry type'
mkdir "$staging/unpacked"
tar -xzf "$staging/$archive" -C "$staging/unpacked"
source_dir=$staging/unpacked/$archive_root
[ -f "$source_dir/alonabot" ] && [ -x "$source_dir/alonabot" ] && [ -d "$source_dir/packages" ] || fail 'archive is missing alonabot or packages'

stop_service
if [ "$os" = linux ]; then
    if ! id alonabot >/dev/null 2>&1; then
        useradd --system --user-group --home-dir "$data_dir" --no-create-home --shell /usr/sbin/nologin alonabot
    fi
fi
mkdir -p "$program_dir" "$data_dir"
if [ "$os" = linux ]; then
    chown alonabot:alonabot "$data_dir"
fi
chmod 700 "$data_dir"
# data/ 在 macOS 程序目录内，始终只替换两个发行入口。
rm -f "$program_dir/alonabot"
rm -rf "$program_dir/packages"
mv "$source_dir/alonabot" "$program_dir/alonabot"
mv "$source_dir/packages" "$program_dir/packages"
chmod 755 "$program_dir" "$program_dir/alonabot"
chmod -R a+rX "$program_dir/packages"
if [ "$os" = linux ]; then
    chown -R root:root "$program_dir/alonabot" "$program_dir/packages"
fi

# 先由当前用户打开控制终端，再把同一描述符交给服务身份，避免终端属主限制。
if ( : </dev/tty >/dev/tty ) 2>/dev/null; then
    if [ "$os" = linux ]; then
        runuser -u alonabot -- "$program_dir/alonabot" admin-token init --if-missing --data-dir "$data_dir" </dev/tty 2>/dev/tty
    else
        "$program_dir/alonabot" admin-token init --if-missing --data-dir "$data_dir" </dev/tty 2>/dev/tty
    fi
fi

if [ "$os" = linux ]; then
    cat > "$service_file" <<EOF_SYSTEMD
[Unit]
Description=AlonaBot
After=network.target

[Service]
Type=simple
User=alonabot
Group=alonabot
WorkingDirectory=/opt/alonabot
ExecStart=/opt/alonabot/alonabot run --data-dir /var/lib/alonabot
UMask=0077

[Install]
WantedBy=multi-user.target
EOF_SYSTEMD
    chmod 644 "$service_file"
    systemctl daemon-reload
    systemctl enable --now "$service"
else
    mkdir -p "$HOME/Library/LaunchAgents"
    xml_program=$(printf '%s' "$program_dir" | sed 's/\&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g; s/'"'"'/\&apos;/g')
    cat > "$service_file" <<EOF_LAUNCHD
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>io.github.alonabot</string>
<key>ProgramArguments</key><array>
<string>$xml_program/alonabot</string><string>run</string>
<string>--data-dir</string><string>$xml_program/data</string>
</array>
<key>WorkingDirectory</key><string>$xml_program</string>
<key>RunAtLoad</key><true/>
<key>Umask</key><integer>63</integer>
</dict></plist>
EOF_LAUNCHD
    chmod 644 "$service_file"
    launchctl enable "$domain/$service"
    launchctl bootstrap "$domain" "$service_file"
fi
printf 'Installed AlonaBot %s. Data directory: %s\n' "$version" "$data_dir"
printf 'If %s/admin_token exists, retrieve the first token with:\n' "$data_dir"
if [ "$os" = linux ]; then
    printf '  runuser -u alonabot -- /opt/alonabot/alonabot admin-token show --data-dir /var/lib/alonabot\n'
else
    # 显示的命令可直接粘贴到 shell，包含空格与单引号的主目录同样可用。
    quoted_program=$(printf '%s' "$program_dir" | sed "s/'/'\\\\''/g")
    printf "  '%s/alonabot' admin-token show --data-dir '%s/data'\n" "$quoted_program" "$quoted_program"
fi
printf 'The first token file is deleted after the first successful login.\n'
