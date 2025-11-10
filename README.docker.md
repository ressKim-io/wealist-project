# weAlist Docker 가이드

> 실무급 Docker 환경 설정 및 사용 가이드

## 📋 목차

- [개요](#개요)
- [빠른 시작](#빠른-시작)
- [디렉토리 구조](#디렉토리-구조)
- [환경별 설정](#환경별-설정)
- [사용 방법](#사용-방법)
- [모니터링](#모니터링)
- [트러블슈팅](#트러블슈팅)
- [마이그레이션 가이드](#마이그레이션-가이드)

---

## 🎯 개요

weAlist 프로젝트는 개발/프로덕션 환경을 명확하게 분리하고, 실무 수준의 보안과 성능을 갖춘 Docker 환경을 제공합니다.

### 주요 특징

- ✅ **환경 분리**: dev, prod 환경 완전 분리
- ✅ **보안 강화**: 네트워크 분리, 리소스 제한, 시크릿 관리
- ✅ **모니터링**: Prometheus + Grafana 통합
- ✅ **개발 편의성**: Hot Reload, 직관적인 스크립트
- ✅ **프로덕션 준비**: Health checks, Logging, Backup 지원

---

## 🚀 빠른 시작

### 1. 환경변수 설정

```bash
# 개발 환경
cp docker/env/.env.dev.example docker/env/.env.dev
# .env.dev 파일을 열어 필요한 값 수정 (특히 OAuth 관련)

# 프로덕션 환경 (필요시)
cp docker/env/.env.prod.example docker/env/.env.prod
# ⚠️ 모든 패스워드와 시크릿을 강력한 값으로 변경!
```

### 2. 개발 환경 실행

```bash
# 방법 1: 스크립트 사용 (권장)
./docker/scripts/dev.sh up

# 방법 2: Docker Compose 직접 사용
docker compose -f docker/compose/docker-compose.yml \
               -f docker/compose/docker-compose.dev.yml up
```

### 3. 서비스 접속

- **Frontend**: http://localhost:3000
- **User API**: http://localhost:8080
- **Board API**: http://localhost:8000
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

---

## 📁 디렉토리 구조

```
wealist-project/
├── docker/
│   ├── compose/                      # Docker Compose 파일들
│   │   ├── docker-compose.yml        # Base 설정 (공통)
│   │   ├── docker-compose.dev.yml    # 개발 환경 오버라이드
│   │   ├── docker-compose.prod.yml   # 프로덕션 환경 오버라이드
│   │   └── docker-compose.monitoring.yml  # 모니터링 스택
│   │
│   ├── env/                          # 환경변수 파일들
│   │   ├── .env.example              # 공통 템플릿
│   │   ├── .env.dev.example          # 개발 환경 템플릿
│   │   ├── .env.prod.example         # 프로덕션 환경 템플릿
│   │   └── .gitignore                # 시크릿 파일 제외
│   │
│   ├── init/                         # 초기화 스크립트
│   │   └── postgres/
│   │       └── init-db.sh            # DB 초기화
│   │
│   ├── nginx/                        # Nginx 설정
│   │   └── nginx.prod.conf           # 프로덕션 Nginx 설정
│   │
│   └── scripts/                      # 개발 편의 스크립트
│       ├── dev.sh                    # 개발 환경 관리
│       ├── prod.sh                   # 프로덕션 환경 관리
│       └── monitoring.sh             # 모니터링 관리
│
├── board-service/                    # Go 서비스
│   └── Dockerfile
├── user-service/                     # Spring Boot 서비스
│   └── Dockerfile
├── frontend/                         # React 프론트엔드
│   ├── Dockerfile                    # 프로덕션용
│   └── Dockerfile.local              # 개발용
│
└── README.docker.md                  # 이 문서
```

---

## ⚙️ 환경별 설정

### 개발 환경 (Development)

**특징:**
- 모든 포트가 호스트에 노출 (디버깅 편의)
- 프론트엔드 Hot Module Reload (HMR) 지원
- 데이터베이스 직접 접근 가능
- 상세한 디버그 로깅
- 리소스 제한 없음

**네트워크:**
- `frontend-net`: 프론트엔드 접근
- `backend-net`: API 서비스
- `database-net`: DB/Redis (외부 접근 가능)

### 프로덕션 환경 (Production)

**특징:**
- 최소 포트만 노출 (80, 443)
- 리소스 제한 적용 (CPU, Memory)
- Read-only 파일시스템
- Health checks 강화
- 데이터베이스 외부 접근 차단
- 프로덕션 수준 로깅

**네트워크:**
- `frontend-net`: 외부 접근
- `backend-net`: 내부 API 통신
- `database-net`: 완전히 격리됨 (internal)

**보안 설정:**
- `no-new-privileges`: 권한 상승 방지
- `cap_drop/cap_add`: 최소 권한 부여
- 네트워크 완전 격리

---

## 💻 사용 방법

### 개발 환경

#### 기본 명령어

```bash
# 시작 (포그라운드)
./docker/scripts/dev.sh up

# 시작 (백그라운드)
./docker/scripts/dev.sh up-d

# 중지
./docker/scripts/dev.sh down

# 재시작
./docker/scripts/dev.sh restart

# 로그 확인
./docker/scripts/dev.sh logs

# 특정 서비스 로그
./docker/scripts/dev.sh logs user-service

# 이미지 다시 빌드
./docker/scripts/dev.sh build

# 빌드 후 시작
./docker/scripts/dev.sh rebuild

# 상태 확인
./docker/scripts/dev.sh ps

# 컨테이너 접속
./docker/scripts/dev.sh exec user-service bash

# 완전 삭제 (볼륨 포함)
./docker/scripts/dev.sh clean
```

#### Docker Compose 직접 사용

```bash
# 시작
docker compose -f docker/compose/docker-compose.yml \
               -f docker/compose/docker-compose.dev.yml up

# 백그라운드 시작
docker compose -f docker/compose/docker-compose.yml \
               -f docker/compose/docker-compose.dev.yml up -d

# 중지
docker compose -f docker/compose/docker-compose.yml \
               -f docker/compose/docker-compose.dev.yml down
```

### 프로덕션 환경

```bash
# 시작
./docker/scripts/prod.sh up

# 중지
./docker/scripts/prod.sh down

# 재시작 (전체)
./docker/scripts/prod.sh restart

# 특정 서비스만 재시작
./docker/scripts/prod.sh restart user-service

# 로그 확인 (최근 100줄)
./docker/scripts/prod.sh logs

# 특정 서비스 로그 (최근 500줄)
./docker/scripts/prod.sh logs user-service 500

# 서비스 상태 확인
./docker/scripts/prod.sh status

# 헬스체크 상태
./docker/scripts/prod.sh health

# 데이터베이스 백업
./docker/scripts/prod.sh backup

# 이미지 업데이트
./docker/scripts/prod.sh pull

# 전체 업데이트 (이미지 갱신 + 재시작)
./docker/scripts/prod.sh update
```

---

## 📊 모니터링

### 모니터링 스택 실행

```bash
# 개발 환경 + 모니터링
./docker/scripts/monitoring.sh up dev

# 프로덕션 환경 + 모니터링
./docker/scripts/monitoring.sh up prod

# 모니터링만 중지
./docker/scripts/monitoring.sh down dev

# 모니터링 상태 확인
./docker/scripts/monitoring.sh status dev
```

### 접속 정보

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001
  - 기본 계정: `admin` / `admin`
- **Node Exporter**: http://localhost:9100/metrics
- **Redis Exporter**: http://localhost:9121/metrics
- **PostgreSQL Exporter**: http://localhost:9187/metrics

### Grafana 설정

1. Grafana 접속 (http://localhost:3001)
2. 로그인 (`admin` / `admin`)
3. 데이터소스 추가:
   - Type: Prometheus
   - URL: `http://prometheus:9090`
   - Save & Test

---

## 🔧 트러블슈팅

### 포트 충돌

```bash
# 사용 중인 포트 확인
sudo lsof -i :8080
sudo lsof -i :5432

# .env 파일에서 포트 변경
USER_HOST_PORT=8081
POSTGRES_HOST_PORT=5433
```

### 권한 문제

```bash
# 스크립트 실행 권한 부여
chmod +x docker/scripts/*.sh

# 볼륨 권한 문제 시
sudo chown -R $USER:$USER ./docker
```

### 빌드 캐시 문제

```bash
# 캐시 없이 다시 빌드
./docker/scripts/dev.sh build

# 또는
docker compose -f docker/compose/docker-compose.yml \
               -f docker/compose/docker-compose.dev.yml \
               build --no-cache
```

### 네트워크 문제

```bash
# 네트워크 정리
docker network prune

# 전체 정리 (주의!)
docker system prune -a --volumes
```

### 로그 확인

```bash
# 전체 로그
./docker/scripts/dev.sh logs

# 실시간 로그 (follow)
docker compose -f docker/compose/docker-compose.yml \
               -f docker/compose/docker-compose.dev.yml \
               logs -f user-service

# 최근 100줄
docker compose logs --tail=100 user-service
```

---

## 🔄 마이그레이션 가이드

### 기존 환경에서 새 구조로 전환

#### 1. 기존 컨테이너 중지 및 정리

```bash
# 기존 docker-compose 중지
docker-compose down

# 또는 기존 docker-compose.yaml 사용 중이었다면
docker compose -f docker-compose.yaml down
```

#### 2. 환경변수 파일 생성

```bash
# 기존 .env 파일 백업
cp .env .env.backup

# 새 환경변수 파일 생성
cp docker/env/.env.dev.example docker/env/.env.dev

# 기존 .env 파일의 값을 docker/env/.env.dev로 복사
# 특히 다음 항목들:
# - 데이터베이스 패스워드
# - JWT_SECRET
# - OAuth 클라이언트 ID/Secret
```

#### 3. 데이터 보존 (선택사항)

기존 볼륨을 그대로 사용하고 싶다면:

```bash
# 기존 볼륨 확인
docker volume ls | grep postgres
docker volume ls | grep redis

# 새 구조에서 기존 볼륨 사용
# docker-compose.yml의 volumes 섹션에서 이름 맞추기
```

또는 데이터를 백업 후 새로 시작:

```bash
# PostgreSQL 백업
docker exec wealist-postgres pg_dumpall -U postgres > backup.sql

# Redis 백업
docker exec wealist-redis redis-cli SAVE
docker cp wealist-redis:/data/dump.rdb ./redis_backup.rdb

# 기존 볼륨 삭제
docker volume rm wealist-postgres-data wealist-redis-data
```

#### 4. 새 구조로 시작

```bash
# 개발 환경 시작
./docker/scripts/dev.sh up-d

# 서비스 상태 확인
./docker/scripts/dev.sh ps

# 로그 확인
./docker/scripts/dev.sh logs
```

#### 5. 데이터 복원 (백업한 경우)

```bash
# PostgreSQL 복원
cat backup.sql | docker exec -i wealist-postgres psql -U postgres

# Redis 복원
docker cp redis_backup.rdb wealist-redis:/data/dump.rdb
docker restart wealist-redis
```

---

## 📝 환경변수 관리

### 필수 환경변수

#### 데이터베이스
- `POSTGRES_SUPERUSER`: PostgreSQL 관리자 계정
- `POSTGRES_SUPERUSER_PASSWORD`: 관리자 비밀번호
- `USER_DB_PASSWORD`: User 서비스 DB 비밀번호
- `BOARD_DB_PASSWORD`: Board 서비스 DB 비밀번호

#### 보안
- `JWT_SECRET`: JWT 토큰 시크릿 (64자 이상 권장)
- `REDIS_PASSWORD`: Redis 비밀번호

#### OAuth
- `GOOGLE_CLIENT_ID`: Google OAuth 클라이언트 ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth 시크릿

### 환경변수 생성 도구

```bash
# 강력한 랜덤 패스워드 생성 (32자)
openssl rand -base64 32

# JWT Secret 생성 (64자)
openssl rand -base64 64

# UUID 생성
uuidgen
```

---

## 🔐 보안 체크리스트

### 개발 환경
- [ ] `.env.dev` 파일이 `.gitignore`에 포함되어 있는지 확인
- [ ] OAuth 개발용 클라이언트 ID 사용
- [ ] 로컬에서만 접근 가능한지 확인

### 프로덕션 환경
- [ ] 모든 기본 패스워드 변경
- [ ] JWT_SECRET 64자 이상 랜덤 문자열 사용
- [ ] 데이터베이스 포트 외부 노출 안됨
- [ ] Redis 포트 외부 노출 안됨
- [ ] CORS 설정 실제 도메인으로 제한
- [ ] Grafana 관리자 비밀번호 변경
- [ ] SSL/TLS 인증서 설정 (권장)
- [ ] 정기적인 백업 설정

---

## 📚 참고 자료

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)

---

## 🆘 도움이 필요하신가요?

- 이슈 제기: [GitHub Issues]
- 문의: [팀 슬랙 채널]

---

**작성일**: 2025-01-10
**마지막 수정**: 2025-01-10
**버전**: 2.0.0
