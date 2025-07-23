#!/bin/bash

# CloudMCP Branch Protection Setup Script  
# PREVENTS AI FROM MERGING while allowing owner control
# Sets up comprehensive branch protection rules for the two-phase CI/CD system

set -euo pipefail

# Configuration
REPO="${REPO:-chadit/CloudMCP}"
DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --repo)
            REPO="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Usage: $0 [--repo owner/repo] [--dry-run]"
            exit 1
            ;;
    esac
done

echo "🛡️ Setting up AI-proof branch protection for: $REPO"
echo "Mode: $([ "$DRY_RUN" == "true" ] && echo "DRY-RUN" || echo "APPLY")"

# Check prerequisites
if ! command -v gh >/dev/null 2>&1; then
    echo "❌ GitHub CLI (gh) is not installed"
    exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
    echo "❌ GitHub CLI is not authenticated. Run 'gh auth login' first."
    exit 1
fi

if [[ "$DRY_RUN" == "true" ]]; then
    echo "🔍 DRY-RUN: Would apply AI-proof branch protection"
    exit 0
fi

# Apply AI-proof branch protection rule
echo "📋 Applying AI-proof branch protection..."

cat <<EOF | gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "/repos/$REPO/branches/main/protection" \
    --input -
{
    "required_status_checks": {
        "strict": true,
        "contexts": [
            "Fast Quality Checks",
            "Comprehensive Tests (1.22, unit)",
            "Comprehensive Tests (1.22, integration)",
            "Comprehensive Tests (1.22, race)",
            "Comprehensive Tests (1.23, unit)",
            "Comprehensive Tests (1.23, integration)",
            "Comprehensive Tests (1.23, race)",
            "Security Analysis & SBOM",
            "Build Testing (linux, amd64, 1)",
            "Build Testing (linux, arm64, 0)",
            "Build Testing (darwin, amd64, 0)",
            "Build Testing (darwin, arm64, 0)",
            "Build Testing (windows, amd64, 0)",
            "Container Integration",
            "CodeQL Analysis (go)"
        ]
    },
    "enforce_admins": true,
    "required_pull_request_reviews": {
        "required_approving_review_count": 1,
        "dismiss_stale_reviews": true,
        "require_code_owner_reviews": false,
        "require_last_push_approval": true
    },
    "restrictions": null,
    "required_linear_history": false,
    "allow_force_pushes": false,
    "allow_deletions": false,
    "block_creations": false,
    "required_conversation_resolution": true,
    "lock_branch": false,
    "allow_fork_syncing": true
}
EOF

echo "
✅ AI-PROOF Branch Protection Applied Successfully!

🛡️ Security Measures Active:
  ✅ 1 Required Approver (prevents AI self-approval)
  ✅ Last Push Approval Required (AI can't approve after push)
  ✅ Stale Review Dismissal (new commits need fresh approval)
  ✅ All 15 Status Checks Required (full CI/CD testing)
  ✅ Admin Enforcement (even owner follows rules)
  ✅ No Force Pushes/Deletions

🤖 How This Prevents AI Merging:
  ❌ AI cannot approve its own PRs (GitHub blocks self-approval)
  ❌ AI commits dismiss existing approvals (fresh approval needed)
  ❌ Only HUMAN owner can provide required approval
  ❌ No CLI bypass options (--admin blocked)

👤 Human Owner Can Still:
  ✅ Create and approve PRs through GitHub web UI
  ✅ Merge after all 15 status checks pass
  ✅ Emergency override via GitHub web interface if needed

🎯 Perfect for AI-assisted development with human oversight!
"