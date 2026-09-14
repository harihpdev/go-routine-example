# Code Review — go-routine-example

Reviewed against `task.md` (a 4-part Go concurrency exercise: fire-and-forget goroutines, `sync.WaitGroup`, unbuffered channels, buffered channels + backpressure) and general senior/production-readiness standards. Read-only review — no source files were modified.

## Overall verdict

**Not yet senior/production-level.** The concurrency primitives are used correctly in places — `channel.go`/`httpClient.go`'s WaitGroup → unbuffered-channel handoff, and `bufferQueue.go`'s `select`+`default` backpressure pattern are both solid and match `task.md` closely. But `batch.go` (Task 1.2) contains a **process-crashing bug** and a **request-leaking deadlock**, and doesn't actually implement its own spec's core requirement (waiting for goroutines before responding). There are no tests anywhere, despite concurrency correctness being the entire subject of this repo — exactly the class of bug that `go test -race` exists to catch.

---

## Task 1.1 — `POST /log` (fire-and-forget)
Files: `internal/api/log.go`, `internal/util/util.go`

- **Medium** — `util.WriteLog` uses `os.WriteFile`, which truncates the file on every call. Each new log **destroys the previous log content** instead of appending. This contradicts basic "log file" semantics and silently loses data under repeated/concurrent calls.
- **Medium** — no request body size limit (unlike `/fetch-channel` and `/enqueue`, which use `io.LimitReader`). `io.ReadAll(r.Body)` here is unbounded — inconsistent hardening across handlers.
- **Low** — stale `// TODO: handle the received log payload` comment left in `log.go` even though `util.WriteLog` already handles it.
- **Low** — the error path only does `fmt.Println("File Write error")`, dropping the actual `err` value — no way to diagnose a failure.
- ✅ Correct — responds `202` immediately, does the write via `go util.WriteLog(body)`. This is genuine fire-and-forget and meets the task's key learning goal.

## Task 1.2 — `POST /fetch-batch` (`sync.WaitGroup`)
File: `internal/api/batch.go` — **the most severe issues in the repo**

- **Critical** — `fetchUrl` does:
  ```go
  resp, err := http.Get(url)
  if err != nil {
      fmt.Printf("Error fetching %s: %v\n", url, err)
  }
  mu.Lock()
  ch <- "Fetched Url: " + url
  mu.Unlock()
  defer resp.Body.Close()
  ```
  On any failed request, `resp` is `nil`, and `resp.Body.Close()` **panics with a nil-pointer dereference**. An unrecovered panic inside a goroutine terminates the **entire process**, not just that request — a single unreachable URL takes the whole server down.
- **Critical** — `respCh` (unbuffered) is never read anywhere in `batchHandler`. Every `fetchUrl` goroutine blocks forever on `ch <- ...`, so the deferred `wg.Done()` (which runs after the blocked send) never executes, `wg.Wait()` in the closer goroutine never returns, and `respCh` is never closed. **This leaks goroutines on every single request** to this endpoint.
- **Critical** — `task.md` explicitly requires "Call `wg.Wait()` before returning the API response." `batchHandler` returns `202` immediately without waiting for anything — it doesn't implement the WaitGroup-blocking pattern the task is teaching. (Moot anyway given the deadlock above: even if it did wait, it would hang forever.)
- **High** — the mutex around `ch <- ...` is unnecessary; channel sends are already safe without external locking. This reveals a misunderstanding of channel semantics that was correctly fixed in `httpClient.go` (whose comment literally says "Removed unused mutex parameter") but never back-ported here.
- **Low** — dead, commented-out debug lines (`// fmt.Println(...)`) left in.
- ✅ Correct — loop-variable capture is handled properly (URL passed as a goroutine parameter).

## Task 1.3 — `POST /fetch-channel` (unbuffered channel handoff)
Files: `internal/api/channel.go`, `internal/util/httpClient.go` — **best-implemented of the four**

- **Medium** — on a per-URL fetch error, the handler just logs and `continue`s. The client still gets `200 OK` with a partially-populated (possibly zero-value) result and no indication anything failed.
- **Medium** — the type switch's `default` branch (`"Unknown data type returned from %s"`) is unreachable dead code — `types.GetModelForURL` can only return one of three known types or an error (already filtered out earlier), so this branch can never execute.
- **Low** — `ChannelHandler` is exported while every other handler in the package (`batchHandler`, `healthHandler`, `logHandler`) is unexported — inconsistent naming for identically-scoped functions.
- ✅ Correct — unbuffered channel hand-off, WaitGroup + closer goroutine, and per-URL result typing are all implemented soundly.

## Task 1.4 — `POST /enqueue` (buffered channel + worker + backpressure)
File: `internal/api/bufferQueue.go`, worker wired in `internal/api/health.go` — **cleanest match to spec**

- **High** — `NewRouter()` has the side effect of creating the task channel *and* starting the `ProcessQueue` worker goroutine. Bootstrapping infrastructure inside what reads as a pure "build a router" factory is a surprising side effect — if `NewRouter()` is ever called more than once (tests, future refactors), it silently spins up duplicate, un-stoppable workers.
- **High** — no shutdown path for `ProcessQueue`: it `range`s over the channel forever with no signal to stop, and `main.go` never closes the channel or waits for in-flight work to drain.
- **Low** — no validation that `task.TaskId`/`Payload` are non-empty before enqueueing.
- ✅ Correct — `select { case ch <- task: ... default: 503 }` backpressure pattern matches `task.md` exactly — the strongest, cleanest implementation in the repo.

---

## Cross-cutting findings

- **High** — `/fetch-batch`, `/fetch-channel`, and `/enqueue` are registered as bare `"/path"` in `NewRouter()` (no method restriction), while `/health` and `/log` correctly use the Go 1.22+ `"METHOD /path"` pattern. Inconsistent, and the three unrestricted routes silently accept any HTTP method (GET, DELETE, etc.).
- **High** — `cmd/server/main.go` calls `http.ListenAndServe` directly with no `os/signal` handling or `http.Server.Shutdown`. No graceful shutdown for in-flight requests or the queue worker — notable given this repo is specifically about background-goroutine lifecycle.
- **Medium** — **zero test files** (`*_test.go`) anywhere in the repo, despite concurrency correctness being the entire subject matter. `go test -race` would have caught the `batch.go` deadlock/panic immediately.
- **Medium** — no `context.Context` propagation anywhere. Outbound HTTP calls rely only on the client's fixed 5s timeout, with no per-request cancellation wired from `r.Context()`.
- **Medium** — size-limit constants are duplicated and non-idiomatically named: `Max_Request_Size` (`channel.go`) and `Max_Payload_Size` (`bufferQueue.go`) are both `1024*1024` but defined twice, in `snake_case` rather than Go's `MaxRequestSize` convention.
- **Low** — no structured/leveled logging anywhere; `fmt.Println`/`fmt.Printf` used ad hoc in `main.go`, `batch.go`, `channel.go`, `bufferQueue.go`, `util.go`.
- **Low** — leftover debug line `fmt.Println("Your Mac Temp Dir:", os.TempDir())` in `main.go`.
- **Low** — `internal/handler/handler.go` is an empty, unused stub package — dead code.
- **Low** — `go.mod`'s `go 1.27.1` directive is an unusual/likely-invalid version string; worth checking against the actual installed toolchain (`go version`).
- **Low** — recent commit messages ("TEST", "FIX") aren't descriptive.

---

## How to verify the critical findings yourself

```bash
go run ./cmd/server
# in another terminal — a guaranteed-refused connection:
curl -X POST localhost:8080/fetch-batch -d '{"urls":["http://localhost:1"]}'
```
This should crash the server process (nil-pointer panic in `fetchUrl`). Repeated requests with valid URLs can be checked for goroutine growth via `runtime.NumGoroutine()` or `net/http/pprof` to confirm the leak in the success path too (goroutines never unblock because nothing drains `respCh`).

## Priority fix order

1. `batch.go` — fix the nil `resp` panic, drain `respCh` (or eliminate it in favor of a mutex-guarded slice, or fix the unbuffered-channel consumer), and actually call `wg.Wait()` before responding.
2. Add method restrictions to `/fetch-batch`, `/fetch-channel`, `/enqueue`.
3. Fix `util.WriteLog` to append instead of truncate.
4. Add graceful shutdown (`http.Server` + `os/signal` + `Shutdown(ctx)`), including a stop path for `ProcessQueue`.
5. Add `go test -race` coverage for each endpoint, at minimum a concurrent-request test against `/fetch-batch` and `/enqueue`.
6. Cross-cutting cleanup: structured logging, `context.Context` propagation, de-duplicated constants, remove dead code (`internal/handler`, debug prints, unreachable `default` case).
