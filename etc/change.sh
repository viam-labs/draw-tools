#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "📝 Creating a new changelog entry"
echo ""

# Prompt for change type
echo "Select the type of change:"
echo "  1) patch - Bug fixes, minor changes"
echo "  2) minor - New features, backwards compatible"
echo "  3) major - Breaking changes"
echo ""
read -p "Enter choice (1-3): " choice

case "$choice" in
    1)
        CHANGE_TYPE="patch"
        ;;
    2)
        CHANGE_TYPE="minor"
        ;;
    3)
        CHANGE_TYPE="major"
        ;;
    *)
        echo -e "${RED}Error: Invalid choice. Please select 1, 2, or 3.${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "Selected: ${GREEN}${CHANGE_TYPE}${NC}"
echo ""

# Prompt for change message
echo "Enter a description of your change:"
read -p "> " CHANGE_MESSAGE

if [[ -z "$CHANGE_MESSAGE" ]]; then
    echo -e "${RED}Error: Change message cannot be empty.${NC}"
    exit 1
fi

# Get git username
GIT_USER=$(git config user.name)
if [[ -z "$GIT_USER" ]]; then
    echo -e "${RED}Error: Git user.name not configured.${NC}"
    echo ""
    echo "Please configure your git username:"
    echo -e "  ${BLUE}git config user.name \"Your Name\"${NC}"
    echo ""
    exit 1
fi

# Get current timestamp
TIMESTAMP=$(date "+%Y%m%d%H%M%S")
TIMESTAMP_READABLE=$(date "+%Y-%m-%d %H:%M:%S %Z")

# Create .changelogs directory if it doesn't exist
mkdir -p .changelogs

# Generate filename based on user, type, and timestamp
# Sanitize username for filename (replace spaces and special chars with hyphens)
SAFE_USER=$(echo "$GIT_USER" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//' | sed 's/-$//')
CHANGELOG_FILE=".changelogs/${SAFE_USER}-${CHANGE_TYPE}-${TIMESTAMP}.json"

# Create the changelog JSON file
cat > "$CHANGELOG_FILE" << EOF
{
  "type": "$CHANGE_TYPE",
  "message": "$CHANGE_MESSAGE",
  "by": "$GIT_USER",
  "at": "$TIMESTAMP_READABLE"
}
EOF

echo ""
echo -e "${GREEN}✓ Changelog created successfully!${NC}"
echo ""
echo "File: $CHANGELOG_FILE"
echo "Type: $CHANGE_TYPE"
echo "Author: $GIT_USER"
echo "Time: $TIMESTAMP_READABLE"
echo "Message: $CHANGE_MESSAGE"
echo ""
echo "Remember to commit this changelog file with your changes:"
echo -e "  ${BLUE}git add $CHANGELOG_FILE${NC}"
echo ""

