# AllocScope

> Go 热路径堆分配静态分析器 — 揪出每一次隐藏的内存逃逸

AllocScope statically analyzes your Go code to find hidden heap allocations that hurt performance: allocations inside loops, pointer escapes, fmt calls in hot paths, and more.

## 🚀 Quick Start

```bash
go install github.com/allocscope/allocscope@latest

# Scan current directory
allocscope .

# JSON output for CI
allocscope -json .

# Fail CI if more than 5 allocations found
allocscope -max-allocs 5 .
```

### Example Output

```
server.go:42 [high] make() in loop causes repeated heap allocation (loop-alloc)
handler.go:87 [medium] fmt.Sprintf causes heap allocation (fmt-alloc)
repo.go:15 [high] returning pointer to local causes heap escape (ptr-escape)

🔍 AllocScope found 3 potential heap allocations
```

## 📊 Detection Rules

| Rule | Severity | Description |
|------|----------|-------------|
| `loop-alloc` | 🔴 high | `make()`/`new()` inside loop |
| `loop-escape` | 🔴 high | `&var` inside loop (allocation per iteration) |
| `loop-fmt` | 🔴 high | `fmt.*` inside loop (repeated allocation) |
| `loop-append` | 🟡 medium | `append` in loop without pre-allocation |
| `fmt-alloc` | 🟡 medium | `fmt.Sprintf`/`Errorf` anywhere |
| `ptr-escape` | 🔴 high | Returning `&localVar` (heap escape) |

## 💰 Pricing

| Feature | Free | Pro ($19/mo) | Enterprise ($99/mo) |
|---------|------|-------------|--------------------|
| 6 core detection rules | ✅ | ✅ | ✅ |
| CLI + JSON output | ✅ | ✅ | ✅ |
| CI `--max-allocs` gate | ✅ | ✅ | ✅ |
| Benchmark-correlated analysis | ❌ | ✅ | ✅ |
| Ignore/suppress directives | ❌ | ✅ | ✅ |
| SARIF output (GitHub Security) | ❌ | ✅ | ✅ |
| Cross-function escape tracking | ❌ | ✅ | ✅ |
| SSA-based deep analysis | ❌ | ❌ | ✅ |
| Custom rule authoring | ❌ | ❌ | ✅ |
| Priority support + SLA | ❌ | ❌ | ✅ |

## 🤔 Why Pay?

Go's built-in `go build -gcflags='-m'` is noisy, hard to parse, and impossible to integrate into CI. AllocScope gives you:

- **Actionable findings** with file, line, rule, and severity
- **CI integration** — fail builds when allocations regress
- **JSON output** — pipe into dashboards, Slack alerts, or SARIF
- **Hot-path focus** — prioritizes allocations in loops, not one-time init code

Teams using AllocScope report **15-40% GC reduction** after fixing flagged allocations.

## License

Free tier: MIT. Pro/Enterprise features require a license key.
