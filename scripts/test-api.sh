#!/bin/bash

# Testing the Auth Service API with curl commands
# Make sure the service is running on http://localhost:8001

BASE_URL="http://localhost:8001"

echo "=== Auth Service API Testing ==="
echo ""

# 1. Health Check
echo "1. Testing health endpoint..."
curl -X GET "$BASE_URL/health" \
  -H "Content-Type: application/json" | jq .
echo ""
echo "---"
echo ""

# 2. Login Example
echo "2. Testing login..."
echo "Note: You need to create a user first to test login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123",
    "tenant_id": "tenant-1"
  }')
echo "$LOGIN_RESPONSE" | jq .
echo ""

# Extract tokens for next requests
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
  echo "Access Token obtained: ${ACCESS_TOKEN:0:50}..."
  echo ""
  
  # 3. Create Tenant (requires SUPER_ADMIN role)
  echo "3. Creating a new tenant (requires SUPER_ADMIN role)..."
  curl -X POST "$BASE_URL/tenants" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d '{
      "name": "SMA Negeri 2"
    }' | jq .
  echo ""
  echo "---"
  echo ""
  
  # 4. Create User
  echo "4. Creating a new user..."
  curl -X POST "$BASE_URL/users" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d '{
      "email": "operator@example.com",
      "password": "password123",
      "tenant_id": "tenant-1",
      "role_ids": ["operator"]
    }' | jq .
  echo ""
  echo "---"
  echo ""
  
  # 5. Refresh Token
  echo "5. Refreshing access token..."
  curl -X POST "$BASE_URL/refresh" \
    -H "Content-Type: application/json" \
    -d "{
      \"refresh_token\": \"$REFRESH_TOKEN\"
    }" | jq .
  echo ""
else
  echo "Login failed. Please ensure:"
  echo "1. The service is running"
  echo "2. The user exists in the database"
  echo "3. The credentials are correct"
fi
