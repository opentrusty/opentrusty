# Runbook: Automated E2E UI Testing

**Target Audience**: CI Systems, Developers (Regression Testing)
**Goal**: Verify full system functionality using the headless automated browser suite.

## 1. Prerequisites

* Docker running (for DB)
* Go 1.22+ installed
* Node.js 18+ installed

## 2. Infrastructure Setup (One-Shot)

Use the master orchestrator to reset the environment and start all services.

```bash
# From workspace root
cd /Users/mw/workspace/repo/github.com/opentrusty/opentrusty
./tests/local/run-all.sh --no-tests
```

*Wait usually ~10 seconds for services to be healthy.*

## 3. Execute Automated Suite

Switch to the Control Panel directory and run Playwright.

```bash
cd /Users/mw/workspace/repo/github.com/opentrusty/opentrusty-control-panel

# Install dependencies if consistent failures occur
npm ci 
npx playwright install chromium

# Run the suite
npx playwright test
```

## 4. Interpret Results

* **PASS**: All 5 specs (`01` to `05`) passed. System is Beta Ready.
* **FAIL**: Check the HTML report:
  ```bash
  npx playwright show-report
  ```

## 5. Cleanup

```bash
# Kill background processes from run-all.sh (CTRL+C)
# Or manually:
pkill opentrusty
pkill demo-app
docker compose -f tests/local/docker-compose.local-test.yml down -v
```
