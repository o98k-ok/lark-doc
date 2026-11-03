#!/bin/bash
set -euo pipefail
root="$(cd "$(dirname "$0")" && pwd)"
case "${1:-}" in
  lark-doc:login)
    /usr/bin/osascript - "$root/login.command" <<'APPLESCRIPT'
on run argv
  tell application "Terminal"
    activate
    do script ("/bin/bash " & quoted form of (item 1 of argv))
  end tell
end run
APPLESCRIPT
    ;;
  lark-doc:logout)
    "$root/lark-doc" logout
    ;;
  https://*.feishu.cn/*)
    /usr/bin/open "$1"
    ;;
  *) echo 'Unsupported action' >&2; exit 1 ;;
esac
