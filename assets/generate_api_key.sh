#!/bin/bash

# Check if clientID is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <clientID>"
    echo "Expected AUTH_SECRET environment variable to be set."
    exit 1
fi

CLIENT_ID=$1

# Check if AUTH_SECRET is set
if [ -z "$AUTH_SECRET" ]; then
    echo "Error: AUTH_SECRET environment variable is not set."
    exit 1
fi

# Generate API Key: HMAC-SHA256(AUTH_SECRET, CLIENT_ID)
API_KEY=$(echo -n "$CLIENT_ID" | openssl dgst -sha256 -hmac "$AUTH_SECRET" | sed 's/^.* //')

echo "-----------------------------------"
echo "Client ID: $CLIENT_ID"
echo "API Key:   $API_KEY"
echo "-----------------------------------"
echo "To get a JWT token, use the following request:"
echo "curl -X POST http://localhost:8888/api/v1/auth/token \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"clientID\": \"$CLIENT_ID\", \"apiKey\": \"$API_KEY\"}'"
