This project uses the **costwise** MCP server. Its tools keep the session cheap: the dominant cost is re-reading context each turn, so keep the window small.

**Route large content out of context, don't paste it inline.**
- For large output, call `stash_context` to park it and get a handle, then `recall(source=<handle>, query=…)` for only what you need.
- Persist facts with `remember`; retrieve with `recall` instead of re-pasting.

**Prefer narrow retrieval over reading whole files.** Read a full file only when a targeted query can't answer it.
- Use the right tool: `find_symbol`/`read_symbol` for symbols, `find_references`/`find_callers` for usage, `search_code` for questions, `get_repository_summary` for structure. `recall` is for facts/stashes. For raw regex, use the host's grep.
- Default budget unless insufficient — one `large` call adds ~10k uncached tokens.

**Start with `session_brief`** to catch up on prior context. Use `sessions=5` to recall the last 5 sessions' work.

**Run `validate_architecture`** after modifying architecture to ensure compliance with `costwise-architecture.yaml` policy.
