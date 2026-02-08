#!/bin/bash

# Script for testing AIService API basic validation
# This script contains curl commands for testing basic validation scenarios

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

# Function to test summarize API with empty boardID
test_summarize_empty_board_id() {
    print_status "Testing Summarize API with empty boardID..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/summarize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-validation-001",
            "userId": "user-12345",
            "requestType": "summarize",
            "board": {
                "boardId": "",
                "imageUrl": "",
                "elements": []
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 400 ]]; then
        print_success "Summarize API correctly rejected empty boardID (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Summarize API should have rejected empty boardID but returned HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test summarize API with empty userID
test_summarize_empty_user_id() {
    print_status "Testing Summarize API with empty userId..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/summarize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-validation-002",
            "userId": "",
            "requestType": "summarize",
            "board": {
                "boardId": "board-valid-001",
                "imageUrl": "",
                "elements": []
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 400 ]]; then
        print_success "Summarize API correctly rejected empty userId (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Summarize API should have rejected empty userId but returned HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test structurize API with empty file data
test_structurize_empty_file() {
    print_status "Testing Structurize API with empty file data..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/structurize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-validation-003",
            "userId": "user-12345",
            "requestType": "structurize",
            "board": {
                "boardId": "board-struct-001",
                "imageUrl": "",
                "elements": []
            },
            "file": {
                "name": "",
                "type": ""
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 400 ]]; then
        print_success "Structurize API correctly rejected empty file data (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Structurize API should have rejected empty file data but returned HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test structurize API with empty userID
test_structurize_empty_user_id() {
    print_status "Testing Structurize API with empty userId..."

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/structurize" \
        -H "Content-Type: application/json" \
        -d '{
            "requestId": "req-validation-004",
            "userId": "",
            "requestType": "structurize",
            "board": {
                "boardId": "board-struct-002",
                "imageUrl": "",
                "elements": []
            },
            "file": {
                "name": "test-file",
                "type": "doc"
            }
        }')

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 400 ]]; then
        print_success "Structurize API correctly rejected empty userId (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Structurize API should have rejected empty userId but returned HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test summarize API with too many elements
test_summarize_too_many_elements() {
    print_status "Testing Summarize API with too many elements (>1000)..."

    # Generate 1001 elements
    ELEMENTS="["
    for i in $(seq 1 1001); do
        if [ $i -gt 1 ]; then
            ELEMENTS="$ELEMENTS,"
        fi
        ELEMENTS="$ELEMENTS{
            \"id\": \"elem-$i\",
            \"type\": \"rectangle\",
            \"x\": 10.0,
            \"y\": 20.0,
            \"width\": 100.0,
            \"height\": 50.0,
            \"rotation\": 0.0
        }"
    done
    ELEMENTS="$ELEMENTS]"

    REQUEST_BODY="{\"requestId\": \"req-validation-005\", \"userId\": \"user-12345\", \"requestType\": \"summarize\", \"board\": {\"boardId\": \"board-many-elem-001\", \"imageUrl\": \"\", \"elements\": $ELEMENTS}}"

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/summarize" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_BODY")

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 400 ]]; then
        print_success "Summarize API correctly rejected too many elements (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Summarize API should have rejected too many elements but returned HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Function to test structurize API with too many elements
test_structurize_too_many_elements() {
    print_status "Testing Structurize API with too many elements (>1000)..."

    # Generate 1001 elements
    ELEMENTS="["
    for i in $(seq 1 1001); do
        if [ $i -gt 1 ]; then
            ELEMENTS="$ELEMENTS,"
        fi
        ELEMENTS="$ELEMENTS{
            \"id\": \"elem-$i\",
            \"type\": \"rectangle\",
            \"x\": 10.0,
            \"y\": 20.0,
            \"width\": 100.0,
            \"height\": 50.0,
            \"rotation\": 0.0
        }"
    done
    ELEMENTS="$ELEMENTS]"

    REQUEST_BODY="{\"requestId\": \"req-validation-006\", \"userId\": \"user-12345\", \"requestType\": \"structurize\", \"board\": {\"boardId\": \"board-many-elem-002\", \"imageUrl\": \"\", \"elements\": $ELEMENTS}, \"file\": {\"name\": \"test-file\", \"type\": \"doc\"}}"

    RESPONSE=$(curl -s -o response.json -w "%{http_code}" \
        -X POST "$BASE_URL/structurize" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_BODY")

    HTTP_CODE=$RESPONSE
    if [[ $HTTP_CODE -eq 400 ]]; then
        print_success "Structurize API correctly rejected too many elements (HTTP $HTTP_CODE)"
        echo "Response:"
        cat response.json | jq '.' 2>/dev/null || cat response.json
    else
        print_error "Structurize API should have rejected too many elements but returned HTTP code: $HTTP_CODE"
        echo "Response:"
        cat response.json
    fi

    echo ""
}

# Main function to run validation tests
main() {
    echo "==========================================="
    echo " AIService API Basic Validation Tests"
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

    # Run validation tests
    test_summarize_empty_board_id
    test_summarize_empty_user_id
    test_structurize_empty_file
    test_structurize_empty_user_id
    test_summarize_too_many_elements
    test_structurize_too_many_elements

    # Cleanup
    rm -f response.json

    print_status "All validation tests completed!"
}

# Show usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -a, --all      Run all validation tests (default)"
    echo ""
    echo "Examples:"
    echo "  $0                     # Run all validation tests"
    echo "  $0 --all              # Run all validation tests"
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        usage
        exit 0
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