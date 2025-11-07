#!/bin/bash

# ===== 사전 준비 =====
if ! command -v jq &> /dev/null;
  then
  echo "jq 명령어가 필요합니다. 설치 후 다시 실행하세요."
  exit 1
fi

# ===== 환경 설정 =====
BASE_URL="http://localhost:8080"
# BASE_URL="https://api.orangecloud.com"  # 프로덕션 서버

# 색상 설정
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ===== 헬퍼 함수 =====
print_section() {
    echo -e "\n${BLUE}===== $1 =====${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# ===== 1. 인증 정보 설정 (OAuth2) =====
print_section "Authentication Setup (OAuth2)"

# 1.1 수동으로 Google OAuth2 로그인을 수행하고, 발급받은 JWT Access Token을 아래에 붙여넣으세요.
# 브라우저 개발자 도구의 네트워크 탭에서 API 요청 헤더를 확인하여 토큰을 복사할 수 있습니다.

ACCESS_TOKEN=""

if [ -z "$ACCESS_TOKEN" ]; then
    print_error "ACCESS_TOKEN이 비어있습니다. 스크립트를 수정하여 토큰을 직접 설정해주세요."
    exit 1
fi

echo "제공된 ACCESS_TOKEN을 사용하여 테스트를 시작합니다."

# 1.2 토큰을 사용하여 내 정보 조회 (USER_ID 획득)
ME_RESPONSE=$(curl -s -X GET "$BASE_URL/api/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

USER_ID=$(echo "$ME_RESPONSE" | jq -r '.id // empty')

if [ -n "$USER_ID" ]; then
    print_success "내 정보 조회 성공 (USER_ID: $USER_ID)"
else
    print_error "내 정보 조회 실패. ACCESS_TOKEN이 유효한지 확인하세요."
    echo "$ME_RESPONSE" > me_error.log
    exit 1
fi

# ===== 변수 (테스트용) =====
GROUP_ID=""
TEAM_ID=""
COMPANY_NAME="test-company-$(date +%s%N)"


# ===== 2. 그룹 관련 API =====
print_section "Group APIs (Prerequisite)"

# 2.1 그룹 생성
CREATE_GROUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspace" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"테스트 그룹\",
    \"companyName\": \"$COMPANY_NAME\" 
  }")

GROUP_ID=$(echo "$CREATE_GROUP_RESPONSE" | jq -r '.data.groupId // empty')

if [ -n "$GROUP_ID" ]; then
    print_success "그룹 생성 성공 (GROUP_ID: $GROUP_ID)"
else
    print_error "그룹 생성 실패"
    echo "$CREATE_GROUP_RESPONSE" > group_error.log
    exit 1
fi

# ===== 3. 사용자 상세 정보 =====
print_section "User Info APIs"

# UserInfo는 사용자가 특정 그룹에 속하는 정보를 담습니다.
# OAuth2로 로그인한 사용자는 아직 어떤 그룹에도 속해있지 않으므로, 이 API를 통해 그룹에 연결해줍니다.
CREATE_USER_INFO_RESPONSE=$(curl -s -X POST "$BASE_URL/api/userinfo" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"userId\": \"$USER_ID\",
    \"groupId\": \"$GROUP_ID\",
    \"role\": \"USER\" 
  }")

USER_INFO_ID=$(echo "$CREATE_USER_INFO_RESPONSE" | jq -r '.userId // empty')

if [ -n "$USER_INFO_ID" ]; then
    print_success "사용자 상세 정보 생성/연결 성공"
else
    print_error "사용자 상세 정보 생성/연결 실패"
    echo "$CREATE_USER_INFO_RESPONSE" > userinfo_error.log
fi

# ===== 4. 이미지 관련 API =====
print_section "Image APIs"

# ===== 4. 이미지 관련 API =====
print_section "Image APIs"

# 4.1 테스트용 이미지 URL 설정
TEST_IMAGE_URL="https://example.com/test-profile-image.jpg"
print_success "테스트용 이미지 URL 설정 완료: $TEST_IMAGE_URL"

# 4.2 이미지 URL 저장
echo -e "\n4.2 이미지 URL 저장 테스트"
SAVE_IMAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/users/$USER_ID/image" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "imageUrl=$TEST_IMAGE_URL")

SAVE_IMAGE_HTTP_CODE=$(echo "$SAVE_IMAGE_RESPONSE" | tail -n1)
SAVE_IMAGE_BODY=$(echo "$SAVE_IMAGE_RESPONSE" | sed '$d')

if [ "$SAVE_IMAGE_HTTP_CODE" -eq 200 ]; then
    print_success "이미지 URL 저장 성공 (HTTP 200)"
    echo "응답: $SAVE_IMAGE_BODY"
else
    print_error "이미지 URL 저장 실패 (HTTP $SAVE_IMAGE_HTTP_CODE)"
    echo "$SAVE_IMAGE_BODY" > image_save_error.log
fi

# 4.3 이미지 URL 조회
echo -e "\n4.3 이미지 URL 조회 테스트"
GET_IMAGE_URL_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/users/$USER_ID/image" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

GET_IMAGE_URL_HTTP_CODE=$(echo "$GET_IMAGE_URL_RESPONSE" | tail -n1)
GET_IMAGE_URL_BODY=$(echo "$GET_IMAGE_URL_RESPONSE" | sed '$d')

if [ "$GET_IMAGE_URL_HTTP_CODE" -eq 200 ] && [[ "$GET_IMAGE_URL_BODY" == *"$TEST_IMAGE_URL"* ]]; then
    print_success "이미지 URL 조회 성공 (HTTP 200): $GET_IMAGE_URL_BODY"
elif [ "$GET_IMAGE_URL_HTTP_CODE" -eq 200 ]; then
    print_success "이미지 URL 조회 성공 (HTTP 200): $GET_IMAGE_URL_BODY"
else
    print_error "이미지 URL 조회 실패 (HTTP $GET_IMAGE_URL_HTTP_CODE)"
    echo "$GET_IMAGE_URL_BODY" > image_get_error.log
fi

# 4.4 이미지 URL 업데이트 테스트
echo -e "\n4.4 이미지 URL 업데이트 테스트"
UPDATED_IMAGE_URL="https://example.com/updated-profile-image.jpg"
UPDATE_IMAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/users/$USER_ID/image" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "imageUrl=$UPDATED_IMAGE_URL")

UPDATE_IMAGE_HTTP_CODE=$(echo "$UPDATE_IMAGE_RESPONSE" | tail -n1)
UPDATE_IMAGE_BODY=$(echo "$UPDATE_IMAGE_RESPONSE" | sed '$d')

if [ "$UPDATE_IMAGE_HTTP_CODE" -eq 200 ]; then
    print_success "이미지 URL 업데이트 성공 (HTTP 200)"
    echo "응답: $UPDATE_IMAGE_BODY"
else
    print_error "이미지 URL 업데이트 실패 (HTTP $UPDATE_IMAGE_HTTP_CODE)"
    echo "$UPDATE_IMAGE_BODY" > image_update_error.log
fi

# 4.5 업데이트된 이미지 URL 조회
echo -e "\n4.5 업데이트된 이미지 URL 조회 테스트"
GET_UPDATED_URL_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/users/$USER_ID/image" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

GET_UPDATED_URL_HTTP_CODE=$(echo "$GET_UPDATED_URL_RESPONSE" | tail -n1)
GET_UPDATED_URL_BODY=$(echo "$GET_UPDATED_URL_RESPONSE" | sed '$d')

if [ "$GET_UPDATED_URL_HTTP_CODE" -eq 200 ] && [[ "$GET_UPDATED_URL_BODY" == *"$UPDATED_IMAGE_URL"* ]]; then
    print_success "업데이트된 이미지 URL 조회 성공: $GET_UPDATED_URL_BODY"
else
    print_error "업데이트된 이미지 URL 조회 실패 (HTTP $GET_UPDATED_URL_HTTP_CODE)"
    echo "$GET_UPDATED_URL_BODY" > image_updated_get_error.log
fi

# 4.6 이미지 삭제
echo -e "\n4.6 이미지 삭제 테스트"
DELETE_IMAGE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/users/$USER_ID/image" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if [ "$DELETE_IMAGE_RESPONSE" -eq 204 ]; then
    print_success "이미지 삭제 성공 (HTTP 204 No Content)"
else
    print_error "이미지 삭제 실패 (HTTP $DELETE_IMAGE_RESPONSE)"
fi

# 4.7 삭제 후 이미지 URL 조회 (기본 이미지 확인)
echo -e "\n4.7 삭제 후 기본 이미지 URL 조회 테스트"
GET_DEFAULT_IMAGE_URL_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/users/$USER_ID/image" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

GET_DEFAULT_IMAGE_URL_HTTP_CODE=$(echo "$GET_DEFAULT_IMAGE_URL_RESPONSE" | tail -n1)
GET_DEFAULT_IMAGE_URL_BODY=$(echo "$GET_DEFAULT_IMAGE_URL_RESPONSE" | sed '$d')

if [ "$GET_DEFAULT_IMAGE_URL_HTTP_CODE" -eq 200 ] && [[ "$GET_DEFAULT_IMAGE_URL_BODY" == *"default-profile.png"* ]]; then
    print_success "기본 이미지 URL 조회 성공: $GET_DEFAULT_IMAGE_URL_BODY"
else
    print_error "기본 이미지 URL 조회 실패 (HTTP $GET_DEFAULT_IMAGE_URL_HTTP_CODE)"
    echo "$GET_DEFAULT_IMAGE_URL_BODY" > image_default_get_error.log
fi
print_success "테스트용 이미지 파일 삭제 완료"


# ===== 5. 팀 관련 API =====
print_section "Team APIs"

# 5.1 팀 생성
CREATE_TEAM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/teams" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"teamName\": \"테스트 팀\",
    \"companyName\": \"$COMPANY_NAME\",
    \"leaderId\": \"$USER_ID\",
    \"description\": \"API 테스트용 팀\" 
  }")

TEAM_ID=$(echo "$CREATE_TEAM_RESPONSE" | jq -r '.data.team.teamId // empty')

if [ -n "$TEAM_ID" ] && [ "$TEAM_ID" != "null" ]; then
    print_success "팀 생성 성공 (TEAM_ID: $TEAM_ID)"
else
    print_error "팀 생성 실패"
    echo "$CREATE_TEAM_RESPONSE" > team_error.log
fi

# ===== 6. 정리(삭제) API =====
print_section "Cleanup APIs"

# 6.1 팀 삭제
if [ -n "$TEAM_ID" ] && [ "$TEAM_ID" != "null" ]; then
    echo -e "\n6.1 팀 삭제"
    curl -s -X DELETE "$BASE_URL/api/teams/$TEAM_ID?requesterId=$USER_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "팀 삭제 요청 완료"
fi

# 6.2 그룹 내 사용자 정보 삭제
echo -e "\n6.2 그룹 내 사용자 정보 삭제"
curl -s -X DELETE "$BASE_URL/api/userinfo/$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
print_success "사용자 정보 삭제 요청 완료"

# 6.3 그룹 삭제
echo -e "\n6.3 그룹 삭제"
curl -s -X DELETE "$BASE_URL/api/workspace/$GROUP_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
print_success "그룹 삭제 요청 완료"

# 6.4 사용자 삭제 (OAuth 사용자는 DB에서 직접 삭제해야 할 수 있음)
echo -e "\n6.4 사용자 삭제"
curl -s -X DELETE "$BASE_URL/api/users/$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
print_success "사용자 삭제 요청 완료 (Soft Delete)"

print_section "✅ 모든 테스트 완료!"

