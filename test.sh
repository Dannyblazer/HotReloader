#!/bin/bash

# Quick test script for Hot Reload Optimizer

echo "Building Hot Reload Optimizer..."
go build -o hotreloader .

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"
echo ""
echo "Testing with demo app..."
echo "Run './hotreloader examples/demo-app' to start watching"
echo ""
echo "In another terminal, try editing files:"
echo "  echo '// test change' >> examples/demo-app/src/utils.js"
echo ""
echo "Press Ctrl+C to see final stats"
