#!/bin/bash

# Auth Service API Testing Script
# This script provides helper functions for testing the API

BASE_URL="http://localhost:8001"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Store tokens
ACCESS_TOKEN=""
REFRESH_TOKEN=""
TENANT_ID=""

echo_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

echo_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test health endpoint
test_health() {
    echo_info "Testing health endpoint..."
    curl -s -X GET "$BASE_URL/health" | jq .
}

# Login user
login() {
    local email=$1
    local password=$2
    local tenant_id=$3

    echo_info "Logging in user: $email"
    
    response=$(curl -s -X POST "$BASE_URL/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\",
            \"tenant_id\": \"$tenant_id\"
        }")
    
    echo "$response" | jq .
    
    ACCESS_TOKEN=$(echo "$response" | jq -r '.access_token')
    REFRESH_TOKEN=$(echo "$response" | jq -r '.refresh_token')
    TENANT_ID=$tenant_id
    
    if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
        echo_success "Login successful!"
        echo_info "Access Token: $ACCESS_TOKEN"
        echo_info "Refresh Token: $REFRESH_TOKEN"
    else
        echo_error "Login failed!"
        return 1
    fi
}

# Refresh token
refresh_token() {
    echo_info "Refreshing token..."
    
    response=$(curl -s -X POST "$BASE_URL/refresh" \
        -H "Content-Type: application/json" \
        -d "{
            \"refresh_token\": \"$REFRESH_TOKEN\"
        }")
    
    echo "$response" | jq .
    
    ACCESS_TOKEN=$(echo "$response" | jq -r '.access_token')
    REFRESH_TOKEN=$(echo "$response" | jq -r '.refresh_token')
    
    echo_success "Token refreshed!"
}

# Create tenant
create_tenant() {
    local name=$1

    if [ -z "$ACCESS_TOKEN" ]; then
        echo_error "Access token not set. Please login first."
        return 1
    fi

    echo_info "Creating tenant: $name"
    
    curl -s -X POST "$BASE_URL/tenants" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"name\": \"$name\"
        }" | jq .
}

# Create user
create_user() {
    local email=$1
    local password=$2
    local role_id=$3

    if [ -z "$ACCESS_TOKEN" ]; then
        echo_error "Access token not set. Please login first."
        return 1
    fi

    if [ -z "$TENANT_ID" ]; then
        echo_error "Tenant ID not set. Please login first."
        return 1
    fi

    echo_info "Creating user: $email with role: $role_id"
    
    curl -s -X POST "$BASE_URL/users" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\",
            \"tenant_id\": \"$TENANT_ID\",
            \"role_ids\": [\"$role_id\"]
        }" | jq .
}

# Decode JWT token
decode_jwt() {
    local token=$1
    
    if [ -z "$token" ]; then
        token=$ACCESS_TOKEN
    fi

    echo_info "Decoding JWT token..."
    
    # Extract payload (middle part) and decode
    payload=$(echo "$token" | cut -d'.' -f2)
    
    # Add padding if needed
    padding=$(( (4 - ${#payload} % 4) % 4 ))
    payload="$payload$(printf '%*s' "$padding" | tr ' ' '=')"
    
    echo "$payload" | base64 -d | jq . 2>/dev/null || echo_error "Invalid token format"
}

# Show current tokens
show_tokens() {
    echo_info "Current Tokens:"
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
    echo "Tenant ID: $TENANT_ID"
}

# Main menu
show_menu() {
    echo ""
    echo "======================================"
    echo "Auth Service API Testing"
    echo "======================================"
    echo "1. Test health endpoint"
    echo "2. Login user"
    echo "3. Refresh token"
    echo "4. Create tenant"
    echo "5. Create user"
    echo "6. Decode JWT"
    echo "7. Show current tokens"
    echo "8. Exit"
    echo "======================================"
}

# Interactive mode
interactive_mode() {
    while true; do
        show_menu
        read -p "Select option (1-8): " choice
        
        case $choice in
            1) test_health ;;
            2) 
                read -p "Email: " email
                read -sp "Password: " password
                echo
                read -p "Tenant ID: " tenant_id
                login "$email" "$password" "$tenant_id"
                ;;
            3) refresh_token ;;
            4)
                read -p "Tenant name: " name
                create_tenant "$name"
                ;;
            5)
                read -p "Email: " email
                read -sp "Password: " password
                echo
                read -p "Role ID: " role_id
                create_user "$email" "$password" "$role_id"
                ;;
            6) decode_jwt ;;
            7) show_tokens ;;
            8) echo "Goodbye!"; exit 0 ;;
            *) echo_error "Invalid option" ;;
        esac
    done
}

# Run script
if [ $# -eq 0 ]; then
    interactive_mode
else
    # Command line mode
    "$@"
fi
