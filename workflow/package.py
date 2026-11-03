#!/usr/bin/env python3
"""Package only workflow assets and a supplied macOS binary, never user data."""
import pathlib
import plistlib
import sys
import zipfile

root = pathlib.Path(__file__).resolve().parent
binary, output = map(pathlib.Path, sys.argv[1:])
objects = [
    {"uid": "search", "type": "alfred.workflow.input.scriptfilter", "version": 3,
     "config": {"keyword": "lark", "title": "搜索飞书文档", "subtext": "首次使用请回车登录飞书",
                "script": 'unset session\n./lark-doc query "$1"', "type": 0, "scriptargtype": 1,
                "withspace": True, "argumenttype": 1, "alfredfiltersresults": False}},
    {"uid": "login", "type": "alfred.workflow.input.keyword", "version": 1,
     "config": {"keyword": "lark-login", "text": "登录飞书", "subtext": "打开本机终端扫码登录",
                "argumenttype": 2, "withspace": False}},
    {"uid": "logout", "type": "alfred.workflow.input.keyword", "version": 1,
     "config": {"keyword": "lark-logout", "text": "退出飞书", "subtext": "清除本机凭证和缓存；不撤销服务端会话",
                "argumenttype": 2, "withspace": False}},
    {"uid": "action", "type": "alfred.workflow.action.script", "version": 2,
     "config": {"script": './action.sh "$1"', "type": 0, "scriptargtype": 1}},
    {"uid": "login-action", "type": "alfred.workflow.action.script", "version": 2,
     "config": {"script": './action.sh lark-doc:login', "type": 0, "scriptargtype": 1}},
    {"uid": "logout-action", "type": "alfred.workflow.action.script", "version": 2,
     "config": {"script": 'unset session\n./action.sh lark-doc:logout', "type": 0, "scriptargtype": 1}},
]
manifest = {"bundleid": "com.o98k.lark-doc", "name": "Lark Doc", "version": "0.1.0",
            "description": "本机扫码登录和飞书文档搜索", "createdby": "o98k-ok",
            "objects": objects,
            "connections": {source: [{"destinationuid": target, "modifiers": 0, "modifiersubtext": "", "vitoclose": False}]
                            for source, target in [("search", "action"), ("login", "login-action"), ("logout", "logout-action")]}}
with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
    archive.writestr("info.plist", plistlib.dumps(manifest))
    for name, source in [("lark-doc", binary), ("action.sh", root / "action.sh"), ("login.command", root / "login.command")]:
        entry = zipfile.ZipInfo(name)
        entry.external_attr = 0o100755 << 16
        archive.writestr(entry, source.read_bytes())
print(output)
