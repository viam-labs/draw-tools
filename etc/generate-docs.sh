#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "📚 Generating documentation"
echo ""

# Check if gomarkdoc is installed
if ! command -v ~/go/bin/gomarkdoc &> /dev/null; then
    echo -e "${YELLOW}gomarkdoc not found, installing...${NC}"
    go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
    echo -e "${GREEN}✓ gomarkdoc installed${NC}"
else
    echo -e "${GREEN}✓ gomarkdoc found${NC}"
fi

# Generate documentation for lib package
echo "Generating documentation for lib package..."
~/go/bin/gomarkdoc -o DOCS.md ./lib

if [ -f "DOCS.md" ]; then
    echo -e "${GREEN}✓ Documentation generated: DOCS.md${NC}"
else
    echo -e "${RED}✗ Failed to generate documentation${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✓ Documentation generation completed!${NC}"
echo ""
