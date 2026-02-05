#!/bin/bash

# Script for testing AIService API endpoints
# This script contains curl commands for testing structurize and summarize APIs

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
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

# Function to test structurize API
test_structurize() {
    print_status "Testing Structurize API..."
    
    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/structurize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-structurize-complex-001",
            "userId": "user-12345",
            "requestType": "structurize",
            "board": {
                "boardId": "board-struct-001",
                "imageUrl": "",
                "elements": [
                    {
                        "id": "elem-1",
                        "type": "rectangle",
                        "x": 10.5,
                        "y": 20.0,
                        "width": 200.0,
                        "height": 150.0,
                        "rotation": 0.0,
                        "fill": "#FF5733",
                        "stroke": "#000000",
                        "strokeWidth": 2,
                        "cornerRadius": 10
                    },
                    {
                        "id": "elem-2",
                        "type": "text",
                        "x": 50.0,
                        "y": 60.0,
                        "width": 120.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "fill": "#FFFFFF",
                        "stroke": "#333333",
                        "strokeWidth": 1,
                        "content": "Main Component"
                    },
                    {
                        "id": "elem-3",
                        "type": "ellipse",
                        "x": 250.0,
                        "y": 100.0,
                        "width": 100.0,
                        "height": 80.0,
                        "rotation": 0.0,
                        "fill": "#33FF57",
                        "stroke": "#000000",
                        "strokeWidth": 2
                    },
                    {
                        "id": "elem-4",
                        "type": "line",
                        "x": 150.0,
                        "y": 100.0,
                        "width": 100.0,
                        "height": 0.0,
                        "rotation": 0.0,
                        "stroke": "#0000FF",
                        "strokeWidth": 3,
                        "points": [150.0, 100.0, 250.0, 100.0],
                        "tension": 0.5
                    },
                    {
                        "id": "elem-5",
                        "type": "text",
                        "x": 280.0,
                        "y": 120.0,
                        "width": 150.0,
                        "height": 40.0,
                        "rotation": 15.0,
                        "fill": "#F0F0F0",
                        "stroke": "#666666",
                        "strokeWidth": 1,
                        "content": "Connected Element with detailed description of functionality"
                    }
                ]
            },
            "file": {
                "name": "project-root",
                "type": "doc",
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
                                        "name": "MainComponent.vue",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "Button.jsx",
                                        "type": "doc"
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
                                                "name": "validators.ts",
                                                "type": "doc"
                                            }
                                        ]
                                    }
                                ]
                            },
                            {
                                "name": "services",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "api.service.ts",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "storage.service.ts",
                                        "type": "doc"
                                    }
                                ]
                            },
                            {
                                "name": "main.ts",
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
                                "name": "assets",
                                "type": "section",
                                "children": [
                                    {
                                        "name": "logo.svg",
                                        "type": "doc"
                                    },
                                    {
                                        "name": "styles.css",
                                        "type": "doc"
                                    }
                                ]
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
            }
        }')
    
    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 || $HTTP_CODE -eq 202 ]]; then
        print_success "Structurize API test completed successfully (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Structurize API test failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi
    
    echo ""
}

# Function to test summarize API with cat drawing
test_summarize() {
    print_status "Testing Summarize API with cat drawing..."
    
    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/summarize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-summarize-cat-001",
            "userId": "user-67890",
            "requestType": "summarize",
            "board": {
                "boardId": "board-summarize-001",
                "imageUrl": "",
                "elements": [
                    {
                        "id": "cat-head-1",
                        "type": "ellipse",
                        "x": 100.0,
                        "y": 100.0,
                        "width": 120.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#FFA500",
                        "stroke": "#000000",
                        "strokeWidth": 3
                    },
                    {
                        "id": "cat-ear-left",
                        "type": "triangle",
                        "x": 110.0,
                        "y": 80.0,
                        "width": 20.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "fill": "#FFA500",
                        "stroke": "#000000",
                        "strokeWidth": 2
                    },
                    {
                        "id": "cat-ear-right",
                        "type": "triangle",
                        "x": 170.0,
                        "y": 80.0,
                        "width": 20.0,
                        "height": 30.0,
                        "rotation": 0.0,
                        "fill": "#FFA500",
                        "stroke": "#000000",
                        "strokeWidth": 2
                    },
                    {
                        "id": "cat-eye-left",
                        "type": "ellipse",
                        "x": 125.0,
                        "y": 120.0,
                        "width": 15.0,
                        "height": 20.0,
                        "rotation": 0.0,
                        "fill": "#FFFFFF",
                        "stroke": "#000000",
                        "strokeWidth": 2
                    },
                    {
                        "id": "cat-eye-right",
                        "type": "ellipse",
                        "x": 155.0,
                        "y": 120.0,
                        "width": 15.0,
                        "height": 20.0,
                        "rotation": 0.0,
                        "fill": "#FFFFFF",
                        "stroke": "#000000",
                        "strokeWidth": 2
                    },
                    {
                        "id": "cat-pupil-left",
                        "type": "ellipse",
                        "x": 128.0,
                        "y": 125.0,
                        "width": 8.0,
                        "height": 12.0,
                        "rotation": 0.0,
                        "fill": "#000000",
                        "stroke": "#000000",
                        "strokeWidth": 1
                    },
                    {
                        "id": "cat-pupil-right",
                        "type": "ellipse",
                        "x": 158.0,
                        "y": 125.0,
                        "width": 8.0,
                        "height": 12.0,
                        "rotation": 0.0,
                        "fill": "#000000",
                        "stroke": "#000000",
                        "strokeWidth": 1
                    },
                    {
                        "id": "cat-nose",
                        "type": "ellipse",
                        "x": 145.0,
                        "y": 145.0,
                        "width": 10.0,
                        "height": 8.0,
                        "rotation": 0.0,
                        "fill": "#FF69B4",
                        "stroke": "#000000",
                        "strokeWidth": 1
                    },
                    {
                        "id": "cat-mouth",
                        "type": "line",
                        "x": 140.0,
                        "y": 155.0,
                        "width": 20.0,
                        "height": 0.0,
                        "rotation": 0.0,
                        "stroke": "#000000",
                        "strokeWidth": 2,
                        "points": [140.0, 155.0, 160.0, 155.0]
                    },
                    {
                        "id": "cat-whisker-left-top",
                        "type": "line",
                        "x": 130.0,
                        "y": 150.0,
                        "width": 30.0,
                        "height": 0.0,
                        "rotation": -15.0,
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "points": [130.0, 150.0, 160.0, 140.0]
                    },
                    {
                        "id": "cat-whisker-left-middle",
                        "type": "line",
                        "x": 130.0,
                        "y": 152.0,
                        "width": 30.0,
                        "height": 0.0,
                        "rotation": 0.0,
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "points": [130.0, 152.0, 160.0, 152.0]
                    },
                    {
                        "id": "cat-whisker-left-bottom",
                        "type": "line",
                        "x": 130.0,
                        "y": 154.0,
                        "width": 30.0,
                        "height": 0.0,
                        "rotation": 15.0,
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "points": [130.0, 154.0, 160.0, 164.0]
                    },
                    {
                        "id": "cat-whisker-right-top",
                        "type": "line",
                        "x": 160.0,
                        "y": 150.0,
                        "width": 30.0,
                        "height": 0.0,
                        "rotation": 15.0,
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "points": [160.0, 150.0, 190.0, 140.0]
                    },
                    {
                        "id": "cat-whisker-right-middle",
                        "type": "line",
                        "x": 160.0,
                        "y": 152.0,
                        "width": 30.0,
                        "height": 0.0,
                        "rotation": 0.0,
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "points": [160.0, 152.0, 190.0, 152.0]
                    },
                    {
                        "id": "cat-whisker-right-bottom",
                        "type": "line",
                        "x": 160.0,
                        "y": 154.0,
                        "width": 30.0,
                        "height": 0.0,
                        "rotation": -15.0,
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "points": [160.0, 154.0, 190.0, 164.0]
                    },
                    {
                        "id": "cat-body",
                        "type": "ellipse",
                        "x": 120.0,
                        "y": 200.0,
                        "width": 80.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#FFA500",
                        "stroke": "#000000",
                        "strokeWidth": 3
                    },
                    {
                        "id": "cat-tail",
                        "type": "line",
                        "x": 200.0,
                        "y": 230.0,
                        "width": 60.0,
                        "height": 0.0,
                        "rotation": -30.0,
                        "stroke": "#000000",
                        "strokeWidth": 8,
                        "points": [200.0, 230.0, 260.0, 180.0]
                    },
                    {
                        "id": "info-box-1",
                        "type": "rectangle",
                        "x": 300.0,
                        "y": 50.0,
                        "width": 200.0,
                        "height": 100.0,
                        "rotation": 0.0,
                        "fill": "#F0F8FF",
                        "stroke": "#4682B4",
                        "strokeWidth": 2,
                        "cornerRadius": 5
                    },
                    {
                        "id": "info-text-1",
                        "type": "text",
                        "x": 320.0,
                        "y": 80.0,
                        "width": 160.0,
                        "height": 40.0,
                        "rotation": 0.0,
                        "fill": "#000000",
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "content": "This is a cute cat diagram showing the main components of the system architecture."
                    },
                    {
                        "id": "process-box-1",
                        "type": "rectangle",
                        "x": 50.0,
                        "y": 350.0,
                        "width": 150.0,
                        "height": 60.0,
                        "rotation": 0.0,
                        "fill": "#E6F3FF",
                        "stroke": "#0066CC",
                        "strokeWidth": 2,
                        "cornerRadius": 8
                    },
                    {
                        "id": "process-text-1",
                        "type": "text",
                        "x": 70.0,
                        "y": 375.0,
                        "width": 110.0,
                        "height": 25.0,
                        "rotation": 0.0,
                        "fill": "#000000",
                        "stroke": "#000000",
                        "strokeWidth": 1,
                        "content": "Data Processing"
                    },
                    {
                        "id": "connection-line-1",
                        "type": "line",
                        "x": 125.0,
                        "y": 300.0,
                        "width": 0.0,
                        "height": 50.0,
                        "rotation": 0.0,
                        "stroke": "#0000FF",
                        "strokeWidth": 2,
                        "points": [125.0, 300.0, 125.0, 350.0]
                    }
                ]
            }
        }')
    
    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 || $HTTP_CODE -eq 202 ]]; then
        print_success "Summarize API test completed successfully (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Summarize API test failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi
    
    echo ""
}

# Function to test health endpoint
test_health() {
    print_status "Testing Health endpoint..."
    
    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X GET "$BASE_URL/health")
    
    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 200 ]]; then
        print_success "Health check passed (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Health check failed with HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi
    
    echo ""
}

# Main function to run tests
main() {
    echo "==========================================="
    echo " AIService API Testing Script"
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
    
    # Test health endpoint first
    test_health
    
    # Test structurize API
    test_structurize
    
    # Test summarize API
    test_summarize
    
    # Cleanup
    rm -f response.json
    
    print_status "All tests completed!"
}

# Show usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -s, --struct   Test only structurize API"
    echo "  -z, --summar   Test only summarize API"
    echo "  -a, --all      Test all APIs (default)"
    echo "  --health       Test only health endpoint"
    echo ""
    echo "Examples:"
    echo "  $0                     # Test all APIs"
    echo "  $0 --struct           # Test only structurize API"
    echo "  $0 --summar           # Test only summarize API"
    echo "  $0 --health           # Test only health endpoint"
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        usage
        exit 0
        ;;
    -s|--struct)
        test_structurize
        ;;
    -z|--summar)
        test_summarize
        ;;
    --health)
        test_health
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