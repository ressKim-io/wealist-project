#!/bin/bash

# =============================================================================
# weAlist EC2 Dev Environment Helper Script
# =============================================================================
# EC2 단일 인스턴스에서 모든 서비스를 관리하는 헬퍼 스크립트입니다.
#
# 사용법:
#   ./docker/scripts/ec2-dev.sh [command]
#
# 명령어:
#   up          - 서비스 시작 (백그라운드)
#   up-fg       - 서비스 시작 (포그라운드, 로그 출력)
#   down        - 서비스 중지
#   restart     - 서비스 재시작
#   logs        - 로그 확인 (전체)
#   logs [svc]  - 특정 서비스 로그 확인
#   ps          - 실행 중인 서비스 확인
#   build       - 이미지 빌드
#   rebuild     - 이미지 재빌드 후 시작
#   clean       - 모든 컨테이너 및 볼륨 삭제
#   health      - 헬스체크 확인
#   setup       - 초기 설정 (환경 변수 파일 생성)
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 프로젝트 루트 디렉토리로 이동
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# 환경 변수 파일 경로
ENV_FILE="docker/env/.env.ec2-dev"
ENV_EXAMPLE="docker/env/.env.ec2-dev.example"
COMPOSE_FILE="docker/compose/docker-compose.ec2-dev.yml"

# 함수: 에러 메시지 출력
error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# 함수: 성공 메시지 출력
success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 함수: 경고 메시지 출력
warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 함수: 정보 메시지 출력
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# 함수: 환경 변수 파일 확인
check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        error "$ENV_FILE 파일이 존재하지 않습니다."
        info "다음 명령어로 환경 변수 파일을 생성하세요:"
        echo ""
        echo "  cp $ENV_EXAMPLE $ENV_FILE"
        echo "  vi $ENV_FILE  # 필수 값 수정 (JWT_SECRET, 비밀번호 등)"
        echo ""
        info "또는 './docker/scripts/ec2-dev.sh setup' 명령어를 실행하세요."
        exit 1
    fi
}

# 함수: Docker Compose 명령어 실행
run_compose() {
    check_env_file
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

# 명령어 처리
case "${1:-}" in
    up)# 1. 환경 변수 파일 확인 및 로드 (AWS 변수 주입)
        check_env_file
        source "$ENV_FILE" 

        # 🚨 [주의]: ECR 로그인은 이 스크립트가 자동 수행할 수 없습니다. 
        #           이전에 안내드린 'docker login'을 사전에 실행해야 합니다.

        # 2. ECR 자동 태그 조회 로직 추가 (연관 배열 대신 일반 배열 사용)
        
        # 일반 배열: 서비스 이름과 레포지토리 이름을 순서대로 매칭합니다.
        SERVICE_NAMES=("USER_SERVICE" "BOARD_SERVICE")
        REPO_NAMES=("wealist-dev-user-service" "wealist-dev-board-service")
        
        info "ECR에서 서비스별 최신 이미지 태그를 조회하여 환경 변수에 설정합니다..."
        
        # 배열 인덱스를 사용하여 순회
        for i in "${!SERVICE_NAMES[@]}"; do
            SERVICE_NAME_BASE="${SERVICE_NAMES[$i]}"
            REPO_NAME="${REPO_NAMES[$i]}"
            VERSION_VAR_NAME="${SERVICE_NAME_BASE}_VERSION" # USER_SERVICE_VERSION, BOARD_SERVICE_VERSION
            
            info "  - 레포지토리: ${REPO_NAME} 조회 중..."
            
            DEFAULT_VERSION=$(grep "^${VERSION_VAR_NAME}=" "$ENV_FILE" | cut -d '=' -f 2 | tr -d '[:space:]' || echo "latest")
            
            LATEST_TAG=$(aws ecr describe-images \
                --repository-name "${REPO_NAME}" \
                --region ${AWS_REGION} \
                --query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags | [0]' \
                --output text 2>/dev/null)
                
            if [ -n "$LATEST_TAG" ] && [ "$LATEST_TAG" != "None" ] && [ "$LATEST_TAG" != "null" ]; then
                export "$VERSION_VAR_NAME"="$LATEST_TAG"
                success "  ✓ ${REPO_NAME}: 태그 ${LATEST_TAG} 으로 설정되었습니다."
            else
                export "$VERSION_VAR_NAME"="${DEFAULT_VERSION}"
                warning "  ✗ ${REPO_NAME}: ECR 태그를 가져올 수 없어 기본값 (${!VERSION_VAR_NAME})을 사용합니다."
                warning "    (AWS CLI 권한 및 ECR 로그인 상태를 확인하세요.)"
            fi
        done
        
        # 3. Docker Compose 실행
        info "EC2 Dev 환경 시작 중..."
        run_compose up -d
        success "모든 서비스가 시작되었습니다."
        info "다음 명령어로 로그를 확인하세요: ./docker/scripts/ec2-dev.sh logs"
        ;;

    up-fg)
        info "EC2 Dev 환경 시작 중 (포그라운드)..."
        # up-fg도 최신 태그를 가져오도록 로직을 복사하는 것이 좋으나,
        # 편의상 up에서만 자동 태그 조회를 유지하고, up-fg는 현재 환경 변수를 사용하도록 둡니다.
        # (만약 up-fg에서도 태그 자동 조회가 필요하다면, up 블록 로직을 복사/적용해야 합니다.)
        run_compose up
        ;;

    down)
        info "EC2 Dev 환경 중지 중..."
        run_compose down
        success "모든 서비스가 중지되었습니다."
        ;;

    restart)
        if [ -n "${2:-}" ]; then
            info "서비스 재시작 중: $2"
            run_compose restart "$2"
            success "$2 서비스가 재시작되었습니다."
        else
            info "모든 서비스 재시작 중..."
            run_compose restart
            success "모든 서비스가 재시작되었습니다."
        fi
        ;;

    logs)
        if [ -n "${2:-}" ]; then
            info "서비스 로그 확인: $2"
            run_compose logs -f "$2"
        else
            info "전체 서비스 로그 확인"
            run_compose logs -f
        fi
        ;;

    ps)
        info "실행 중인 서비스 확인"
        run_compose ps
        ;;

    build)
        info "이미지 빌드 중..."
        run_compose build
        success "이미지 빌드 완료"
        ;;

    rebuild)
        info "이미지 재빌드 및 재시작 중..."
        run_compose down
        run_compose build --no-cache
        run_compose up -d
        success "재빌드 및 재시작 완료"
        ;;

    clean)
        warning "모든 컨테이너, 네트워크, 볼륨을 삭제합니다."
        read -p "계속하시겠습니까? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            info "정리 중..."
            run_compose down -v --remove-orphans
            success "정리 완료"
        else
            info "작업 취소됨"
        fi
        ;;

    health)
        info "헬스체크 확인 중..."
        echo ""

        # 서비스별 헬스체크
        services=("user-service:8080/actuator/health" "board-service:8000/health" "prometheus:9090/-/healthy" "grafana:3001/api/health")

        for service in "${services[@]}"; do
            IFS=':' read -r name endpoint <<< "$service"
            if curl -sf "http://localhost:${endpoint}" > /dev/null 2>&1; then
                echo -e "${GREEN}✓${NC} $name: 정상"
            else
                echo -e "${RED}✗${NC} $name: 비정상"
            fi
        done

        echo ""
        info "컨테이너 상태:"
        run_compose ps
        ;;

    setup)
        info "EC2 Dev 환경 초기 설정"
        echo ""

        # 환경 변수 파일 생성
        if [ -f "$ENV_FILE" ]; then
            warning "$ENV_FILE 파일이 이미 존재합니다."
            read -p "덮어쓰시겠습니까? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                info "설정 취소됨"
                exit 0
            fi
        fi

        cp "$ENV_EXAMPLE" "$ENV_FILE"
        success "환경 변수 파일 생성 완료: $ENV_FILE"
        echo ""

        warning "다음 항목을 반드시 수정해야 합니다:"
        echo "  1. JWT_SECRET (64자 이상)"
        echo "  2. 모든 PASSWORD 항목"
        echo "  3. GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET"
        echo "  4. EC2 IP 관련 설정 (CORS_ORIGINS, VITE_* 등)"
        echo ""
        info "편집하려면: vi $ENV_FILE"

        # JWT 시크릿 자동 생성 옵션
        read -p "JWT_SECRET를 자동 생성하시겠습니까? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')
            # macOS와 Linux 모두 호환되도록 sed 사용
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" "$ENV_FILE"
            else
                sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" "$ENV_FILE"
            fi
            success "JWT_SECRET 자동 생성 완료"
        fi

        echo ""
        info "설정이 완료되면 다음 명령어로 시작하세요:"
        echo "  ./docker/scripts/ec2-dev.sh up"
        ;;

    exec)
        if [ -z "${2:-}" ]; then
            error "서비스 이름을 지정해주세요."
            echo "예: ./docker/scripts/ec2-dev.sh exec user-service"
            exit 1
        fi
        info "컨테이너 접속: $2"
        run_compose exec "$2" /bin/sh
        ;;

    help|--help|-h|"")
        echo "weAlist EC2 Dev Environment Helper"
        echo ""
        echo "사용법: ./docker/scripts/ec2-dev.sh [command]"
        echo ""
        echo "명령어:"
        echo "  up          - 서비스 시작 (백그라운드)"
        echo "  up-fg       - 서비스 시작 (포그라운드)"
        echo "  down        - 서비스 중지"
        echo "  restart     - 서비스 재시작"
        echo "  logs [svc]  - 로그 확인 (전체 또는 특정 서비스)"
        echo "  ps          - 실행 중인 서비스 확인"
        echo "  build       - 이미지 빌드"
        echo "  rebuild     - 이미지 재빌드 후 시작"
        echo "  clean       - 모든 컨테이너 및 볼륨 삭제"
        echo "  health      - 헬스체크 확인"
        echo "  setup       - 초기 설정 (환경 변수 파일 생성)"
        echo "  exec [svc]  - 컨테이너 접속"
        echo "  help        - 이 도움말 표시"
        echo ""
        echo "예시:"
        echo "  ./docker/scripts/ec2-dev.sh setup"
        echo "  ./docker/scripts/ec2-dev.sh up"
        echo "  ./docker/scripts/ec2-dev.sh logs user-service"
        echo "  ./docker/scripts/ec2-dev.sh health"
        echo ""
        ;;

    *)
        error "알 수 없는 명령어: $1"
        info "사용 가능한 명령어를 보려면 './docker/scripts/ec2-dev.sh help'를 실행하세요."
        exit 1
        ;;
esac
