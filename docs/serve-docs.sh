#!/bin/bash

# Simple documentation server for local development
# Serves the Swagger UI and API documentation

set -e

PORT=${PORT:-8080}
DOCS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Starting documentation server..."
echo "📁 Serving from: $DOCS_DIR"
echo "🌐 URL: http://localhost:$PORT"
echo "📖 Swagger UI: http://localhost:$PORT/swagger-ui/"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Check if Python is available and start a simple HTTP server
if command -v python3 &> /dev/null; then
    echo "Using Python 3 HTTP server..."
    cd "$DOCS_DIR"
    python3 -m http.server $PORT
elif command -v python &> /dev/null; then
    echo "Using Python 2 HTTP server..."
    cd "$DOCS_DIR"
    python -m SimpleHTTPServer $PORT
elif command -v node &> /dev/null; then
    echo "Using Node.js http-server..."
    if ! command -v http-server &> /dev/null; then
        echo "Installing http-server..."
        npm install -g http-server
    fi
    cd "$DOCS_DIR"
    http-server -p $PORT
else
    echo "❌ No suitable HTTP server found. Please install Python or Node.js."
    echo ""
    echo "Alternatives:"
    echo "  - Python 3: python3 -m http.server $PORT"
    echo "  - Python 2: python -m SimpleHTTPServer $PORT"
    echo "  - Node.js:  npx http-server -p $PORT"
    echo "  - Go:       go run -m http.server $PORT"
    exit 1
fi