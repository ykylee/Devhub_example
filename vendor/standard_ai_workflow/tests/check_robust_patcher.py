#!/usr/bin/env python3
import os
import sys
import tempfile
import subprocess
from pathlib import Path

# 패치 엔진의 경로 설정
SCRIPT_DIR = Path(__file__).resolve().parent
SOURCE_ROOT = SCRIPT_DIR.parent
SKILL_ROOT = SOURCE_ROOT / "skills" / "robust_patcher"
ENGINE_PATH = SKILL_ROOT / "scripts" / "patch_engine.py"

S = "<<<<<<<" + " SEARCH"
E = "======="
R = ">>>>>>>" + " REPLACE"

import json

def run_patch(target_file, patch_content):
    with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as patch_file:
        patch_file.write(patch_content)
        patch_path = patch_file.name

    try:
        result = subprocess.run(
            [sys.executable, str(ENGINE_PATH), "--file", str(target_file), "--patch-file", patch_path],
            capture_output=True,
            text=True
        )
        try:
            payload = json.loads(result.stdout)
            return payload.get("status") == "ok", payload.get("message", "") or payload.get("error", "")
        except json.JSONDecodeError:
            return False, f"Invalid JSON output: {result.stdout}\n{result.stderr}"
    finally:
        os.remove(patch_path)

def test_exact_match():
    print("Running test_exact_match...")
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write("def hello():\n    print('world')\n")
        target_path = f.name

    patch = f"""{S}
def hello():
    print('world')
{E}
def hello():
    print('universe')
{R}"""

    try:
        success, out = run_patch(target_path, patch)
        assert success, f"Exact match failed: {out}"
        content = Path(target_path).read_text()
        assert "universe" in content
        print("  - OK")
    finally:
        os.remove(target_path)

def test_fuzzy_match():
    print("Running test_fuzzy_match...")
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write("def calc():\n    a = 1\n    \n    b = 2\n    return a + b\n")
        target_path = f.name

    # 패치에 들여쓰기가 다르고, 빈 줄이 빠져 있음
    patch = f"""{S}
def calc():
  a = 1
  b = 2
  return a + b
{E}
def calc():
    a = 10
    b = 20
    return a + b
{R}"""

    try:
        success, out = run_patch(target_path, patch)
        assert success, f"Fuzzy match failed: {out}"
        content = Path(target_path).read_text()
        assert "a = 10" in content
        assert "b = 20" in content
        print("  - OK")
    finally:
        os.remove(target_path)

def test_multi_block():
    print("Running test_multi_block...")
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write("def one():\n    return 1\n\ndef two():\n    return 2\n")
        target_path = f.name

    patch = f"""{S}
def one():
    return 1
{E}
def one():
    return 10
{R}
{S}
def two():
    return 2
{E}
def two():
    return 20
{R}"""

    try:
        success, out = run_patch(target_path, patch)
        assert success, f"Multi block failed: {out}"
        content = Path(target_path).read_text()
        assert "return 10" in content
        assert "return 20" in content
        print("  - OK")
    finally:
        os.remove(target_path)

def test_syntax_error():
    print("Running test_syntax_error...")
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write("def valid():\n    pass\n")
        target_path = f.name

    # 문법 오류 유발 (콜론 빠짐)
    patch = f"""{S}
def valid():
    pass
{E}
def valid()
    return True
{R}"""

    try:
        success, out = run_patch(target_path, patch)
        assert not success, "Syntax error test should have failed."
        assert "SyntaxError" in out
        content = Path(target_path).read_text()
        assert "pass" in content # 원본 유지 확인
        print("  - OK")
    finally:
        os.remove(target_path)

if __name__ == "__main__":
    print("Starting robust-patcher tests...\n")
    test_exact_match()
    test_fuzzy_match()
    test_multi_block()
    test_syntax_error()
    print("\nAll tests passed successfully!")
