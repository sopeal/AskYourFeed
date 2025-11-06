#!/bin/bash

# Test script for POST /api/v1/qa endpoint - Happy Path
# This script tests the successful creation of a Q&A interaction

echo "=========================================="
echo "Testing POST /api/v1/qa - Happy Path"
echo "=========================================="
echo ""

# Configuration
BASE_URL="http://localhost:8080"
ENDPOINT="/api/v1/qa"

# Test 1: Happy path with explicit date range
echo "Test 1: Creating Q&A with explicit date range"
echo "----------------------------------------------"
curl -X POST "${BASE_URL}${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mock-token" \
  -d '{
    "question": "What are the main topics discussed this week?",
    "date_from": "2025-01-01T00:00:00Z",
    "date_to": "2025-01-07T23:59:59Z"
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo ""

# Test 2: Happy path with date defaults (last 24 hours)
echo "Test 2: Creating Q&A with default date range (last 24h)"
echo "--------------------------------------------------------"
curl -X POST "${BASE_URL}${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mock-token" \
  -d '{
    "question": "Jakie były główne tematy dyskusji w ostatnich 24 godzinach?"
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo ""

# Test 3: Happy path with longer question
echo "Test 3: Creating Q&A with detailed question"
echo "--------------------------------------------"
curl -X POST "${BASE_URL}${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mock-token" \
  -d '{
    "question": "Can you summarize the key insights and trends from the posts in my feed over the past week? Please focus on technology, AI developments, and any significant news.",
    "date_from": "2025-10-24T00:00:00Z",
    "date_to": "2025-10-31T23:59:59Z"
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "=========================================="
echo "Tests completed!"
echo "=========================================="
