# func-e configuration

The [XDG Base Directory Specification][xdg] defines standard locations for
user-specific files:

- **Data files**: Persist across sessions, e.g. downloaded binaries
- **State files**: Persist between restarts but non-portable, e.g. stdout.log
- **Runtime files**: Ephemeral files, e.g. admin-address.txt

func-e adopts these conventions to separate downloaded Envoy binaries, logs,
and ephemeral admin addresses. Doing so allows library consumers like Envoy AI
Gateway to define their own home directories under an XDG base convention.

## Configuration mappings

| Environment Variable  | Default Path            | API Option          |
|-----------------------|-------------------------|---------------------|
| `FUNC_E_CONFIG_HOME`  | `~/.config/func-e`      | `api.ConfigHome()`  |
| `FUNC_E_DATA_HOME`    | `~/.local/share/func-e` | `api.DataHome()`    |
| `FUNC_E_STATE_HOME`   | `~/.local/state/func-e` | `api.StateHome()`   |
| `FUNC_E_RUNTIME_DIR`  | `/tmp/func-e-${UID}`    | `api.RuntimeDir()`  |
| `FUNC_E_RUN_ID`       | auto-generated          | `api.RunID()`       |

| File Type              | Purpose                                      | Default Path                                                                     |
|------------------------|----------------------------------------------|----------------------------------------------------------------------------------|
| Selected Envoy Version | Version preference (persistent, shared)      | `${FUNC_E_CONFIG_HOME}/envoy-version`                                            |
| Envoy Binaries         | Downloaded executables (persistent, shared)  | `${FUNC_E_DATA_HOME}/envoy-versions/{version}/bin/envoy`                         |
| Envoy Run State        | Per-run logs & config (persistent debugging) | `${FUNC_E_STATE_HOME}/envoy-runs/{runID}/stdout.log,stderr.log,config_dump.json` |
| Admin Address Default  | Generated endpoint (ephemeral, per-run)      | `${FUNC_E_RUNTIME_DIR}/{runID}/admin-address.txt`                                |

- **Correlation ID (`runID`)**: `YYYYMMDD_HHMMSS_UUU`
  - (epoch date, time, last 3 digits of micros to allow concurrent runs)
  - Can be customized via `FUNC_E_RUN_ID` environment variable or `--run-id` flag
  - Custom runID cannot contain path separators (/ or \)
  - Example: `--run-id 0` for predictable Docker/Kubernetes deployments
- **Directory per run** isolates concurrent runs and ensures correlation

## Legacy Mapping

**Deprecation Warning**:
```
WARNING: FUNC_E_HOME is deprecated and will be removed in a future version.
Please migrate to FUNC_E_CONFIG_HOME, FUNC_E_DATA_HOME, FUNC_E_STATE_HOME or FUNC_E_RUNTIME_DIR.
```

| File Type              | Legacy Path Pattern                           | XDG Path Pattern                                            |
|------------------------|-----------------------------------------------|-------------------------------------------------------------|
| Selected Envoy Version | `$FUNC_E_HOME/version`                        | `$FUNC_E_CONFIG_HOME/envoy-version`                         |
| Envoy Binaries         | `$FUNC_E_HOME/versions/{version}/bin/envoy`   | `$FUNC_E_DATA_HOME/envoy-versions/{version}/bin/envoy`      |
| Run Logs               | `$FUNC_E_HOME/runs/{epoch}/stdout.log`        | `$FUNC_E_STATE_HOME/envoy-runs/{runID}/stdout.log`          |
| Admin Address          | `$FUNC_E_HOME/runs/{epoch}/admin-address.txt` | `$FUNC_E_RUNTIME_DIR/{runID}/admin-address.txt`             |

These legacy patterns will be supported only when `FUNC_E_HOME` is set and will
be removed in a future version. A file envoy.pid will not be written as it
isn't necessary.

---
[xdg]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
