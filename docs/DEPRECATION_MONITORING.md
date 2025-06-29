# Deprecation Monitoring Guide

This document outlines CloudMCP's automated deprecation monitoring system and how to respond to deprecation warnings.

## Overview

CloudMCP uses automated tools to monitor for deprecated APIs and dependencies to ensure code quality and prevent future breaking changes. The monitoring system runs continuously and alerts the team when deprecations are detected.

## Monitoring Components

### 1. Continuous Integration (CI) Monitoring

**Location:** `.github/workflows/ci.yml`
**Trigger:** Every push and pull request
**Purpose:** Immediate feedback on deprecation status

The CI pipeline includes a deprecation check step that:
- Scans code for SA1019 (deprecated API usage) warnings
- Reports findings as GitHub Actions warnings/notices
- Uploads deprecation reports as artifacts for investigation
- Provides immediate feedback to developers

### 2. Dedicated Deprecation Workflow

**Location:** `.github/workflows/deprecation-monitoring.yml`
**Triggers:**
- Manual runs
- Pull requests (with detailed reporting)
- Weekly schedule (Mondays at 9 AM UTC)

This workflow provides:
- Detailed deprecation analysis and reporting
- Automatic PR comments with deprecation status
- Weekly monitoring reports via GitHub Issues
- Historical tracking of deprecation trends

### 3. Pull Request Enforcement

**Behavior:** PRs with new deprecation warnings are automatically failed
**Purpose:** Prevent introduction of new deprecated API usage
**Override:** Can be bypassed with explicit team approval and migration plan

## Workflow Details

### CI Integration Check

```yaml
- name: Check for deprecation warnings
  run: |
    golangci-lint run --enable=staticcheck --disable-all --out-format=line-number 2>&1 | \
    grep -E "SA1019|deprecated" > deprecations.txt || true
    
    DEPRECATION_COUNT=$(wc -l < deprecations.txt || echo "0")
    
    if [ "$DEPRECATION_COUNT" -gt 0 ]; then
      echo "::warning::Found $DEPRECATION_COUNT deprecation warning(s)"
    else
      echo "::notice::No deprecation warnings found ✅"
    fi
```

### Weekly Monitoring

The system automatically creates GitHub Issues for weekly deprecation reports:
- **Title:** "Weekly Deprecation Report: X warning(s) found"
- **Labels:** `maintenance`, `deprecation`, `technical-debt`
- **Content:** Summary with links to detailed reports

## Responding to Deprecation Warnings

### 1. Immediate Response (PR Failures)

When a PR fails due to deprecation warnings:

1. **Review the warning details** in the CI logs or artifact
2. **Identify the deprecated API** and its replacement
3. **Check the migration plan** in `DEPRECATION_MIGRATION_PLAN.md`
4. **Update the code** to use current APIs
5. **Test the changes** thoroughly
6. **Re-run the CI** to verify fixes

### 2. Planned Migration (Weekly Reports)

For existing deprecations found in weekly scans:

1. **Assess priority** based on deprecation timeline
2. **Create migration tasks** in project management system
3. **Update migration plan** with new findings
4. **Schedule migration work** according to priorities
5. **Test in staging** before production deployment

### 3. Emergency Response (Critical Deprecations)

For deprecations with imminent removal deadlines:

1. **Create emergency migration task**
2. **Notify team leads** immediately
3. **Fast-track development** with dedicated resources
4. **Expedite testing** and deployment
5. **Monitor deployment** closely

## Migration Process

### Step 1: Analysis
```bash
# Run local deprecation scan
golangci-lint run --enable=staticcheck --disable-all --out-format=line-number 2>&1 | \
grep -E "SA1019|deprecated"

# Check for API documentation
go doc <deprecated-package>.<deprecated-function>
```

### Step 2: Planning
- Check official documentation for replacement APIs
- Review impact on existing functionality
- Plan backwards compatibility if needed
- Estimate testing requirements

### Step 3: Implementation
- Update code to use current APIs
- Run local tests: `go test ./...`
- Run integration tests if available
- Update documentation and comments

### Step 4: Validation
- Verify all tests pass
- Check that deprecation warnings are resolved
- Test functionality in development environment
- Review code changes with team

## Maintenance Schedule

### Daily
- CI automatically checks all PRs and pushes
- Immediate feedback on new deprecations

### Weekly
- Automated scan creates GitHub Issues for tracking
- Team reviews deprecation status in weekly meetings

### Monthly
- Review dependency updates for new deprecations
- Update deprecation monitoring tools if needed
- Assess migration plan progress

### Quarterly
- Comprehensive review of deprecation monitoring effectiveness
- Update monitoring tools and processes
- Plan migrations for upcoming deprecation deadlines

## Tools and Commands

### Local Deprecation Scanning

```bash
# Quick deprecation check
golangci-lint run --enable=staticcheck --disable-all | grep -i deprecated

# Detailed analysis
golangci-lint run --enable=staticcheck --disable-all --out-format=line-number

# Check specific packages
golangci-lint run --enable=staticcheck --disable-all ./internal/services/linode/
```

### Dependency Analysis

```bash
# Check for outdated dependencies
go list -u -m all

# Update to latest versions (use with caution)
go get -u ./...
go mod tidy
```

### Testing After Migration

```bash
# Run all tests
go test -race ./...

# Run integration tests (if available)
go test -tags=integration -race ./internal/services/linode/...

# Check test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration

### GitHub Actions Settings

The workflows require no special configuration but can be customized:

- **Schedule frequency:** Modify cron expression in `deprecation-monitoring.yml`
- **Go version:** Update matrix in workflow files
- **Linter version:** Update golangci-lint version
- **Report retention:** Adjust artifact retention days

### Notification Settings

- **Weekly reports:** Automatically create GitHub Issues
- **PR comments:** Automatic comments on deprecation status
- **CI warnings:** Shown in GitHub Actions UI and logs

## Troubleshooting

### False Positives

If legitimate usage is flagged as deprecated:

1. Verify the API is actually deprecated in upstream documentation
2. Check if there's a newer version of the dependency
3. If confirmed false positive, add to exclusion list
4. Document the decision in migration plan

### Missing Deprecations

If known deprecations aren't detected:

1. Verify golangci-lint configuration
2. Check if staticcheck is enabled
3. Update linter to latest version
4. Review grep patterns for SA1019 detection

### Workflow Failures

If monitoring workflows fail:

1. Check GitHub Actions logs for specific errors
2. Verify Go version compatibility
3. Check dependency download issues
4. Review golangci-lint installation steps

## References

- [CloudMCP Deprecation Migration Plan](../DEPRECATION_MIGRATION_PLAN.md)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [Go SA1019 Static Analysis](https://staticcheck.io/docs/checks#SA1019)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)