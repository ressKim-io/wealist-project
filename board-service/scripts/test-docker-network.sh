#!/bin/bash

# =============================================================================
# Docker Network Connectivity Test Script
# =============================================================================
# Docker Compose 환경에서 board-service 컨테이너 내부에서
# user-service로의 네트워크 연결을 테스트합니다.
#
# 사용법:
#   ./scripts/test-docker-network.sh
#
# 이 스크립트는 호스트에서 실행되며, board-service 컨테이너 내부에서
# 명령어를 실행합니다.
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 컨테이너 이름
BOARD_SERVICE_CONTAINER="board-service"
USER_SERVICE_CONTAINER="user-service"
USER_SERVICE_URL="http://user-service:8080"

echo -e "${BLUE}=== Docker 네트워크 연결 테스트 ===${NC}"
echo ""

# =============================================================================
# 1. Docker 환경 확인
# =============================================================================
echo -e "${YELLOW}=== 1. Docker 환경 확인 ===${NC}"

# Docker가 설치되어 있는지 확인
if ! command -v docker &> /dev/null; then
    echo -e "${RED}오류: Docker가 설치되어 있지 않습니다${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker 설치 확인${NC}"
echo ""

# =============================================================================
# 2. 컨테이너 실행 상태 확인
# =============================================================================
echo -e "${YELLOW}=== 2. 컨테이너 실행 상태 확인 ===${NC}"

# board-service 컨테이너 확인
if docker ps --format '{{.Names}}' | grep -q "^${BOARD_SERVICE_CONTAINER}$"; then
    echo -e "${GREEN}✓ board-service 컨테이너 실행 중${NC}"
    docker ps --filter "name=${BOARD_SERVICE_CONTAINER}" --format "  상태: {{.Status}}"
else
    echo -e "${RED}✗ board-service 컨테이너가 실행 중이지 않습니다${NC}"
    echo "다음 명령어로 컨테이너를 시작하세요:"
    echo "  docker compose -f docker/compose/docker-compose.yml -f docker/compose/docker-compose.dev.yml up -d"
    exit 1
fi

echo ""

# user-service 컨테이너 확인
if docker ps --format '{{.Names}}' | grep -q "^${USER_SERVICE_CONTAINER}$"; then
    echo -e "${GREEN}✓ user-service 컨테이너 실행 중${NC}"
    docker ps --filter "name=${USER_SERVICE_CONTAINER}" --format "  상태: {{.Status}}"
else
    echo -e "${RED}✗ user-service 컨테이너가 실행 중이지 않습니다${NC}"
    echo "다음 명령어로 컨테이너를 시작하세요:"
    echo "  docker compose -f docker/compose/docker-compose.yml -f docker/compose/docker-compose.dev.yml up -d"
    exit 1
fi

echo ""

# =============================================================================
# 3. Docker 네트워크 확인
# =============================================================================
echo -e "${YELLOW}=== 3. Docker 네트워크 확인 ===${NC}"

# board-service가 연결된 네트워크 확인
echo "board-service 네트워크:"
docker inspect ${BOARD_SERVICE_CONTAINER} --format '{{range $key, $value := .NetworkSettings.Networks}}  - {{$key}}{{end}}'

echo ""

# user-service가 연결된 네트워크 확인
echo "user-service 네트워크:"
docker inspect ${USER_SERVICE_CONTAINER} --format '{{range $key, $value := .NetworkSettings.Networks}}  - {{$key}}{{end}}'

echo ""

# 공통 네트워크 확인
BOARD_NETWORKS=$(docker inspect ${BOARD_SERVICE_CONTAINER} --format '{{range $key, $value := .NetworkSettings.Networks}}{{$key}} {{end}}')
USER_NETWORKS=$(docker inspect ${USER_SERVICE_CONTAINER} --format '{{range $key, $value := .NetworkSettings.Networks}}{{$key}} {{end}}')

COMMON_NETWORK=""
for net in $BOARD_NETWORKS; do
    if echo "$USER_NETWORKS" | grep -q "$net"; then
        COMMON_NETWORK="$net"
        break
    fi
done

if [ -n "$COMMON_NETWORK" ]; then
    echo -e "${GREEN}✓ 공통 네트워크 발견: ${COMMON_NETWORK}${NC}"
else
    echo -e "${RED}✗ 공통 네트워크를 찾을 수 없습니다${NC}"
    echo "두 컨테이너가 같은 네트워크에 있어야 통신할 수 있습니다."
fi

echo ""

# =============================================================================
# 4. DNS 해석 테스트 (컨테이너 내부)
# =============================================================================
echo -e "${YELLOW}=== 4. DNS 해석 테스트 ===${NC}"

# nslookup 또는 getent를 사용하여 DNS 해석 테스트
if docker exec ${BOARD_SERVICE_CONTAINER} sh -c "command -v nslookup" &> /dev/null; then
    if docker exec ${BOARD_SERVICE_CONTAINER} nslookup user-service &> /dev/null; then
        echo -e "${GREEN}✓ DNS 해석 성공: user-service${NC}"
        docker exec ${BOARD_SERVICE_CONTAINER} nslookup user-service | grep -A 2 "Name:"
    else
        echo -e "${RED}✗ DNS 해석 실패: user-service${NC}"
    fi
elif docker exec ${BOARD_SERVICE_CONTAINER} sh -c "command -v getent" &> /dev/null; then
    if docker exec ${BOARD_SERVICE_CONTAINER} getent hosts user-service &> /dev/null; then
        echo -e "${GREEN}✓ DNS 해석 성공: user-service${NC}"
        docker exec ${BOARD_SERVICE_CONTAINER} getent hosts user-service
    else
        echo -e "${RED}✗ DNS 해석 실패: user-service${NC}"
    fi
else
    echo -e "${YELLOW}DNS 조회 도구를 찾을 수 없습니다 (nslookup, getent)${NC}"
fi

echo ""

# =============================================================================
# 5. TCP 연결 테스트 (컨테이너 내부)
# =============================================================================
echo -e "${YELLOW}=== 5. TCP 연결 테스트 ===${NC}"

# nc (netcat)를 사용하여 TCP 연결 테스트
if docker exec ${BOARD_SERVICE_CONTAINER} sh -c "command -v nc" &> /dev/null; then
    if docker exec ${BOARD_SERVICE_CONTAINER} nc -z -w 5 user-service 8080 &> /dev/null; then
        echo -e "${GREEN}✓ TCP 연결 성공: user-service:8080${NC}"
    else
        echo -e "${RED}✗ TCP 연결 실패: user-service:8080${NC}"
    fi
else
    echo -e "${YELLOW}netcat을 찾을 수 없습니다. HTTP 테스트로 대체합니다.${NC}"
fi

echo ""

# =============================================================================
# 6. Health 엔드포인트 테스트 (컨테이너 내부)
# =============================================================================
echo -e "${YELLOW}=== 6. Health 엔드포인트 테스트 ===${NC}"

HEALTH_URL="${USER_SERVICE_URL}/actuator/health"
echo "테스트 URL: ${HEALTH_URL}"
echo ""

# curl을 사용하여 HTTP 요청
if docker exec ${BOARD_SERVICE_CONTAINER} sh -c "command -v curl" &> /dev/null; then
    RESPONSE=$(docker exec ${BOARD_SERVICE_CONTAINER} curl -s -w "\n%{http_code}" --connect-timeout 5 --max-time 5 "${HEALTH_URL}" 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
    BODY=$(echo "$RESPONSE" | head -n -1)
    
    echo "HTTP 상태 코드: ${HTTP_CODE}"
    echo "응답 본문: ${BODY}"
    echo ""
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Health 엔드포인트 정상${NC}"
    else
        echo -e "${RED}✗ Health 엔드포인트 오류 (HTTP ${HTTP_CODE})${NC}"
    fi
elif docker exec ${BOARD_SERVICE_CONTAINER} sh -c "command -v wget" &> /dev/null; then
    if docker exec ${BOARD_SERVICE_CONTAINER} wget -q -O - --timeout=5 "${HEALTH_URL}" &> /dev/null; then
        echo -e "${GREEN}✓ Health 엔드포인트 정상${NC}"
    else
        echo -e "${RED}✗ Health 엔드포인트 오류${NC}"
    fi
else
    echo -e "${RED}HTTP 클라이언트를 찾을 수 없습니다 (curl, wget)${NC}"
fi

echo ""

# =============================================================================
# 7. API 엔드포인트 테스트 (컨테이너 내부)
# =============================================================================
echo -e "${YELLOW}=== 7. API 엔드포인트 테스트 ===${NC}"

# 테스트할 엔드포인트 목록
declare -a ENDPOINTS=(
    "/api/users/00000000-0000-0000-0000-000000000000"
    "/api/workspaces/00000000-0000-0000-0000-000000000000"
    "/api/profiles/workspace/00000000-0000-0000-0000-000000000000"
    "/api/workspaces/00000000-0000-0000-0000-000000000000/validate-member/00000000-0000-0000-0000-000000000000"
)

if docker exec ${BOARD_SERVICE_CONTAINER} sh -c "command -v curl" &> /dev/null; then
    for endpoint in "${ENDPOINTS[@]}"; do
        URL="${USER_SERVICE_URL}${endpoint}"
        
        RESPONSE=$(docker exec ${BOARD_SERVICE_CONTAINER} curl -s -w "\n%{http_code}" --connect-timeout 5 --max-time 5 "${URL}" 2>&1)
        HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
        
        echo "엔드포인트: ${endpoint}"
        echo "  HTTP 상태 코드: ${HTTP_CODE}"
        
        # 404가 아니면 엔드포인트가 존재하는 것으로 간주
        if [ "$HTTP_CODE" != "404" ] && [ "$HTTP_CODE" != "000" ]; then
            echo -e "  ${GREEN}✓ 엔드포인트 존재${NC}"
        else
            echo -e "  ${RED}✗ 엔드포인트 없음 (404)${NC}"
        fi
        echo ""
    done
else
    echo -e "${RED}curl을 찾을 수 없습니다${NC}"
fi

# =============================================================================
# 8. 환경 변수 확인 (컨테이너 내부)
# =============================================================================
echo -e "${YELLOW}=== 8. 환경 변수 확인 ===${NC}"

echo "board-service 컨테이너의 환경 변수:"
docker exec ${BOARD_SERVICE_CONTAINER} sh -c 'env | grep -E "(USER_SERVICE|USER_API|SERVER_PORT|ENV)" | sort'

echo ""

# =============================================================================
# 요약
# =============================================================================
echo -e "${BLUE}=== 테스트 완료 ===${NC}"
echo ""
echo "Docker 네트워크 연결 테스트가 완료되었습니다."
echo ""
echo "추가 디버깅이 필요한 경우:"
echo "  # board-service 컨테이너에 접속"
echo "  docker exec -it ${BOARD_SERVICE_CONTAINER} /bin/sh"
echo ""
echo "  # 컨테이너 내부에서 직접 테스트"
echo "  curl http://user-service:8080/actuator/health"
echo "  curl http://user-service:8080/api/workspaces/00000000-0000-0000-0000-000000000000"
echo ""
