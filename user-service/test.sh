#!/bin/bash

# 서버 주소
BASE_URL="http://localhost:8080"

# 색상 코드
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 도우미 함수: API 요청 및 결과 출력
function test_api() {
    local method="$1"
    local endpoint="$2"
    local description="$3"
    local data="$4"
    local token="$5"
    local headers="Content-Type: application/json"

    if [ -n "$token" ]; then
        headers="$headers\nAuthorization: Bearer $token"
    fi

    echo -e "${YELLOW}=======================================================================${NC}"
    echo -e "${YELLOW}Testing: $description${NC}"
    echo -e "Request: $method $BASE_URL$endpoint"
    if [ -n "$data" ]; then
        echo "Body: $data"
    fi
    echo -e "${YELLOW}-----------------------------------------------------------------------${NC}"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            $(if [ -n "$token" ]; then echo "-H \"Authorization: Bearer $token\""; fi) \
            -d "$data" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" \
            $(if [ -n "$token" ]; then echo "-H \"Authorization: Bearer $token\""; fi) \
            "$BASE_URL$endpoint")
    fi

    # 응답에서 상태 코드와 본문 분리
    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
    body=$(echo "$response" | sed '$d')

    if [[ "$http_status" -ge 200 && "$http_status" -lt 300 ]]; then
        echo -e "${GREEN}SUCCESS (Status: $http_status)${NC}"
        echo "Response: $body"
    else
        echo -e "${RED}FAILURE (Status: $http_status)${NC}"
        echo "Response: $body"
    fi
    echo -e "${YELLOW}=======================================================================${NC}\n"
    # 다음 요청을 위해 잠시 대기
    sleep 1
}

# =================================================================================
# 테스트 시작
# =================================================================================

echo "Starting API tests..."

# 1. Health Check
test_api "GET" "/health" "Health Check"

# 2. 인증 (Auth)
# 테스트용 토큰 발급
echo "Fetching test token..."
auth_response=$(curl -s -X GET "$BASE_URL/api/auth/test")
access_token=eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJjYjc4YjczZC0xNTM2LTRlYmUtYjY5OC1lMDgzNTViZTgyNDciLCJpYXQiOjE3NjI5OTQ4MTgsImV4cCI6MTc2Mjk5ODQxOH0.5QfExGMupmEVvPPRcl8Qcz3ZhAFZQoSiTADCINbc2LDEx7SymPnD8_QjYVKvDqX941Nka9j7-lYUFaEuSJxEDA
refresh_token=$(echo "$auth_response" | jq -r '.refreshToken')
user_id=$(curl -s -H "Authorization: Bearer $access_token" "$BASE_URL/api/auth/me" | jq -r '.id')

if [ -z "$access_token" ] || [ "$access_token" == "null" ]; then
    echo -e "${RED}Failed to get test token. Exiting.${NC}"
    exit 1
fi
echo -e "${GREEN}Test token and user ID acquired.${NC}"
echo "Access Token: $access_token"
echo "User ID: $user_id"

# 내 정보 조회
test_api "GET" "/api/auth/me" "Get My Info" "" "$access_token"

# 토큰 갱신
test_api "POST" "/api/auth/refresh" "Refresh Token" "{\"refreshToken\": \"$refresh_token\"}"

# 로그아웃
test_api "POST" "/api/auth/logout" "Logout" "" "$access_token"
# 로그아웃 후 내 정보 조회 (실패해야 정상)
test_api "GET" "/api/auth/me" "Get My Info (After Logout)" "" "$access_token"


# 재로그인 (테스트 연속성을 위해)
echo "Re-logging in for subsequent tests..."
auth_response=$(curl -s -X GET "$BASE_URL/api/auth/test")
access_token=$(echo "$auth_response" | jq -r '.accessToken')
user_id=$(curl -s -H "Authorization: Bearer $access_token" "$BASE_URL/api/auth/me" | jq -r '.id')
echo -e "${GREEN}New access token acquired.${NC}"


# 3. 사용자 (User)
test_api "GET" "/api/users/me" "Get My User Info" "" "$access_token"
test_api "GET" "/api/users/$user_id" "Get User Info by ID" "" "$access_token"
test_api "PUT" "/api/users/$user_id" "Update User" "{\"name\": \"Updated Name\"}" "$access_token"
# test_api "GET" "/api/users/test/$user_id" "Get test token for user" "" # This is likely for dev purposes

# 4. 사용자 프로필 (UserProfile)
# 워크스페이스 먼저 생성
workspace_name="Test-Workspace-$(date +%s)"
create_workspace_data="{\"workspaceName\": \"$workspace_name\", \"isPublic\": true, \"allowAutoJoin\": true}"
workspace_response=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $access_token" -d "$create_workspace_data" "$BASE_URL/api/workspaces")
workspace_id=$(echo "$workspace_response" | jq -r '.workspaceId')
echo "Created workspace with ID: $workspace_id"

test_api "POST" "/api/profiles" "Create Profile" "{\"workspaceId\": \"$workspace_id\", \"displayName\": \"My Test Profile\"}" "$access_token"
test_api "GET" "/api/profiles/me" "Get My Default Profile" "" "$access_token"
test_api "GET" "/api/profiles/workspace/$workspace_id" "Get My Profile by Workspace" "" "$access_token"
test_api "GET" "/api/profiles/all/me" "Get All My Profiles" "" "$access_token"
test_api "PUT" "/api/profiles/me" "Update My Profile" "{\"userId\": \"$user_id\", \"displayName\": \"My Updated Profile\"}" "$access_token"


# 5. 워크스페이스 (Workspace)
test_api "GET" "/api/workspaces/all" "Get All My Workspaces" "" "$access_token"
test_api "GET" "/api/workspaces/public/Test" "Search Public Workspaces" "" "$access_token"
test_api "GET" "/api/workspaces/ids/$workspace_id" "Get Workspace by ID" "" "$access_token"
test_api "PUT" "/api/workspaces/ids/$workspace_id" "Update Workspace" "{\"workspaceName\": \"$workspace_name-updated\"}" "$access_token"
test_api "GET" "/api/workspaces/$workspace_id/settings" "Get Workspace Settings" "" "$access_token"
test_api "PUT" "/api/workspaces/$workspace_id/settings" "Update Workspace Settings" "{\"isPublic\": false, \"allowAutoJoin\": false}" "$access_token"
test_api "POST" "/api/workspaces/default" "Set Default Workspace" "{\"workspaceId\": \"$workspace_id\"}" "$access_token"

# 6. 워크스페이스 멤버 및 가입
test_api "GET" "/api/workspaces/$workspace_id/members" "Get Workspace Members" "" "$access_token"
# 초대, 승인, 거절 등은 다른 사용자가 필요하므로 기본적인 호출만 테스트
test_api "POST" "/api/workspaces/$workspace_id/members/invite" "Invite User" "{\"email\": \"invite@test.com\", \"name\": \"Invited User\"}" "$access_token"
test_api "GET" "/api/workspaces/$workspace_id/pendingMembers" "Get Pending Members" "" "$access_token"
test_api "POST" "/api/workspaces/join-requests" "Request to Join Workspace" "{\"workspaceId\": \"$workspace_id\"}" "$access_token"
test_api "GET" "/api/workspaces/$workspace_id/joinRequests" "Get Join Requests" "" "$access_token"


# 7. 정리 (Cleanup)
test_api "DELETE" "/api/profiles/$workspace_id" "Delete Profile" "" "$access_token"
test_api "DELETE" "/api/workspaces/$workspace_id" "Delete Workspace" "" "$access_token"
test_api "DELETE" "/api/users/me" "Delete My Account" "" "$access_token"
test_api "PUT" "/api/users/$user_id/restore" "Restore My Account" "" "$access_token"


echo "All tests finished."
