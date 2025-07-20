#!/bin/bash
# File permissions security checker for CloudMCP
# Ensures proper file permissions for security

set -euo pipefail

echo "🔒 CloudMCP File Permissions Security Check"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
ISSUES=0
WARNINGS=0
FIXED=0

# Function to print status
print_status() {
    local status=$1
    local message=$2
    case $status in
        "OK")
            echo -e "${GREEN}✓${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠${NC} $message"
            ((WARNINGS++))
            ;;
        "ERROR")
            echo -e "${RED}✗${NC} $message"
            ((ISSUES++))
            ;;
        "FIXED")
            echo -e "${GREEN}🔧${NC} $message"
            ((FIXED++))
            ;;
    esac
}

# Check executable scripts
echo
echo "Checking executable scripts..."
while IFS= read -r -d '' script; do
    if [[ -x "$script" ]]; then
        print_status "OK" "Script $script has executable permissions"
    else
        print_status "FIXED" "Making $script executable"
        chmod +x "$script"
    fi
done < <(find . -name "*.sh" -o -name "*.bash" -o -name "*.zsh" -print0)

# Check for world-writable files
echo
echo "Checking for world-writable files..."
if world_writable=$(find . -type f -perm -o+w 2>/dev/null); then
    if [[ -n "$world_writable" ]]; then
        while IFS= read -r file; do
            print_status "ERROR" "World-writable file found: $file"
            print_status "FIXED" "Removing world-write permission from $file"
            chmod o-w "$file"
        done <<< "$world_writable"
    else
        print_status "OK" "No world-writable files found"
    fi
else
    print_status "OK" "No world-writable files found"
fi

# Check for group-writable files that shouldn't be
echo
echo "Checking for unnecessarily group-writable files..."
if group_writable=$(find . -type f -perm -g+w ! -path "./.git/*" 2>/dev/null); then
    if [[ -n "$group_writable" ]]; then
        while IFS= read -r file; do
            # Skip files that legitimately need group write (like shared config files)
            if [[ "$file" =~ \.(sh|py|pl)$ ]]; then
                continue # Scripts may need group write in some environments
            fi
            print_status "WARN" "Group-writable file: $file (may be intentional)"
        done <<< "$group_writable"
    else
        print_status "OK" "No unnecessary group-writable files found"
    fi
else
    print_status "OK" "No unnecessary group-writable files found"
fi

# Check sensitive configuration files
echo
echo "Checking sensitive configuration files..."
sensitive_patterns=(
    ".claude/settings*.json"
    "*.env"
    "*.env.*"
    "*secret*"
    "*key*"
    "*token*"
    "*credential*"
    "*.pem"
    "*.key"
)

for pattern in "${sensitive_patterns[@]}"; do
    while IFS= read -r -d '' file; do
        perms=$(stat -f "%A" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
        if [[ "$perms" =~ ^6[0-7][0-7]$ ]]; then
            print_status "OK" "Sensitive file $file has secure permissions ($perms)"
        elif [[ "$perms" =~ ^64[0-7]$ ]]; then
            print_status "FIXED" "Securing sensitive file $file (was $perms, now 600)"
            chmod 600 "$file"
        else
            print_status "ERROR" "Sensitive file $file has insecure permissions: $perms"
            if [[ "$perms" != "600" ]]; then
                print_status "FIXED" "Setting secure permissions (600) on $file"
                chmod 600 "$file"
            fi
        fi
    done < <(find . -name "$pattern" -type f -print0 2>/dev/null)
done

# Check configuration files for reasonable permissions
echo
echo "Checking configuration file permissions..."
config_patterns=(
    "*.toml"
    "*.yaml"
    "*.yml"
    "*.json"
)

for pattern in "${config_patterns[@]}"; do
    while IFS= read -r -d '' file; do
        # Skip sensitive files already checked
        if [[ "$file" =~ \.claude/settings.*\.json$ ]]; then
            continue
        fi
        
        perms=$(stat -f "%A" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
        if [[ "$perms" =~ ^6[0-4][0-4]$ ]]; then
            print_status "OK" "Config file $file has appropriate permissions ($perms)"
        elif [[ "$perms" =~ ^[67][0-7][0-7]$ ]] && [[ ! "$perms" =~ ^[67][0-7][0-7]$ ]]; then
            print_status "WARN" "Config file $file may have overly permissive permissions: $perms"
        fi
    done < <(find . -name "$pattern" -type f ! -path "./.git/*" -print0 2>/dev/null)
done

# Check for executables without extensions
echo
echo "Checking executables without extensions..."
while IFS= read -r -d '' file; do
    if [[ -x "$file" ]] && [[ ! "$file" =~ \.[a-zA-Z]+$ ]]; then
        perms=$(stat -f "%A" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
        if [[ "$perms" =~ ^7[0-7][0-57]$ ]]; then
            print_status "OK" "Executable $file has appropriate permissions ($perms)"
        else
            print_status "WARN" "Executable $file permissions: $perms"
        fi
    fi
done < <(find . -type f ! -path "./.git/*" ! -path "./bin/*" ! -name "*.go" ! -name "*.md" ! -name "*.json" ! -name "*.yml" ! -name "*.yaml" ! -name "*.toml" -print0)

# Summary
echo
echo "============================================="
echo "📊 File Permissions Security Check Summary"
echo "============================================="
print_status "OK" "Issues fixed: $FIXED"
if [[ $WARNINGS -gt 0 ]]; then
    print_status "WARN" "Warnings: $WARNINGS"
fi
if [[ $ISSUES -gt 0 ]]; then
    print_status "ERROR" "Issues requiring attention: $ISSUES"
    exit 1
else
    print_status "OK" "All file permissions are secure!"
    exit 0
fi