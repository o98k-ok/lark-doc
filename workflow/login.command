#!/bin/bash
set -euo pipefail
root="$(cd "$(dirname "$0")" && pwd)"
unset session
exec "$root/lark-doc" login
