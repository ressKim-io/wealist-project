#!/bin/bash
# =============================================================================
# weAlist - Production Environment Management Script
# =============================================================================
# 프로덕션 환경을 관리하는 스크립트입니다.
#
# 사용법:
#   ./docker/scripts/prod.sh [command]
#
# Commands:
#   up         - 프로덕션 환경 시작
#   down       - 프로덕션 환경 중지
#   restart    - 프로덕션 환경 재시작
#   logs       - 로그 확인
#   status     - 서비스 상태 확인
#   backup     - 데이터베이스 백업
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 프로젝트 루트 디렉토리로 이동
cd "$(dirname "$0")/../.."

# 환경변수 파일 확인
ENV_FILE="docker/env/.env.prod"
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ 프로덕션 환경변수 파일이 없습니다: $ENV_FILE${NC}"
    echo -e "${YELLOW}   docker/env/.env.prod.example을 참고하여 생성하세요.${NC}"
    exit 1
fi

# Docker Compose 파일 경로
COMPOSE_FILES="-f docker/compose/docker-compose.yml -f docker/compose/docker-compose.prod.yml"

# 커맨드 처리
COMMAND=${1:-up}

case $COMMAND in
    up)
        echo -e "${BLUE}🚀 프로덕션 환경을 시작합니다...${NC}"
        echo -e "${YELLOW}⚠️  프로덕션 환경을 시작하려고 합니다. 계속하시겠습니까?${NC}"
        read -p "확인 (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker compose $COMPOSE_FILES up -d
            echo -e "${GREEN}✅ 프로덕션 환경이 시작되었습니다.${NC}"
            echo ""
            echo -e "${BLUE}📊 서비스 상태를 확인하세요:${NC}"
            docker compose $COMPOSE_FILES ps
        else
            echo -e "${YELLOW}취소되었습니다.${NC}"
            exit 0
        fi
        ;;

    down)
        echo -e "${YELLOW}⏹️  프로덕션 환경을 중지합니다...${NC}"
        echo -e "${RED}⚠️  프로덕션 서비스를 중지하려고 합니다. 계속하시겠습니까?${NC}"
        read -p "확인 (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker compose $COMPOSE_FILES down
            echo -e "${GREEN}✅ 프로덕션 환경이 중지되었습니다.${NC}"
        else
            echo -e "${YELLOW}취소되었습니다.${NC}"
            exit 0
        fi
        ;;

    restart)
        echo -e "${YELLOW}🔄 프로덕션 환경을 재시작합니다...${NC}"
        SERVICE=${2:-}
        if [ -z "$SERVICE" ]; then
            docker compose $COMPOSE_FILES restart
        else
            docker compose $COMPOSE_FILES restart "$SERVICE"
        fi
        echo -e "${GREEN}✅ 재시작이 완료되었습니다.${NC}"
        ;;

    logs)
        SERVICE=${2:-}
        LINES=${3:-100}
        if [ -z "$SERVICE" ]; then
            docker compose $COMPOSE_FILES logs --tail="$LINES" -f
        else
            docker compose $COMPOSE_FILES logs --tail="$LINES" -f "$SERVICE"
        fi
        ;;

    status)
        echo -e "${BLUE}📊 프로덕션 서비스 상태:${NC}"
        docker compose $COMPOSE_FILES ps
        echo ""
        echo -e "${BLUE}💾 볼륨 사용량:${NC}"
        docker volume ls | grep wealist
        ;;

    health)
        echo -e "${BLUE}🏥 서비스 헬스체크:${NC}"
        docker compose $COMPOSE_FILES ps --format json | jq -r '.[] | "\(.Name): \(.Health)"'
        ;;

    backup)
        echo -e "${BLUE}💾 데이터베이스 백업을 시작합니다...${NC}"
        BACKUP_DIR="./backups"
        mkdir -p "$BACKUP_DIR"
        TIMESTAMP=$(date +%Y%m%d_%H%M%S)

        # PostgreSQL 백업
        echo -e "${YELLOW}PostgreSQL 백업 중...${NC}"
        docker compose $COMPOSE_FILES exec -T postgres pg_dumpall -U postgres > "$BACKUP_DIR/postgres_backup_$TIMESTAMP.sql"

        # Redis 백업 (RDB 스냅샷)
        echo -e "${YELLOW}Redis 백업 중...${NC}"
        docker compose $COMPOSE_FILES exec -T redis redis-cli --no-auth-warning -a "${REDIS_PASSWORD:-redis}" SAVE
        docker cp wealist-redis:/data/dump.rdb "$BACKUP_DIR/redis_backup_$TIMESTAMP.rdb"

        echo -e "${GREEN}✅ 백업이 완료되었습니다: $BACKUP_DIR${NC}"
        ;;

    pull)
        echo -e "${BLUE}📥 최신 이미지를 가져옵니다...${NC}"
        docker compose $COMPOSE_FILES pull
        echo -e "${GREEN}✅ 이미지 업데이트 완료${NC}"
        ;;

    update)
        echo -e "${BLUE}🔄 프로덕션 환경을 업데이트합니다...${NC}"
        echo -e "${YELLOW}⚠️  서비스가 재시작됩니다. 계속하시겠습니까?${NC}"
        read -p "확인 (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker compose $COMPOSE_FILES pull
            docker compose $COMPOSE_FILES up -d --build
            echo -e "${GREEN}✅ 업데이트가 완료되었습니다.${NC}"
        else
            echo -e "${YELLOW}취소되었습니다.${NC}"
        fi
        ;;

    *)
        echo -e "${RED}❌ 알 수 없는 명령어: $COMMAND${NC}"
        echo ""
        echo "사용 가능한 명령어:"
        echo "  up         - 프로덕션 환경 시작"
        echo "  down       - 프로덕션 환경 중지"
        echo "  restart    - 재시작 (restart [service])"
        echo "  logs       - 로그 확인 (logs [service] [lines])"
        echo "  status     - 서비스 상태 확인"
        echo "  health     - 헬스체크 상태"
        echo "  backup     - 데이터베이스 백업"
        echo "  pull       - 최신 이미지 가져오기"
        echo "  update     - 전체 업데이트 (pull + up)"
        exit 1
        ;;
esac
