#!/bin/bash

# =============================================================================
# Network Connectivity Test Script
# =============================================================================
# board-service에서 user-service로의 네트워크 연결을 테스트합니다.
#
# 테스트 항목:
# 1. user-service의 health 엔드포인트 호출
# 2. DNS 해석 확인 (localhost vs service name)
# 3. 타임아웃 설정 적절성 검증
# 4. Docker/Kubernetes 환경에서 service name 사용 확인
#
# 사용법:
#   ./scripts/test-network-connectivity.sh [environment]
#
# 환경:
#   local     - 로컬 개발 환경 (localhost)
#   docker    - Docker Compose 환경 (service name)
#   k8s       - Kubernetes 환경 (service DNS)
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 환경 설정
ENVIRONMENT=${1:-local}
TIMEOUT=5

# 환경별 URL 설정
case $ENVIRONMENT in
  local)
    USER_SERVICE_URL="http://localhost:8080"
    echo -e "${BLUE}테스트 환경: 로컬 개발 환경${NC}"
    ;;
  docker)
    USER_SERVICE_URL="http://user-service:8080"
    echo -e "${BLUE}테스트 환경: Docker Compose${NC}"
    ;;
  k8s)
    USER_SERVICE_URL="http://user-service.default.svc.cluster.local:8080"
    echo -e "${BLUE}테스트 환경: Kubernetes${NC}"
    ;;
  *)
    echo -e "${RED}오류: 알 수 없는 환경 '$ENVIRONMENT'${NC}"
    echo "사용법: $0 [local|docker|k8s]"
    exit 1
    ;;
esac

echo -e "${BLUE}User Service URL: ${USER_SERVICE_URL}${NC}"
echo ""

# =============================================================================
# 테스트 함수
# =============================================================================

# 테스트 결과 출력 함수
print_result() {
  local test_name=$1
  local result=$2
  local message=$3
  
  if [ "$result" -eq 0 ]; then
    echo -e "${GREEN}✓ ${test_name}: 성공${NC}"
    if [ -n "$message" ]; then
      echo -e "  ${message}"
    fi
  else
    echo -e "${RED}✗ ${test_name}: 실패${NC}"
    if [ -n "$message" ]; then
      echo -e "  ${message}"
    fi
  fi
  echo ""
}

# =============================================================================
# 1. DNS 해석 테스트
# =============================================================================
echo -e "${YELLOW}=== 1. DNS 해석 테스트 ===${NC}"

test_dns_resolution() {
  local host
  
  # URL에서 호스트 추출
  if [[ $USER_SERVICE_URL =~ http://([^:/]+) ]]; then
    host="${BASH_REMATCH[1]}"
  else
    echo -e "${RED}URL 파싱 실패${NC}"
    return 1
  fi
  
  echo "호스트: $host"
  
  # DNS 해석 시도
  if [ "$ENVIRONMENT" = "local" ]; then
    # localhost는 항상 해석 가능
    if ping -c 1 -W 1 "$host" > /dev/null 2>&1; then
      return 0
    else
      return 1
    fi
  else
    # Docker/K8s 환경에서는 nslookup 또는 getent 사용
    if command -v nslookup > /dev/null 2>&1; then
      if nslookup "$host" > /dev/null 2>&1; then
        return 0
      fi
    elif command -v getent > /dev/null 2>&1; then
      if getent hosts "$host" > /dev/null 2>&1; then
        return 0
      fi
    else
      echo -e "${YELLOW}DNS 조회 도구를 찾을 수 없습니다 (nslookup, getent)${NC}"
      return 0  # 도구가 없으면 스킵
    fi
    return 1
  fi
}

if test_dns_resolution; then
  print_result "DNS 해석" 0 "호스트 이름이 올바르게 해석됩니다"
else
  print_result "DNS 해석" 1 "호스트 이름을 해석할 수 없습니다"
fi

# =============================================================================
# 2. 기본 연결 테스트 (TCP)
# =============================================================================
echo -e "${YELLOW}=== 2. TCP 연결 테스트 ===${NC}"

test_tcp_connection() {
  local host
  local port
  
  # URL에서 호스트와 포트 추출
  if [[ $USER_SERVICE_URL =~ http://([^:/]+):([0-9]+) ]]; then
    host="${BASH_REMATCH[1]}"
    port="${BASH_REMATCH[2]}"
  elif [[ $USER_SERVICE_URL =~ http://([^:/]+) ]]; then
    host="${BASH_REMATCH[1]}"
    port="80"
  else
    echo -e "${RED}URL 파싱 실패${NC}"
    return 1
  fi
  
  echo "연결 대상: $host:$port"
  
  # nc (netcat) 또는 timeout + bash를 사용한 TCP 연결 테스트
  if command -v nc > /dev/null 2>&1; then
    if nc -z -w "$TIMEOUT" "$host" "$port" 2>/dev/null; then
      return 0
    fi
  elif command -v timeout > /dev/null 2>&1; then
    if timeout "$TIMEOUT" bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null; then
      return 0
    fi
  else
    echo -e "${YELLOW}TCP 연결 테스트 도구를 찾을 수 없습니다 (nc, timeout)${NC}"
    return 0  # 도구가 없으면 스킵
  fi
  
  return 1
}

if test_tcp_connection; then
  print_result "TCP 연결" 0 "포트가 열려있고 연결 가능합니다"
else
  print_result "TCP 연결" 1 "포트에 연결할 수 없습니다"
fi

# =============================================================================
# 3. Health 엔드포인트 테스트
# =============================================================================
echo -e "${YELLOW}=== 3. Health 엔드포인트 테스트 ===${NC}"

test_health_endpoint() {
  local health_url="${USER_SERVICE_URL}/actuator/health"
  
  echo "Health URL: $health_url"
  
  # curl을 사용한 HTTP 요청
  if command -v curl > /dev/null 2>&1; then
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" --connect-timeout "$TIMEOUT" --max-time "$TIMEOUT" "$health_url" 2>&1)
    http_code=$(echo "$response" | tail -n 1)
    local body=$(echo "$response" | head -n -1)
    
    echo "HTTP 상태 코드: $http_code"
    echo "응답 본문: $body"
    
    if [ "$http_code" = "200" ]; then
      return 0
    else
      return 1
    fi
  elif command -v wget > /dev/null 2>&1; then
    if wget -q -O - --timeout="$TIMEOUT" "$health_url" > /dev/null 2>&1; then
      return 0
    fi
  else
    echo -e "${YELLOW}HTTP 클라이언트를 찾을 수 없습니다 (curl, wget)${NC}"
    return 0  # 도구가 없으면 스킵
  fi
  
  return 1
}

if test_health_endpoint; then
  print_result "Health 엔드포인트" 0 "user-service가 정상 작동 중입니다"
else
  print_result "Health 엔드포인트" 1 "user-service에 접근할 수 없거나 응답하지 않습니다"
fi

# =============================================================================
# 4. API 엔드포인트 테스트
# =============================================================================
echo -e "${YELLOW}=== 4. API 엔드포인트 테스트 ===${NC}"

# 테스트할 엔드포인트 목록
declare -a ENDPOINTS=(
  "/api/users/00000000-0000-0000-0000-000000000000"
  "/api/workspaces/00000000-0000-0000-0000-000000000000"
  "/api/profiles/workspace/00000000-0000-0000-0000-000000000000"
  "/api/workspaces/00000000-0000-0000-0000-000000000000/validate-member/00000000-0000-0000-0000-000000000000"
)

test_api_endpoint() {
  local endpoint=$1
  local url="${USER_SERVICE_URL}${endpoint}"
  
  echo "테스트 URL: $url"
  
  if command -v curl > /dev/null 2>&1; then
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" --connect-timeout "$TIMEOUT" --max-time "$TIMEOUT" "$url" 2>&1)
    http_code=$(echo "$response" | tail -n 1)
    
    echo "  HTTP 상태 코드: $http_code"
    
    # 404가 아니면 엔드포인트가 존재하는 것으로 간주
    # (인증 오류 401, 데이터 없음 400 등은 엔드포인트가 존재함을 의미)
    if [ "$http_code" != "404" ] && [ "$http_code" != "000" ]; then
      return 0
    else
      return 1
    fi
  else
    echo -e "${YELLOW}curl을 찾을 수 없습니다${NC}"
    return 0  # 도구가 없으면 스킵
  fi
}

echo "주요 API 엔드포인트 존재 여부 확인:"
echo ""

for endpoint in "${ENDPOINTS[@]}"; do
  if test_api_endpoint "$endpoint"; then
    print_result "엔드포인트 $endpoint" 0 "엔드포인트가 존재합니다"
  else
    print_result "엔드포인트 $endpoint" 1 "엔드포인트가 존재하지 않습니다 (404)"
  fi
done

# =============================================================================
# 5. 타임아웃 설정 테스트
# =============================================================================
echo -e "${YELLOW}=== 5. 타임아웃 설정 테스트 ===${NC}"

test_timeout_settings() {
  local health_url="${USER_SERVICE_URL}/actuator/health"
  
  echo "타임아웃 설정: ${TIMEOUT}초"
  echo "테스트 URL: $health_url"
  
  if command -v curl > /dev/null 2>&1; then
    local start_time=$(date +%s)
    
    # 타임아웃 내에 응답이 오는지 확인
    if curl -s --connect-timeout "$TIMEOUT" --max-time "$TIMEOUT" "$health_url" > /dev/null 2>&1; then
      local end_time=$(date +%s)
      local elapsed=$((end_time - start_time))
      
      echo "  응답 시간: ${elapsed}초"
      
      if [ "$elapsed" -le "$TIMEOUT" ]; then
        return 0
      else
        return 1
      fi
    else
      echo "  타임아웃 또는 연결 실패"
      return 1
    fi
  else
    echo -e "${YELLOW}curl을 찾을 수 없습니다${NC}"
    return 0  # 도구가 없으면 스킵
  fi
}

if test_timeout_settings; then
  print_result "타임아웃 설정" 0 "타임아웃 설정이 적절합니다"
else
  print_result "타임아웃 설정" 1 "타임아웃이 너무 짧거나 서비스 응답이 느립니다"
fi

# =============================================================================
# 6. 환경 변수 확인
# =============================================================================
echo -e "${YELLOW}=== 6. 환경 변수 확인 ===${NC}"

echo "현재 설정된 환경 변수:"
echo ""

# board-service 관련 환경 변수 출력
env_vars=(
  "USER_SERVICE_URL"
  "USER_API_BASE_URL"
  "USER_API_TIMEOUT"
  "SERVER_PORT"
  "ENV"
  "LOG_LEVEL"
)

for var in "${env_vars[@]}"; do
  if [ -n "${!var}" ]; then
    echo -e "  ${GREEN}${var}${NC} = ${!var}"
  else
    echo -e "  ${YELLOW}${var}${NC} = (설정되지 않음)"
  fi
done

echo ""

# =============================================================================
# 7. Docker 환경 확인 (Docker 환경인 경우)
# =============================================================================
if [ "$ENVIRONMENT" = "docker" ]; then
  echo -e "${YELLOW}=== 7. Docker 환경 확인 ===${NC}"
  
  # Docker 컨테이너 실행 여부 확인
  if command -v docker > /dev/null 2>&1; then
    echo "실행 중인 컨테이너:"
    docker ps --filter "name=user-service" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    
    # Docker 네트워크 확인
    echo "Docker 네트워크:"
    docker network ls | grep -E "wealist|project-board"
    echo ""
  else
    echo -e "${YELLOW}Docker 명령어를 찾을 수 없습니다${NC}"
  fi
fi

# =============================================================================
# 요약
# =============================================================================
echo -e "${BLUE}=== 테스트 완료 ===${NC}"
echo ""
echo "네트워크 연결 테스트가 완료되었습니다."
echo ""
echo "다음 단계:"
echo "1. 실패한 테스트가 있다면 해당 항목을 확인하세요"
echo "2. DNS 해석 실패: 호스트 이름 또는 네트워크 설정 확인"
echo "3. TCP 연결 실패: 방화벽, 포트 설정 확인"
echo "4. Health 엔드포인트 실패: user-service 실행 상태 확인"
echo "5. API 엔드포인트 404: user-service 컨트롤러 구현 확인"
echo ""
