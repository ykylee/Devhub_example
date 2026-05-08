# MCP v1.0 Migration Specification

## 1. 개요
현재 사용 중인 커스텀 `read_only_mcp_sdk` 및 `mcp_main` 러너를 공식 MCP(Model Context Protocol) Python SDK로 전환하기 위한 기술 명세입니다.

## 2. 변경 대상 (Migration Targets)

### 2.1 서버 라이브러리
- **AS-IS**: `workflow_kit.server.read_only_mcp_sdk`
- **TO-BE**: `mcp.server.fastmcp.FastMCP` 또는 `mcp.server.Server`

### 2.2 도구 선언 (Tool Definitions)
- **AS-IS**: `registry.register_tool` 데코레이터 및 수동 스키마 작성.
- **TO-BE**: `@mcp.tool()` 데코레이터 및 Python 타입 힌트(Type Hints) 자동 추출.

### 2.3 데이터 직렬화 (Serialization)
- **AS-IS**: 일반 Python 딕셔너리 리턴.
- **TO-BE**: Pydantic 모델 리턴 (SDK에서 자동 변환 지원).

## 3. 마이그레이션 단계 (Migration Steps)

### 3.1 의존성 업데이트
`requirements.txt` 또는 `pyproject.toml`에 `mcp>=1.0.0` 추가.

### 3.2 서버 코드 리팩토링 예시
```python
# AS-IS (legacy)
from workflow_kit.server.read_only_mcp_sdk import MCPServer
server = MCPServer("backlog-server")

@server.tool("get_backlog", description="Get current backlog")
def get_backlog():
    return {"status": "success", "data": [...]}

# TO-BE (MCP v1.0)
from mcp.server.fastmcp import FastMCP
mcp = FastMCP("Standard Workflow Server")

@mcp.tool()
def get_backlog() -> dict:
    """Get current backlog"""
    return {"status": "success", "data": [...]}
```

### 3.3 러너(Runner) 수정
`mcp_main.py` 유틸리티가 공식 SDK의 `stdio` 트랜스포트를 지원하도록 업데이트.

## 4. 검증 계획
1. **타입 안전성**: 모든 도구가 타입 힌트를 명확히 선언했는지 검수.
2. **상호운용성**: `mcp-cli`를 사용하여 공식 프로토콜 규격(Initialize, ListTools, CallTool) 준수 여부 확인.
3. **하네스 연동**: Antigravity 및 Codex 하네스에서 공식 SDK 기반 서버가 정상적으로 로드되는지 확인.
