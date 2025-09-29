#!/bin/bash

# Kneutral Operator API Test Script
# This script demonstrates how to interact with the Kneutral Operator API

set -e

# Configuration
API_BASE="${KNEUTRAL_API_URL:-http://localhost:8090}"
NAMESPACE="${NAMESPACE:-monitoring}"
VERBOSE="${VERBOSE:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
log() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Function to make API request with error handling
api_request() {
    local method=$1
    local endpoint=$2
    local data=${3:-""}
    local description=${4:-"API request"}

    log $BLUE "🔄 $description..."

    if [ "$VERBOSE" = "true" ]; then
        log $YELLOW "   Method: $method"
        log $YELLOW "   Endpoint: $API_BASE$endpoint"
        if [ -n "$data" ]; then
            log $YELLOW "   Data: $data"
        fi
    fi

    local curl_cmd="curl -s -w 'HTTP_STATUS:%{http_code}' -X $method '$API_BASE$endpoint'"

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi

    local response=$(eval $curl_cmd)
    local http_status=$(echo "$response" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)
    local body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')

    if [[ "$http_status" -ge 200 && "$http_status" -lt 300 ]]; then
        log $GREEN "   ✅ Success ($http_status)"
        if [ "$VERBOSE" = "true" ] && [ -n "$body" ]; then
            echo "$body" | jq . 2>/dev/null || echo "$body"
        fi
        echo "$body"
    else
        log $RED "   ❌ Failed ($http_status)"
        if [ -n "$body" ]; then
            echo "$body" | jq . 2>/dev/null || echo "$body"
        fi
        return 1
    fi
}

# Function to wait for condition
wait_for_condition() {
    local check_cmd="$1"
    local description="$2"
    local timeout=${3:-30}
    local interval=${4:-2}

    log $BLUE "⏳ Waiting for $description (timeout: ${timeout}s)..."

    local elapsed=0
    while [ $elapsed -lt $timeout ]; do
        if eval "$check_cmd" >/dev/null 2>&1; then
            log $GREEN "   ✅ Condition met after ${elapsed}s"
            return 0
        fi
        sleep $interval
        elapsed=$((elapsed + interval))
    done

    log $RED "   ❌ Timeout waiting for $description"
    return 1
}

# Main test function
main() {
    log $BLUE "🚀 Starting Kneutral Operator API Tests"
    log $BLUE "📍 API Base URL: $API_BASE"
    log $BLUE "📦 Target Namespace: $NAMESPACE"
    echo

    # Test 1: Health Check
    log $YELLOW "=== Test 1: Health Check ==="
    health_response=$(api_request "GET" "/health" "" "Health check")
    echo

    # Test 2: List all AlertRules (should be empty initially)
    log $YELLOW "=== Test 2: List All AlertRules ==="
    api_request "GET" "/api/v1/alertrules" "" "List all AlertRules"
    echo

    # Test 3: List AlertRules in specific namespace
    log $YELLOW "=== Test 3: List AlertRules in Namespace ==="
    api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules" "" "List AlertRules in $NAMESPACE"
    echo

    # Test 4: Create a basic AlertRule
    log $YELLOW "=== Test 4: Create Basic AlertRule ==="
    basic_alert='{
      "metadata": {
        "name": "test-cpu-alert"
      },
      "spec": {
        "groups": [
          {
            "name": "test.rules",
            "interval": "30s",
            "rules": [
              {
                "alert": "TestHighCPU",
                "expr": "cpu_usage > 80",
                "for": "5m",
                "labels": {
                  "severity": "warning",
                  "test": "true"
                },
                "annotations": {
                  "summary": "Test alert for high CPU",
                  "description": "This is a test alert created via API"
                }
              }
            ]
          }
        ]
      }
    }'
    create_response=$(api_request "POST" "/api/v1/namespaces/$NAMESPACE/alertrules" "$basic_alert" "Create basic AlertRule")
    echo

    # Test 5: Get the created AlertRule
    log $YELLOW "=== Test 5: Get Created AlertRule ==="
    api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules/test-cpu-alert" "" "Get test-cpu-alert"
    echo

    # Test 6: Update the AlertRule
    log $YELLOW "=== Test 6: Update AlertRule ==="
    updated_spec='{
      "groups": [
        {
          "name": "test.rules.updated",
          "interval": "60s",
          "rules": [
            {
              "alert": "TestHighCPUUpdated",
              "expr": "cpu_usage > 85",
              "for": "10m",
              "labels": {
                "severity": "critical",
                "test": "true",
                "updated": "true"
              },
              "annotations": {
                "summary": "Updated test alert for high CPU",
                "description": "This alert has been updated via API"
              }
            }
          ]
        }
      ]
    }'
    api_request "PUT" "/api/v1/namespaces/$NAMESPACE/alertrules/test-cpu-alert" "$updated_spec" "Update test-cpu-alert"
    echo

    # Test 7: Verify update
    log $YELLOW "=== Test 7: Verify Update ==="
    updated_alert=$(api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules/test-cpu-alert" "" "Get updated AlertRule")
    if echo "$updated_alert" | jq -e '.spec.groups[0].name == "test.rules.updated"' >/dev/null; then
        log $GREEN "   ✅ Update verified successfully"
    else
        log $RED "   ❌ Update verification failed"
    fi
    echo

    # Test 8: Create AlertRule from example file
    log $YELLOW "=== Test 8: Create from Example File ==="
    if [ -f "basic-cpu-alert.json" ]; then
        example_alert=$(cat basic-cpu-alert.json | jq '.metadata.name = "example-cpu-alert"')
        api_request "POST" "/api/v1/namespaces/$NAMESPACE/alertrules" "$example_alert" "Create from example file"
    else
        log $YELLOW "   ⏭️  Skipping - example file not found"
    fi
    echo

    # Test 9: List AlertRules again (should show created ones)
    log $YELLOW "=== Test 9: List AlertRules After Creation ==="
    final_list=$(api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules" "" "List AlertRules after creation")
    alert_count=$(echo "$final_list" | jq '.items | length' 2>/dev/null || echo "0")
    log $GREEN "   📊 Found $alert_count AlertRules"
    echo

    # Test 10: Error handling - try to get non-existent AlertRule
    log $YELLOW "=== Test 10: Error Handling ==="
    log $BLUE "🔄 Testing 404 error handling..."
    if api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules/non-existent" "" "Get non-existent AlertRule" 2>/dev/null; then
        log $RED "   ❌ Expected 404 error but got success"
    else
        log $GREEN "   ✅ 404 error handled correctly"
    fi
    echo

    # Test 11: Create AlertRule with invalid data
    log $YELLOW "=== Test 11: Invalid Data Handling ==="
    invalid_alert='{"metadata": {"name": "invalid"}, "spec": {}}'
    log $BLUE "🔄 Testing 400 error handling..."
    if api_request "POST" "/api/v1/namespaces/$NAMESPACE/alertrules" "$invalid_alert" "Create invalid AlertRule" 2>/dev/null; then
        log $RED "   ❌ Expected 400 error but got success"
    else
        log $GREEN "   ✅ 400 error handled correctly"
    fi
    echo

    # Cleanup (optional)
    if [ "${CLEANUP:-true}" = "true" ]; then
        log $YELLOW "=== Cleanup ==="
        api_request "DELETE" "/api/v1/namespaces/$NAMESPACE/alertrules/test-cpu-alert" "" "Delete test-cpu-alert" || true
        if [ -f "basic-cpu-alert.json" ]; then
            api_request "DELETE" "/api/v1/namespaces/$NAMESPACE/alertrules/example-cpu-alert" "" "Delete example-cpu-alert" || true
        fi
        echo
    fi

    log $GREEN "🎉 All tests completed!"
}

# Helper function to run interactive demo
interactive_demo() {
    log $BLUE "🎭 Interactive Demo Mode"
    echo "This demo will walk you through the API step by step."
    echo "Press Enter to continue between steps, or Ctrl+C to exit."
    echo

    read -p "Press Enter to start..."

    # Step 1: Health check
    log $YELLOW "Step 1: Check API health"
    api_request "GET" "/health" "" "Health check"
    read -p "Press Enter to continue..."

    # Step 2: Create AlertRule
    log $YELLOW "Step 2: Create a sample AlertRule"
    sample_alert='{
      "metadata": {"name": "demo-alert"},
      "spec": {
        "groups": [{
          "name": "demo.rules",
          "rules": [{
            "alert": "DemoAlert",
            "expr": "up == 0",
            "labels": {"severity": "warning"},
            "annotations": {"summary": "Demo alert"}
          }]
        }]
      }
    }'
    api_request "POST" "/api/v1/namespaces/$NAMESPACE/alertrules" "$sample_alert" "Create demo AlertRule"
    read -p "Press Enter to continue..."

    # Step 3: List AlertRules
    log $YELLOW "Step 3: List all AlertRules"
    api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules" "" "List AlertRules"
    read -p "Press Enter to continue..."

    # Step 4: Get specific AlertRule
    log $YELLOW "Step 4: Get the created AlertRule"
    api_request "GET" "/api/v1/namespaces/$NAMESPACE/alertrules/demo-alert" "" "Get demo-alert"
    read -p "Press Enter to continue..."

    # Step 5: Delete AlertRule
    log $YELLOW "Step 5: Clean up - delete the AlertRule"
    api_request "DELETE" "/api/v1/namespaces/$NAMESPACE/alertrules/demo-alert" "" "Delete demo-alert"

    log $GREEN "🎉 Demo completed!"
}

# Usage function
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -h, --help        Show this help message"
    echo "  -v, --verbose     Verbose output"
    echo "  -i, --interactive Run interactive demo"
    echo "  --no-cleanup      Skip cleanup after tests"
    echo
    echo "Environment variables:"
    echo "  KNEUTRAL_API_URL  API base URL (default: http://localhost:8090)"
    echo "  NAMESPACE         Target namespace (default: monitoring)"
    echo "  VERBOSE           Enable verbose output (default: false)"
    echo "  CLEANUP           Cleanup after tests (default: true)"
    echo
    echo "Examples:"
    echo "  $0                           # Run all tests"
    echo "  $0 -v                        # Run tests with verbose output"
    echo "  $0 -i                        # Run interactive demo"
    echo "  KNEUTRAL_API_URL=http://my-api:8090 $0  # Use custom API URL"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -i|--interactive)
            interactive_demo
            exit 0
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        *)
            log $RED "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Check if required tools are available
for tool in curl jq; do
    if ! command -v $tool &> /dev/null; then
        log $RED "❌ Required tool '$tool' is not installed"
        exit 1
    fi
done

# Run main test suite
main