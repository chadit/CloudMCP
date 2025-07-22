#!/bin/bash

# CloudMCP Branch Protection Setup Script
# Sets up comprehensive branch protection rules for the two-phase CI/CD system
# Usage: ./scripts/setup-branch-protection.sh [--repo REPO] [--dry-run]

set -euo pipefail

# Configuration
REPO="${REPO:-chadit/CloudMCP}"
DRY_RUN=false

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Usage function
show_usage() {
    cat << EOF
CloudMCP Branch Protection Setup Script

USAGE:
    $0 [--repo REPO] [--dry-run]

OPTIONS:
    --repo REPO     Repository in format owner/repo (default: chadit/CloudMCP)
    --dry-run       Show what would be done without making changes
    --help          Show this help message

EXAMPLES:
    # Setup branch protection for CloudMCP
    $0

    # Setup for different repository
    $0 --repo myorg/myrepo

    # Preview changes without applying
    $0 --dry-run

REQUIREMENTS:
    - gh CLI tool must be installed and authenticated
    - User must have admin access to the repository
EOF
}

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
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v gh >/dev/null 2>&1; then
        log_error "GitHub CLI (gh) is not installed. Please install it first."
        log_info "Install with: brew install gh  # or appropriate method for your OS"
        exit 1
    fi
    
    if ! gh auth status >/dev/null 2>&1; then
        log_error "GitHub CLI is not authenticated. Please run 'gh auth login' first."
        exit 1
    fi
    
    # Check if repo exists and we have access
    if ! gh repo view "$REPO" >/dev/null 2>&1; then
        log_error "Cannot access repository: $REPO"
        log_info "Please check repository name and your permissions."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Apply branch protection rule
apply_branch_protection() {
    local branch_pattern="$1"
    local description="$2"
    
    log_info "Setting up branch protection for: $branch_pattern ($description)"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "[DRY-RUN] Would apply branch protection to: $branch_pattern"
        return 0
    fi
    
    # Create comprehensive branch protection rule using GitHub API via gh CLI
    # Use --input to pass JSON data properly  
    # Simplified for personal repositories (no team/user restrictions)
    cat <<EOF | gh api \
        --method PUT \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "/repos/$REPO/branches/$branch_pattern/protection" \
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
            "CodeQL Analysis"
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
}

# Setup using GitHub Rulesets (2024 approach)
setup_rulesets() {
    log_info "Setting up GitHub Rulesets for comprehensive protection..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "[DRY-RUN] Would create GitHub Ruleset"
        return 0
    fi
    
    # Create ruleset using the modern GitHub Rulesets API
    # Use --input to pass JSON data properly
    cat <<EOF | gh api \
        --method POST \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "/repos/$REPO/rulesets" \
        --input -
{
    "name": "CloudMCP Two-Phase CI/CD Protection",
    "target": "branch",
    "enforcement": "active",
    "conditions": {
        "ref_name": {
            "include": ["~DEFAULT_BRANCH", "refs/heads/main", "refs/heads/develop"],
            "exclude": ["refs/heads/temp/*", "refs/heads/experimental/*"]
        }
    },
    "rules": [
        {
            "type": "deletion"
        },
        {
            "type": "non_fast_forward"
        },
        {
            "type": "required_status_checks",
            "parameters": {
                "required_status_checks": [
                    {
                        "context": "Fast Quality Checks"
                    },
                    {
                        "context": "Comprehensive Tests (1.22, unit)"
                    },
                    {
                        "context": "Comprehensive Tests (1.22, integration)"
                    },
                    {
                        "context": "Comprehensive Tests (1.22, race)"
                    },
                    {
                        "context": "Comprehensive Tests (1.23, unit)"
                    },
                    {
                        "context": "Comprehensive Tests (1.23, integration)"
                    },
                    {
                        "context": "Comprehensive Tests (1.23, race)"
                    },
                    {
                        "context": "Security Analysis & SBOM"
                    },
                    {
                        "context": "Build Testing (linux, amd64, 1)"
                    },
                    {
                        "context": "Build Testing (linux, arm64, 0)"
                    },
                    {
                        "context": "Build Testing (darwin, amd64, 0)"
                    },
                    {
                        "context": "Build Testing (darwin, arm64, 0)"
                    },
                    {
                        "context": "Build Testing (windows, amd64, 0)"
                    },
                    {
                        "context": "Container Integration"
                    },
                    {
                        "context": "CodeQL Analysis"
                    }
                ],
                "strict_required_status_checks_policy": true
            }
        },
        {
            "type": "pull_request",
            "parameters": {
                "required_approving_review_count": 1,
                "dismiss_stale_reviews_on_push": true,
                "require_code_owner_review": false,
                "require_last_push_approval": true,
                "required_review_thread_resolution": true
            }
        },
        {
            "type": "required_deployments",
            "parameters": {
                "required_deployment_environments": []
            }
        }
    ],
    "bypass_actors": []
}
EOF
}

# Main execution
main() {
    log_info "CloudMCP Branch Protection Setup"
    log_info "Repository: $REPO"
    log_info "Mode: $([ "$DRY_RUN" == "true" ] && echo "DRY-RUN" || echo "APPLY")"
    echo

    check_prerequisites
    
    log_info "Setting up 2024 GitHub branch protection best practices..."
    
    # Method 1: Traditional Branch Protection Rules (for compatibility)
    log_info "Applying traditional branch protection rules..."
    apply_branch_protection "main" "Main branch protection"
    
    # Method 2: Modern GitHub Rulesets (2024 approach)
    log_info "Setting up modern GitHub Rulesets..."
    setup_rulesets
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo
        log_warning "DRY-RUN completed. No changes were made."
        log_info "To apply these changes, run the script without --dry-run"
    else
        echo
        log_success "Branch protection setup completed successfully!"
        echo
        log_info "Protection features enabled:"
        echo "  ✅ Two-phase CI/CD status checks required"
        echo "  ✅ Pull request reviews required (1 approver)"
        echo "  ✅ Dismiss stale reviews on push"
        echo "  ✅ Require last push approval"
        echo "  ✅ Require conversation resolution"
        echo "  ✅ Prevent force pushes and deletions"
        echo "  ✅ Up-to-date branch requirements (strict)"
        echo "  ✅ Admin enforcement enabled"
        echo
        log_info "Emergency bypass available for repository admins via pull request"
    fi
}

# Run main function
main "$@"