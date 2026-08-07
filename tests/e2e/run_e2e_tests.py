#!/usr/bin/env python3
"""
CostWise MCP Server — Cross-Platform End-to-End Test Suite

Runs against any platform (Linux, macOS, Windows) by spawning the `costwise`
binary in JSON-RPC stdio mode and executing full protocol and tool integration tests.
"""

import sys
import os
import json
import shutil
import tempfile
import subprocess
import time
from pathlib import Path

# ANSI colors for cross-platform terminal output
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    CYAN = '\033[96m'
    BOLD = '\033[1m'
    RESET = '\033[0m'

    @staticmethod
    def disable():
        Colors.GREEN = Colors.RED = Colors.YELLOW = Colors.CYAN = Colors.BOLD = Colors.RESET = ''

if sys.platform == "win32":
    os.system("") # Enable ANSI escape codes in Windows CMD/PowerShell

class CostWiseTestRunner:
    def __init__(self, binary_path=None):
        self.binary_path = binary_path or self._find_binary()
        self.temp_repo = None
        self.proc = None
        self.msg_id = 0
        self.passed = 0
        self.failed = 0

    def _find_binary(self):
        """Find costwise executable in PATH, ~/.local/bin, or current project build."""
        # 1. Check user profile / .local/bin
        local_bin = Path.home() / ".local" / "bin"
        executable = "costwise.exe" if sys.platform == "win32" else "costwise"
        
        target_path = local_bin / executable
        if target_path.exists():
            return str(target_path)
            
        # 2. Check shutil.which
        which_path = shutil.which("costwise")
        if which_path:
            return which_path

        # 3. Check current workspace build
        proj_build = Path.cwd() / executable
        if proj_build.exists():
            return str(proj_build)
            
        raise FileNotFoundError(
            f"Could not locate '{executable}'. Please build costwise or specify --binary path."
        )

    def _create_temp_repository(self):
        """Create a mock repository with sample code files to index and test."""
        self.temp_repo = tempfile.mkdtemp(prefix="costwise_e2e_")
        repo_dir = Path(self.temp_repo)

        # 1. Main Go file
        (repo_dir / "main.go").write_text("""package main

import "fmt"

func ProcessUserData(userID string) string {
    fmt.Println("Processing user data for:", userID)
    return "SUCCESS"
}

func main() {
    res := ProcessUserData("user_123")
    fmt.Println(res)
}
""", encoding="utf-8")

        # 2. Python utils file
        (repo_dir / "utils.py").write_text("""import sys

def calculate_metrics(data_list):
    \"\"\"Calculate basic statistical metrics.\"\"\"
    if not data_list:
        return 0
    return sum(data_list) / len(data_list)

class DataProcessor:
    def __init__(self, name):
        self.name = name

    def execute(self):
        print(f"Executing processor {self.name}")
""", encoding="utf-8")

        # 3. JavaScript app file
        (repo_dir / "app.js").write_text("""function authenticateUser(token) {
    console.log("Authenticating token:", token);
    return true;
}

module.exports = { authenticateUser };
""", encoding="utf-8")

        return str(repo_dir)

    def start_server(self):
        """Spawn the costwise process in stdio mode."""
        print(f"{Colors.CYAN}Starting CostWise server: {self.binary_path} serve{Colors.RESET}")
        self.proc = subprocess.Popen(
            [self.binary_path, "serve"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1
        )

    def stop_server(self):
        """Terminate the server process and clean up temporary files."""
        if self.proc:
            try:
                self.proc.terminate()
                self.proc.wait(timeout=3)
            except Exception:
                self.proc.kill()
        if self.temp_repo and os.path.exists(self.temp_repo):
            shutil.rmtree(self.temp_repo, ignore_errors=True)

    def _send_request(self, method, params=None):
        """Send a JSON-RPC 2.0 request and receive the response."""
        self.msg_id += 1
        request = {
            "jsonrpc": "2.0",
            "id": self.msg_id,
            "method": method
        }
        if params is not None:
            request["params"] = params

        raw_req = json.dumps(request) + "\n"
        self.proc.stdin.write(raw_req)
        self.proc.stdin.flush()

        # Read line from stdout
        raw_resp = self.proc.stdout.readline()
        if not raw_resp:
            stderr_out = self.proc.stderr.read()
            raise RuntimeError(f"Server closed connection unexpectedly. Stderr: {stderr_out}")

        return json.loads(raw_resp)

    def _call_tool(self, tool_name, arguments):
        """Invoke an MCP tool via tools/call."""
        return self._send_request("tools/call", {
            "name": tool_name,
            "arguments": arguments
        })

    def run_test(self, test_name, test_fn):
        """Run an individual test case and track score."""
        print(f"Testing {Colors.BOLD}{test_name}{Colors.RESET}... ", end="", flush=True)
        try:
            test_fn()
            print(f"{Colors.GREEN}PASSED{Colors.RESET}")
            self.passed += 1
        except Exception as e:
            print(f"{Colors.RED}FAILED{Colors.RESET}")
            print(f"  {Colors.RED}Error: {e}{Colors.RESET}")
            self.failed += 1

    def test_initialize(self):
        """Test JSON-RPC initialize handshake."""
        resp = self._send_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "CostWise-E2E-Tester", "version": "1.0.0"}
        })
        assert "result" in resp, f"Expected result in initialize response, got: {resp}"
        assert resp["result"].get("protocolVersion") == "2024-11-05"

        # Send initialized notification
        self.proc.stdin.write(json.dumps({"jsonrpc": "2.0", "method": "notifications/initialized"}) + "\n")
        self.proc.stdin.flush()

    def test_tools_list(self):
        """Test tools/list returns all 12 registered MCP tools."""
        resp = self._send_request("tools/list")
        assert "result" in resp and "tools" in resp["result"]
        tools = resp["result"]["tools"]
        tool_names = {t["name"] for t in tools}

        expected_tools = {
            "search_code", "find_symbol", "read_symbol", "find_references",
            "find_callers", "get_repository_summary", "index_repository",
            "remember", "stash_context", "recall", "allow_dir", "session_brief"
        }

        missing = expected_tools - tool_names
        assert not missing, f"Missing tools from tools/list: {missing}"

    def test_tool_index_repository(self):
        """Test index_repository tool."""
        resp = self._call_tool("index_repository", {"repo_path": self.temp_repo})
        assert "result" in resp, f"Failed index_repository: {resp}"
        content = resp["result"]["content"][0]["text"]
        assert "Indexed" in content or "re-index" in content or "OK" in content

    def test_tool_get_repository_summary(self):
        """Test get_repository_summary tool."""
        resp = self._call_tool("get_repository_summary", {
            "repo_path": self.temp_repo,
            "budget": "small"
        })
        assert "result" in resp
        content = resp["result"]["content"][0]["text"]
        assert len(content) > 0

    def test_tool_search_code(self):
        """Test search_code tool."""
        resp = self._call_tool("search_code", {
            "repo_path": self.temp_repo,
            "query": "ProcessUserData",
            "budget": "medium"
        })
        assert "result" in resp
        content = resp["result"]["content"][0]["text"]
        assert "ProcessUserData" in content or "main.go" in content

    def test_tool_find_symbol(self):
        """Test find_symbol tool."""
        resp = self._call_tool("find_symbol", {
            "repo_path": self.temp_repo,
            "symbol": "calculate_metrics"
        })
        assert "result" in resp
        content = resp["result"]["content"][0]["text"]
        assert "utils.py" in content or "calculate_metrics" in content

    def test_tool_read_symbol(self):
        """Test read_symbol tool."""
        resp = self._call_tool("read_symbol", {
            "repo_path": self.temp_repo,
            "symbol": "DataProcessor"
        })
        assert "result" in resp
        content = resp["result"]["content"][0]["text"]
        assert "DataProcessor" in content

    def test_tool_remember_and_recall(self):
        """Test remember and recall tools."""
        # 1. Remember fact
        resp1 = self._call_tool("remember", {
            "repo_path": self.temp_repo,
            "key": "auth-status",
            "fact": "Authentication is token based."
        })
        assert "result" in resp1

        # 2. Recall fact
        resp2 = self._call_tool("recall", {
            "repo_path": self.temp_repo,
            "query": "Authentication",
            "source": "facts"
        })
        assert "result" in resp2
        content = resp2["result"]["content"][0]["text"]
        assert "token based" in content or "auth-status" in content

    def test_tool_stash_context(self):
        """Test stash_context tool."""
        large_text = "Line 1: Log output line\n" * 100
        resp = self._call_tool("stash_context", {
            "repo_path": self.temp_repo,
            "content": large_text,
            "label": "test-log"
        })
        assert "result" in resp
        content = resp["result"]["content"][0]["text"]
        assert "stash" in content or "handle" in content or "Stashed" in content

    def test_tool_session_brief(self):
        """Test session_brief tool."""
        resp = self._call_tool("session_brief", {
            "repo_path": self.temp_repo,
            "scope": "all"
        })
        assert "result" in resp
        content = resp["result"]["content"][0]["text"]
        assert len(content) > 0

    def run_all(self):
        """Execute the full E2E test suite."""
        print(f"{Colors.BOLD}===================================================={Colors.RESET}")
        print(f"{Colors.BOLD} CostWise MCP Cross-Platform E2E Test Suite{Colors.RESET}")
        print(f"{Colors.BOLD} Platform: {sys.platform} ({os.name}){Colors.RESET}")
        print(f"{Colors.BOLD} Binary: {self.binary_path}{Colors.RESET}")
        print(f"{Colors.BOLD}===================================================={Colors.RESET}\n")

        repo_dir = self._create_temp_repository()
        self.start_server()

        try:
            self.run_test("JSON-RPC Protocol Initialize", self.test_initialize)
            self.run_test("MCP Tools Registration List", self.test_tools_list)
            self.run_test("Tool: index_repository", self.test_tool_index_repository)
            self.run_test("Tool: get_repository_summary", self.test_tool_get_repository_summary)
            self.run_test("Tool: search_code", self.test_tool_search_code)
            self.run_test("Tool: find_symbol", self.test_tool_find_symbol)
            self.run_test("Tool: read_symbol", self.test_tool_read_symbol)
            self.run_test("Tool: remember & recall", self.test_tool_remember_and_recall)
            self.run_test("Tool: stash_context", self.test_tool_stash_context)
            self.run_test("Tool: session_brief", self.test_tool_session_brief)
        finally:
            self.stop_server()

        print(f"\n{Colors.BOLD}===================================================={Colors.RESET}")
        print(f"Results: {Colors.GREEN}{self.passed} PASSED{Colors.RESET}, {Colors.RED}{self.failed} FAILED{Colors.RESET}")
        print(f"{Colors.BOLD}===================================================={Colors.RESET}")

        if self.failed > 0:
            sys.exit(1)

def main():
    binary_arg = None
    if len(sys.argv) > 1 and sys.argv[1].startswith("--binary="):
        binary_arg = sys.argv[1].split("=", 1)[1]
    
    runner = CostWiseTestRunner(binary_arg)
    runner.run_all()

if __name__ == "__main__":
    main()
