# Roadmap

The goal is to turn CKAD Trainer from a "generate a batch, run a script" tool
into an interactive practice environment with live, per-task feedback.

## Phase 1 — Foundations & fixes ✅

- [x] `--model` no longer forces the OpenRouter model onto OpenAI; each provider gets its own default (`gpt-4o-mini` / `deepseek/deepseek-chat:free`)
- [x] One request per scenario instead of one huge request, so output is not truncated at `--max-tokens`
- [x] Parallel generation (`--concurrency`) with retries and exponential backoff on 429/5xx, malformed JSON and invalid scenarios
- [x] JSON mode (`response_format: json_object`) and a brace-aware JSON extractor
- [x] Truncated responses (`finish_reason: length`) are reported clearly
- [x] Re-running `generate` appends new scenarios instead of overwriting `scenario-01…`
- [x] One failed scenario no longer discards the whole batch
- [x] Domains rotate across scenarios, `--topic` targets one skill area, existing titles are passed to avoid duplicates

## Phase 2 — Per-task validation ✅

- [x] New `scenario.json` format: domain, difficulty, namespace, setup, cleanup, and tasks with hint, solution and checks
- [x] Each check is a read-only `kubectl` command run **without a shell** (no pipes, redirects or chaining; `delete`/`apply`/etc. are rejected)
- [x] Match modes: `equals`, `contains`, `regex`, `exit0`
- [x] Every scenario gets a dedicated namespace (`ckad-scenario-NN`)
- [x] `validate` prints ✓/✗ per task and expected vs. actual output for failures
- [x] Legacy `validate.sh` scenarios only run with `--allow-scripts`
- [x] `scenario.md` renders without hints or solutions (no spoilers)

## Phase 3 — Interactive `practice` mode

- [ ] `ckad-trainer practice [--scenario N]` full-screen TUI (Bubble Tea + Lip Gloss)
- [ ] Task list with live ✓/✗/○ status and a per-scenario countdown timer
- [ ] Watch mode: re-run checks every few seconds while the user works in another terminal
- [ ] Key bindings: `c` check, `h` hint, `s` solution, `r` reset, `n` next, `q` quit
- [ ] Run `setup` on start and `cleanup` on reset/next (with confirmation)
- [ ] Progressive hints: hint first, then solution, with a score penalty per reveal
- [ ] "Explain my failure": send failing check output + `kubectl get -o yaml` to the LLM for a short diagnosis

## Phase 4 — Exam mode & progress tracking

- [ ] `ckad-trainer exam`: 15–20 questions, 2-hour timer, weighted scoring, final report
- [ ] Persist results in `~/.ckad-trainer/progress.json`
- [ ] `ckad-trainer stats`: pass rate and average time per domain
- [ ] `generate --difficulty` and `--weak-areas` to target the domains failed most

## Phase 5 — Polish

- [ ] `ckad-trainer list` to browse generated scenarios with status
- [ ] Cluster preflight: kubectl reachable, current context, metrics-server (HPA), ingress controller
- [ ] Config file (`~/.config/ckad-trainer.yaml`) and an interactive `init` wizard
- [ ] Shell completion (`ckad-trainer completion bash|zsh|fish`)
- [ ] Spinner/progress bar during generation
- [ ] Embedded curated scenario bank (`go:embed`) so the trainer works offline without an API key
- [ ] CI: `go vet`, `go test`, release binaries
