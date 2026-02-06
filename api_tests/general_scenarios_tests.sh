#!/bin/bash

# Script for testing AIService API general scenarios
# This script contains curl commands for testing general structurization and summarization scenarios

set -e  # Exit on any error

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
TIMEOUT=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to test simple summarize scenario
test_simple_summarize() {
    print_status "Testing simple Summarize API scenario..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/summarize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-simple-sum-001",
            "userId": "user-simple-123",
            "requestType": "summarize",
            "board": {
                "boardId": "simple-board-001",
                "imageUrl": "",
                "elements": [
                    {
                        "id": "elem-title",
                        "type": "text",
                        "x": 50.0,
                        "y": 50.0,
                        "width": 200.0,
                        "height": 40.0,
                        "rotation": 0.0,
                        "content": "Project Overview"
                    },
                    {
                        "id": "elem-desc",
                        "type": "text",
                        "x": 50.0,
                        "y": 100.0,
                        "width": 300.0,
                        "height": 80.0,
                        "rotation": 0.0,
                        "content": "This is a simple project with basic components and functionality."
                    },
                    {
                        "id": "elem-note",
                        "type": "text",
                        "x": 50.0,
                        "y": 200.0,
                        "width": 250.0,
                        "height": 60.0,
                        "rotation": 0.0,
                        "content": "Additional notes about the project"
                    }
                ]
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 || $HTTP_CODE -eq 202 ]]; then
        print_success "Simple Summarize API scenario completed successfully (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Simple Summarize API scenario failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test complex summarize scenario
test_complex_summarize() {
    print_status "Testing complex Summarize API scenario..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/summarize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-complex-sum-001",
            "userId": "user-complex-456",
            "requestType": "summarize",
            "board": {
                "boardId": "complex-board-001",
                "imageUrl": "",
                "elements": [
                    {
                        "id": "header-1",
                        "type": "text",
                        "x": 10.0,
                        "y": 10.0,
                        "width": 400.0,
                        "height": 50.0,
                        "rotation": 0.0,
                        "content": "System Architecture Overview"
                    },
                    {
                        "id": "frontend-box",
                        "type": "rectangle",
                        "x": 50.0,
                        "y": 80.0,
                        "width": 150.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#FFE4B5",
                        "stroke": "#8B4513",
                        "strokeWidth": 2
                    },
                    {
                        "id": "frontend-label",
                        "type": "text",
                        "x": 75.0,
                        "y": 115.0,
                        "width": 100.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "content": "Frontend"
                    },
                    {
                        "id": "backend-box",
                        "type": "rectangle",
                        "x": 250.0,
                        "y": 80.0,
                        "width": 150.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#E6F3FF",
                        "stroke": "#0066CC",
                        "strokeWidth": 2
                    },
                    {
                        "id": "backend-label",
                        "type": "text",
                        "x": 275.0,
                        "y": 115.0,
                        "width": 100.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "content": "Backend"
                    },
                    {
                        "id": "db-box",
                        "type": "rectangle",
                        "x": 150.0,
                        "y": 220.0,
                        "width": 200.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#F0FFF0",
                        "stroke": "#228B22",
                        "strokeWidth": 2
                    },
                    {
                        "id": "db-label",
                        "type": "text",
                        "x": 200.0,
                        "y": 255.0,
                        "width": 100.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "content": "Database"
                    },
                    {
                        "id": "connection-1",
                        "type": "line",
                        "x": 125.0,
                        "y": 180.0,
                        "width": 0.0,
                        "height": 40.0,
                        "rotation": 0.0,
                        "stroke": "#000000",
                        "strokeWidth": 2,
                        "points": [125.0, 180.0, 125.0, 220.0]
                    },
                    {
                        "id": "connection-2",
                        "type": "line",
                        "x": 325.0,
                        "y": 180.0,
                        "width": 0.0,
                        "height": 40.0,
                        "rotation": 0.0,
                        "stroke": "#000000",
                        "strokeWidth": 2,
                        "points": [325.0, 180.0, 325.0, 220.0]
                    },
                    {
                        "id": "note-1",
                        "type": "text",
                        "x": 100.0,
                        "y": 340.0,
                        "width": 300.0,
                        "height": 60.0,
                        "rotation": 0.0,
                        "content": "This architecture shows a typical three-tier application with clear separation of concerns."
                    }
                ]
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 || $HTTP_CODE -eq 202 ]]; then
        print_success "Complex Summarize API scenario completed successfully (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Complex Summarize API scenario failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test simple structurize scenario
test_simple_structurize() {
    print_status "Testing simple Structurize API scenario..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/structurize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-simple-struct-001",
            "userId": "user-simple-789",
            "requestType": "structurize",
            "board": {
                "boardId": "simple-struct-board-001",
                "imageUrl": "",
                "elements": [
                    {
                        "id": "elem-title",
                        "type": "text",
                        "x": 50.0,
                        "y": 50.0,
                        "width": 200.0,
                        "height": 40.0,
                        "rotation": 0.0,
                        "content": "Simple Project Structure"
                    },
                    {
                        "id": "elem-desc",
                        "type": "text",
                        "x": 50.0,
                        "y": 100.0,
                        "width": 300.0,
                        "height": 80.0,
                        "rotation": 0.0,
                        "content": "Basic project with minimal components"
                    }
                ]
            },
            "file": {
                "name": "simple-project",
                "type": "doc",
                "children": [
                    {
                        "name": "main.js",
                        "type": "doc"
                    },
                    {
                        "name": "README.md",
                        "type": "doc"
                    }
                ]
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 || $HTTP_CODE -eq 202 ]]; then
        print_success "Simple Structurize API scenario completed successfully (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Simple Structurize API scenario failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test complex structurize scenario
test_complex_structurize() {
    print_status "Testing complex Structurize API scenario..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/structurize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-complex-struct-001",
            "userId": "user-complex-012",
            "requestType": "structurize",
            "board": {
                "boardId": "complex-struct-board-001",
                "imageUrl": "",
                "elements": [
                    {
                        "id": "header-1",
                        "type": "text",
                        "x": 10.0,
                        "y": 10.0,
                        "width": 400.0,
                        "height": 50.0,
                        "rotation": 0.0,
                        "content": "Complex System Structure"
                    },
                    {
                        "id": "frontend-box",
                        "type": "rectangle",
                        "x": 50.0,
                        "y": 80.0,
                        "width": 150.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#FFE4B5",
                        "stroke": "#8B4513",
                        "strokeWidth": 2
                    },
                    {
                        "id": "frontend-label",
                        "type": "text",
                        "x": 75.0,
                        "y": 115.0,
                        "width": 100.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "content": "Frontend Components"
                    },
                    {
                        "id": "backend-box",
                        "type": "rectangle",
                        "x": 250.0,
                        "y": 80.0,
                        "width": 150.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#E6F3FF",
                        "stroke": "#0066CC",
                        "strokeWidth": 2
                    },
                    {
                        "id": "backend-label",
                        "type": "text",
                        "x": 275.0,
                        "y": 115.0,
                        "width": 100.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "content": "Backend Services"
                    },
                    {
                        "id": "note-1",
                        "type": "text",
                        "x": 100.0,
                        "y": 200.0,
                        "width": 300.0,
                        "height": 60.0,
                        "rotation": 0.0,
                        "content": "Detailed system architecture with multiple layers and components"
                    }
                ]
            },
            "file": {
                "name": "complex-system",
                "type": "doc",
                "children": [
                    {
                        "name": "client",
                        "type": "section",
                        "children": [
                            {
                                "name": "src",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "components",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "Header.jsx",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "Footer.jsx",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "Navigation.jsx",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "pages",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "Home.jsx",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "About.jsx",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "Contact.jsx",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "utils",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "helpers.js",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "constants.js",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "App.jsx",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "index.js",
                                        "type": "doc"
                                    }
                                ]
                            },
                            {
                                "name": "public",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "index.html",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "favicon.ico",
                                        "type": "doc"
                                    }
                                ]
                            },
                            {
                                "name": "package.json",
                                "type": "doc"
                            },
                            {
                                "name": "README.md",
                                "type": "doc"
                            }
                        ]
                    },
                    {
                        "name": "server",
                        "type": "section",
                        "children": [
                            {
                                "name": "src",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "controllers",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "userController.js",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "productController.js",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "routes",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "userRoutes.js",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "productRoutes.js",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "middleware",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "authMiddleware.js",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "errorHandler.js",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "models",
                                        "type": "section",
                                        "children": [
                                            {
                                                "name": "User.js",
                                                "type": "doc"
                                            },
                                            {
                                                "name": "Product.js",
                                                "type": "doc"
                                            }
                                        ]
                                    },
                                    {
                                        "name": "server.js",
                                        "type": "doc"
                                    }
                                ]
                            },
                            {
                                "name": "package.json",
                                "type": "doc"
                            }
                        ]
                    },
                    {
                        "name": "database",
                        "type": "section",
                        "children": [
                            {
                                "name": "schemas",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "users.sql",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "products.sql",
                                        "type": "doc"
                                    }
                                ]
                            },
                            {
                                "name": "seeds",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "users.seed.js",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "products.seed.js",
                                        "type": "doc"
                                    }
                                ]
                            }
                        ]
                    },
                    {
                        "name": "docs",
                        "type": "section",
                        "children": [
                            {
                                "name": "api.md",
                                "type": "doc"
                            },
                            {
                                "name": "architecture.md",
                                "type": "doc"
                            }
                        ]
                    },
                    {
                        "name": "README.md",
                        "type": "doc"
                    },
                    {
                        "name": "Dockerfile",
                        "type": "doc"
                    },
                    {
                        "name": ".gitignore",
                        "type": "doc"
                    }
                ]
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 || $HTTP_CODE -eq 202 ]]; then
        print_success "Complex Structurize API scenario completed successfully (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Complex Structurize API scenario failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Main function to run general scenario tests
main() {
    echo "==========================================="
    echo " AIService API General Scenarios Tests"
    echo "==========================================="
    echo ""

    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed. Please install curl to run this script."
        exit 1
    fi

    # Check if jq is installed for pretty printing
    if ! command -v jq &> /dev/null; then
        print_warning "jq is not installed. Response will be displayed as raw JSON."
    fi

    # Run general scenario tests
    test_simple_summarize
    test_complex_summarize
    test_simple_structurize
    test_complex_structurize

    # Cleanup
    rm -f response.json

    print_status "All general scenario tests completed!"
}

# Show usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -a, --all      Run all general scenario tests (default)"
    echo "  -s, --sum      Run only summarization tests"
    echo "  -t, --struc    Run only structurization tests"
    echo ""
    echo "Examples:"
    echo "  $0                     # Run all general scenario tests"
    echo "  $0 --all              # Run all general scenario tests"
    echo "  $0 --sum              # Run only summarization tests"
    echo "  $0 --struc            # Run only structurization tests"
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        usage
        exit 0
        ;;
    -s|--sum)
        print_status "Running only summarization tests..."
        test_simple_summarize
        test_complex_summarize
        ;;
    -t|--struc)
        print_status "Running only structurization tests..."
        test_simple_structurize
        test_complex_structurize
        ;;
    -a|--all|"")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        usage
        exit 1
        ;;
esac