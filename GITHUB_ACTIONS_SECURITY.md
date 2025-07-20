# GitHub Actions Security: SHA Pinning Update Strategy

## Overview

This document outlines the security strategy implemented for GitHub Actions in this repository and provides a comprehensive guide for maintaining SHA-pinned actions.

## Security Implementation

All GitHub Actions in this repository are pinned to specific commit SHA hashes instead of mutable version tags. This prevents supply chain attacks where action tags could be moved to point to malicious code.

### Current SHA-Pinned Actions

| Action | Version | SHA Hash | Usage |
|--------|---------|----------|-------|
| `actions/checkout` | v4.1.7 | `692973e3d937129bcbf40652eb9f2f61becf3332` | Code checkout across all workflows |
| `actions/setup-go` | v5.5.0 | `d35c59abb061a4a6fb18e82ac0862c26744d6ab5` | Go environment setup |
| `actions/cache` | v4.2.3 | `5a3ec84eff668545956fd18022155c47e93e2684` | Dependency caching |
| `actions/upload-artifact` | v4.6.2 | `ea165f8d65b6e75b540449e92b4886f43607fa02` | Artifact uploads |
| `codecov/codecov-action` | v4.6.0 | `b9fd7d16f6d7d1b5d2bec1a2887e65ceed900238` | Code coverage reporting |
| `golangci/golangci-lint-action` | v6.5.1 | `4696ba8babb6127d732c3c6dde519db15edab9ea` | Go linting |
| `actions/github-script` | v7.0.1 | `60a0d83039c74a4aee543508d2ffcb1c3799cdea` | GitHub API scripting |
| `softprops/action-gh-release` | v1 | `de2c0eb89ae2a093876385947365aca7b0e5f844` | GitHub release creation |

## Security Benefits

### Immutable References
- SHA hashes are immutable - once committed, they cannot be changed
- Prevents tag hijacking attacks where malicious actors could move tags to point to compromised code
- Ensures the exact same code runs every time the workflow executes

### Supply Chain Protection
- Eliminates the risk of upstream repository compromises affecting our workflows
- Provides audit trail for exactly which version of each action was used
- Enables security reviews of specific action versions before adoption

### Compliance Requirements
- Meets security best practices for CI/CD pipelines
- Satisfies supply chain security requirements for production systems
- Provides deterministic builds and deployments

## Update Strategy

### When to Update

1. **Security Vulnerabilities**: Immediately when security issues are discovered in pinned actions
2. **New Features**: When new action features are needed for development workflow improvements  
3. **Bug Fixes**: When critical bugs in current pinned versions affect workflow functionality
4. **Regular Maintenance**: Quarterly review of all pinned actions for updates

### Update Process

#### Step 1: Research New Versions
1. Visit the action's GitHub repository releases page
2. Review changelog and security notes for new versions
3. Identify the latest stable version that meets our requirements

#### Step 2: Find SHA Hash
For each action repository:
```bash
# Method 1: Check releases page
https://github.com/owner/action/releases/tag/vX.Y.Z

# Method 2: Use GitHub API
curl -s "https://api.github.com/repos/owner/action/git/refs/tags/vX.Y.Z" | jq -r '.object.sha'

# Method 3: Clone and check locally
git clone https://github.com/owner/action.git
cd action
git rev-parse vX.Y.Z
```

#### Step 3: Verify Authenticity
1. Confirm the SHA hash matches the official release
2. Check that the release was signed by the action maintainers
3. Review the action's source code for any suspicious changes

#### Step 4: Update Workflow Files
Replace the old SHA hash with the new one in all affected workflow files:
```yaml
# Before
uses: actions/checkout@OLD_SHA_HASH # vX.Y.Z

# After  
uses: actions/checkout@NEW_SHA_HASH # vX.Y.Z+1
```

#### Step 5: Test Changes
1. Create a test branch with the updated SHA hashes
2. Run all affected workflows to ensure they complete successfully
3. Verify that the new action versions work correctly with our workflow logic

#### Step 6: Deploy Updates
1. Create a pull request with the SHA updates
2. Review the changes for security implications
3. Merge after successful testing and review

## Automation Opportunities

### Dependabot Configuration
While Dependabot doesn't directly support SHA pinning, it can be configured to monitor action updates:

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    # Note: Manual SHA update still required
```

### Custom Update Script
A custom script could automate SHA hash lookups:

```bash
#!/bin/bash
# update-action-shas.sh
# Script to check for new SHA hashes for pinned actions

actions=(
  "actions/checkout:v4"
  "actions/setup-go:v5" 
  "actions/cache:v4"
  # Add other actions...
)

for action in "${actions[@]}"; do
  repo="${action%:*}"
  version="${action#*:}"
  
  echo "Checking $repo at $version..."
  # Add logic to fetch latest SHA for version
done
```

### Monitoring and Alerts
Set up monitoring for:
- New security advisories affecting pinned actions
- New releases of critical actions
- Vulnerability databases for action dependencies

## Security Considerations

### Review Process
- All SHA updates must be reviewed by at least one maintainer
- Security-critical actions require additional review
- Test workflows must pass before merging updates

### Emergency Updates
For critical security vulnerabilities:
1. Identify affected actions immediately
2. Find patched versions and corresponding SHA hashes
3. Update workflows directly on main branch if necessary
4. Document the emergency update rationale

### Rollback Strategy
If updated actions cause issues:
1. Revert to previous known-good SHA hashes
2. Investigate the root cause of the failure
3. Plan a proper fix or alternative approach
4. Re-test before re-deploying

## Maintenance Schedule

### Weekly
- Monitor security advisories for pinned actions
- Check for critical bug fixes in action repositories

### Monthly  
- Review action release notes for new versions
- Evaluate new features that might benefit our workflows

### Quarterly
- Complete audit of all pinned action versions
- Update to latest stable versions unless contraindicated
- Review and update this documentation as needed

## Tools and Resources

### SHA Hash Verification
- [GitHub API Documentation](https://docs.github.com/en/rest/git/refs)
- [git rev-parse documentation](https://git-scm.com/docs/git-rev-parse)
- [jq for JSON processing](https://stedolan.github.io/jq/)

### Security Resources
- [GitHub Security Advisories](https://github.com/advisories)
- [StepSecurity Action Advisor](https://app.stepsecurity.io/action-advisor)
- [NIST National Vulnerability Database](https://nvd.nist.gov/)

### GitHub Actions Documentation
- [GitHub Actions Security Best Practices](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
- [Using third-party actions](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions)

## Incident Response

### If a Pinned Action is Compromised
1. **Immediate Response**
   - Disable affected workflows temporarily
   - Assess the scope of potential compromise
   - Check workflow run logs for suspicious activity

2. **Investigation**
   - Review the compromised action's source code
   - Check if our pinned SHA was affected
   - Analyze any data that may have been exposed

3. **Recovery**
   - Update to a clean version of the action
   - Re-run any potentially affected workflows
   - Document lessons learned and improve process

## Contact Information

For questions about this security strategy or to report security concerns:
- Create an issue in this repository with the `security` label
- Contact maintainers directly for sensitive security matters
- Follow responsible disclosure practices for vulnerability reports

---

**Last Updated**: 2024-07-20  
**Next Review**: 2024-10-20  
**Document Version**: 1.0