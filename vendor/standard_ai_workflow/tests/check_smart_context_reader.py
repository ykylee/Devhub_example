#!/usr/bin/env python3
"""Smoke tests for the smart_context_reader MCP tool."""

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"
if str(SOURCE_ROOT) not in sys.path:
    sys.path.insert(0, str(SOURCE_ROOT))

from workflow_kit.server.read_only_tools import invoke_read_only_tool

class TestSmartContextReader(unittest.TestCase):
    def setUp(self):
        self.temp_file = tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False)
        self.temp_file.write("""
def calculate_total(a, b):
    # This calculates total
    return a + b

class UserContext:
    def __init__(self, name):
        self.name = name

    def get_name(self):
        return self.name

async def fetch_data():
    return {"data": 123}
""")
        self.temp_file.close()
        self.test_file_path = self.temp_file.name

    def tearDown(self):
        os.remove(self.test_file_path)

    def test_extract_all_symbols(self):
        payload = {"file_path": self.test_file_path, "symbols": []}
        result = invoke_read_only_tool(
            tool_name="smart_context_reader",
            payload=payload,
            tool_version="1.0.0"
        )
        self.assertEqual(result["status"], "ok")
        self.assertEqual(len(result["extracted_content"]), 5)
        self.assertIn("calculate_total", result["extracted_content"][0])
        self.assertIn("UserContext", result["extracted_content"][1])
        self.assertIn("fetch_data", result["extracted_content"][2])
        self.assertEqual(len(result["not_found_symbols"]), 0)

    def test_extract_specific_symbol(self):
        payload = {"file_path": self.test_file_path, "symbols": ["UserContext"]}
        result = invoke_read_only_tool(
            tool_name="smart_context_reader",
            payload=payload,
            tool_version="1.0.0"
        )
        self.assertEqual(result["status"], "ok")
        self.assertEqual(len(result["extracted_content"]), 1)
        self.assertIn("class UserContext:", result["extracted_content"][0])
        self.assertEqual(len(result["not_found_symbols"]), 0)

    def test_missing_symbol(self):
        payload = {"file_path": self.test_file_path, "symbols": ["UserContext", "MissingClass"]}
        result = invoke_read_only_tool(
            tool_name="smart_context_reader",
            payload=payload,
            tool_version="1.0.0"
        )
        self.assertEqual(result["status"], "ok")
        self.assertEqual(len(result["extracted_content"]), 1)
        self.assertIn("class UserContext:", result["extracted_content"][0])
        self.assertEqual(result["not_found_symbols"], ["MissingClass"])
        self.assertTrue(any("MissingClass" in w for w in result["warnings"]))

    def test_invalid_file_extension(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as txt_file:
            txt_file.write("def foo(): pass")
            txt_file_path = txt_file.name

        try:
            payload = {"file_path": txt_file_path}
            result = invoke_read_only_tool(
                tool_name="smart_context_reader",
                payload=payload,
                tool_version="1.0.0"
            )
            self.assertEqual(result["status"], "error")
            self.assertTrue(any("지원하지 않는 파일 형식" in w for w in result["warnings"]))
        finally:
            os.remove(txt_file_path)

    def test_syntax_error(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as err_file:
            err_file.write("def missing_colon()\n    pass")
            err_path = err_file.name

        try:
            payload = {"file_path": err_path}
            result = invoke_read_only_tool(
                tool_name="smart_context_reader",
                payload=payload,
                tool_version="1.0.0"
            )
            self.assertEqual(result["status"], "error")
            self.assertTrue(any("구문 오류" in w for w in result["warnings"]))
        finally:
            os.remove(err_path)

if __name__ == "__main__":
    unittest.main()
