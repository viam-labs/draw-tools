#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: 'jq' is not installed.${NC}"
    echo ""
    echo "Please install 'jq' to continue:"
    echo ""
    
    case "$(uname -s)" in
        Linux)
            echo "For Linux, use your package manager:"
            echo "  • Debian/Ubuntu:  sudo apt-get install jq"
            echo "  • Fedora/RHEL:    sudo dnf install jq"
            echo "  • openSUSE:       sudo zypper install jq"
            echo "  • Arch Linux:     sudo pacman -S jq"
            echo ""
            echo "Or download binaries from: https://jqlang.org/download/"
            ;;
        Darwin)
            echo "For macOS, use Homebrew or MacPorts:"
            echo "  • Homebrew:  brew install jq"
            echo "  • MacPorts:  sudo port install jq"
            echo ""
            echo "Or download binaries from: https://jqlang.org/download/"
            ;;
        CYGWIN*|MINGW*|MSYS*|MINGW32*|MINGW64*)
            echo "For Windows, use a package manager:"
            echo "  • winget:      winget install jqlang.jq"
            echo "  • Chocolatey:  choco install jq"
            echo "  • Scoop:       scoop install jq"
            echo ""
            echo "Or download binaries from: https://jqlang.org/download/"
            ;;
        *)
            echo "Please visit: https://jqlang.org/download/"
            ;;
    esac
    echo ""
    exit 1
fi

echo ""
echo "🚀 Bumping version based on changelogs"
echo ""

# Check if there are any changelog files
if [ ! -d ".changelogs" ] || [ -z "$(ls -A .changelogs 2>/dev/null)" ]; then
    echo -e "${RED}Error: No changelog files found in .changelogs/${NC}"
    echo ""
    echo "Please create a changelog first using:"
    echo -e "  ${BLUE}make change${NC}"
    echo ""
    exit 1
fi

# Check current branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${RED}Error: You must be on the 'main' branch to bump version.${NC}"
    echo -e "Current branch: ${YELLOW}$CURRENT_BRANCH${NC}"
    echo ""
    echo "Please checkout main:"
    echo -e "  ${BLUE}git checkout main${NC}"
    echo ""
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: You have uncommitted changes.${NC}"
    echo ""
    echo "Please commit or stash your changes before bumping version:"
    echo -e "  ${BLUE}git commit -am \"Your message\"${NC}"
    echo "or"
    echo -e "  ${BLUE}git stash${NC}"
    echo ""
    exit 1
fi

# Generate documentation
echo "Generating documentation..."
make docs
echo ""

# Read current version
CURRENT_VERSION=$(jq -r '.current' .version)
echo -e "Current version: ${BLUE}$CURRENT_VERSION${NC}"

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Determine the highest change type from all changelog files
HIGHEST_TYPE="patch"
for changelog_file in .changelogs/*.json; do
    CHANGE_TYPE=$(jq -r '.type' "$changelog_file")
    
    case "$CHANGE_TYPE" in
        major)
            HIGHEST_TYPE="major"
            ;;
        minor)
            if [ "$HIGHEST_TYPE" != "major" ]; then
                HIGHEST_TYPE="minor"
            fi
            ;;
    esac
done

echo -e "Highest change type: ${GREEN}$HIGHEST_TYPE${NC}"

# Calculate next version
case "$HIGHEST_TYPE" in
    patch)
        PATCH=$((PATCH + 1))
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
esac

NEXT_VERSION="$MAJOR.$MINOR.$PATCH"
echo -e "Next version: ${GREEN}$NEXT_VERSION${NC}"
echo ""

# Create new branch
BRANCH_NAME="version-$NEXT_VERSION"
echo "Creating branch: $BRANCH_NAME"
git checkout -b "$BRANCH_NAME"
echo ""

# Get git user info for metadata
GIT_USER=$(git config user.name)
GIT_EMAIL=$(git config user.email)
TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S %Z")

# Prepare CHANGELOG.md entry
echo "Updating CHANGELOG.md..."

# Create CHANGELOG.md if it doesn't exist
if [ ! -f "CHANGELOG.md" ]; then
    cat > CHANGELOG.md << EOF
# Changelog

Current Version: $NEXT_VERSION

---

EOF
fi

# Prepare changelog entry (will be written to temp file below)

# Update existing changelog
# First, update the Current Version line
sed -i.bak "s/^Current Version: .*/Current Version: $NEXT_VERSION/" CHANGELOG.md && rm -f CHANGELOG.md.bak

# Create a temporary file with the new changelog entry
cat > /tmp/new_changelog_entry.txt << EOF

## $CURRENT_VERSION → $NEXT_VERSION

*Released by $GIT_USER ($GIT_EMAIL) on $TIMESTAMP*

EOF

# Add changes grouped by type
for change_type in major minor patch; do
    for changelog_file in .changelogs/*.json; do
        if [ ! -f "$changelog_file" ]; then
            continue
        fi
        
        TYPE=$(jq -r '.type' "$changelog_file")
        MESSAGE=$(jq -r '.message' "$changelog_file")
        BY=$(jq -r '.by // "Unknown"' "$changelog_file")
        AT=$(jq -r '.at // "Unknown"' "$changelog_file")
        
        if [ "$TYPE" = "$change_type" ]; then
            echo "- **[$TYPE]** $MESSAGE *(by $BY at $AT)*" >> /tmp/new_changelog_entry.txt
        fi
    done
done

# Insert the new entry after the "---" line
awk '
    /^---$/ {
        print
        while ((getline line < "/tmp/new_changelog_entry.txt") > 0) {
            print line
        }
        close("/tmp/new_changelog_entry.txt")
        next
    }
    { print }
' CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md

# Clean up temporary file
rm -f /tmp/new_changelog_entry.txt

echo -e "${GREEN}✓ CHANGELOG.md updated${NC}"

# Update .version file
jq --arg version "$NEXT_VERSION" '.current = $version' .version > .version.tmp && mv .version.tmp .version
echo -e "${GREEN}✓ .version updated${NC}"

# Remove all changelog files
rm -f .changelogs/*.json
echo -e "${GREEN}✓ Changelog files cleared${NC}"
echo ""

# Stage changes (including deleted changelog files)
git add CHANGELOG.md .version DOCS.md
git add -A .changelogs/
echo "Changes staged"

# Create commit
COMMIT_MESSAGE="$CURRENT_VERSION → $NEXT_VERSION"
git commit -m "$COMMIT_MESSAGE"
echo -e "${GREEN}✓ Committed: $COMMIT_MESSAGE${NC}"
echo ""

# Show summary
echo "=================================================="
echo -e "${GREEN}✓ Version bump completed successfully!${NC}"
echo "=================================================="
echo ""
echo "Summary:"
echo "  • Version: $CURRENT_VERSION → $NEXT_VERSION"
echo "  • Branch: $BRANCH_NAME"
echo "  • Commit: $COMMIT_MESSAGE"
echo ""

# Prompt to push
echo -e "${YELLOW}Ready to push the branch?${NC}"
echo ""
read -p "Push now? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "Pushing branch..."
    git push origin "$BRANCH_NAME"
    echo ""
    echo -e "${GREEN}✓ Branch pushed successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Create a pull request to merge $BRANCH_NAME into main"
    echo "  2. Once the PR is approved and merged to main:"
    echo "     → A GitHub Action will automatically create the tag: $NEXT_VERSION"
    echo "     → The tag will trigger the module build and upload"
else
    echo ""
    echo "Not pushing. To push later, run:"
    echo -e "  ${BLUE}git push origin $BRANCH_NAME${NC}"
    echo ""
    echo "After pushing, create a PR to merge into main."
    echo "The tag will be created automatically when the PR is merged."
fi

echo ""

