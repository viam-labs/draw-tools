#!/bin/bash
set -e

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: 'jq' is not installed."
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

DRY_RUN=$1

echo ""
echo "📦 Uploading module to Viam"
echo ""

# Check if this is a dry-run
if [[ "$DRY_RUN" == "1" ]]; then
    echo "🔍 DRY RUN MODE - No actual uploads will be performed"
    echo ""
fi

# Read current version from .version
CURRENT_VERSION=$(jq -r '.current' .version)
echo "Current version: $CURRENT_VERSION"
echo ""

# Read platforms from meta.json
PLATFORMS=$(jq -r '.build.arch[]' meta.json)

if [[ "$DRY_RUN" == "1" ]]; then
    # Dry-run mode: just show what would be executed
    echo "Would upload to platforms:"
    for PLATFORM in $PLATFORMS; do
        echo "  - $PLATFORM"
        echo "    Command: viam module upload --version $CURRENT_VERSION --platform $PLATFORM ."
    done
    
    echo ""
    echo "=================================================="
    echo "✓ Dry run for version $CURRENT_VERSION completed successfully!"
    echo "=================================================="
    echo ""
    echo "No changes were made."
    echo ""
else
    # Actual upload mode
    echo "Uploading to platforms..."
    for PLATFORM in $PLATFORMS; do
        echo "  - $PLATFORM"
        if ! viam module upload --version "$CURRENT_VERSION" --platform "$PLATFORM" .; then
            echo ""
            echo "Error: Upload failed for platform $PLATFORM"
            exit 1
        fi
    done
    
    echo ""
    echo "=================================================="
    echo "✓ All uploads successful!"
    echo "=================================================="
    echo ""
    echo "Version $CURRENT_VERSION uploaded to all platforms."
    echo ""
fi
