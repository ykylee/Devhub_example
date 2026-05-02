{
  "descriptor_target": "mcp_tools_list_draft",
  "harness_examples": {
    "codex": {
      "apply_mode": "manual_review_only",
      "bridge_entrypoint": "workflow_kit.server.read_only_jsonrpc",
      "content": "# Draft only: generated from schemas/read_only_transport_descriptors.json.\n# transport_ready=false; do not paste this as an active server until an MCP SDK server loop exists.\n# Tools described: latest_backlog, check_doc_metadata, check_doc_links, suggest_impacted_docs, create_backlog_entry, create_session_handoff_draft, create_environment_record_stub, check_quickstart_stale_links, summarize_git_history, rotate_workflow_logs, assess_milestone_progress, smart_context_reader\n# [mcp_servers.standardAiWorkflowReadOnly]\n# command = \"python3\"\n# args = [\"-m\", \"workflow_kit.server.read_only_jsonrpc\", \"--stdio-lines\"]\n# NOTE: current bridge is a JSON-RPC draft fixture, not a full MCP SDK server.",
      "format": "toml_snippet_draft",
      "server_alias": "standardAiWorkflowReadOnly",
      "target_path": "~/.codex/config.toml"
    },
    "opencode": {
      "apply_mode": "manual_review_only",
      "bridge_entrypoint": "workflow_kit.server.read_only_jsonrpc",
      "content": "{\n  // Draft only: generated from schemas/read_only_transport_descriptors.json.\n  // transport_ready=false; do not enable until an MCP SDK server loop exists.\n  // Tools described: latest_backlog, check_doc_metadata, check_doc_links, suggest_impacted_docs, create_backlog_entry, create_session_handoff_draft, create_environment_record_stub, check_quickstart_stale_links, summarize_git_history, rotate_workflow_logs, assess_milestone_progress, smart_context_reader\n  \"mcp\": {\n    // \"standardAiWorkflowReadOnly\": {\n    //   \"type\": \"local\",\n    //   \"command\": \"python3\",\n    //   \"args\": [\"-m\", \"workflow_kit.server.read_only_jsonrpc\", \"--stdio-lines\"]\n    // }\n  }\n}",
      "format": "jsonc_snippet_draft",
      "server_alias": "standardAiWorkflowReadOnly",
      "target_path": "opencode.json"
    }
  },
  "source_descriptor_path": "schemas/read_only_transport_descriptors.json",
  "status": "ok",
  "tool_count": 12,
  "tool_names": [
    "latest_backlog",
    "check_doc_metadata",
    "check_doc_links",
    "suggest_impacted_docs",
    "create_backlog_entry",
    "create_session_handoff_draft",
    "create_environment_record_stub",
    "check_quickstart_stale_links",
    "summarize_git_history",
    "rotate_workflow_logs",
    "assess_milestone_progress",
    "smart_context_reader"
  ],
  "tool_version": "beta-v3.0.1",
  "transport_ready": false
}
