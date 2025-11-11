# Docker 구조 마이그레이션 가이드

> 기존 Docker 설정에서 새로운 구조로 전환하는 가이드입니다.

## 📋 변경 사항 요약

### 이전 구조 (Before)
```
wealist-project/
├── docker-compose.yaml          # 메인 (모든 서비스 포함)
├── docker-compose.base.yml      # 중복된 내용
├── docker-compose.local.yml     # 프론트엔드 개발용
├── .env.example                 # 루트 환경변수
└── init-db.sh                   # 루트에 위치
```

### 새로운 구조 (After)
```
wealist-project/
├── docker/
│   ├── compose/
│   │   ├── docker-compose.yml           # Base (공통)
│   │   ├── docker-compose.dev.yml       # 개발 환경
│   │   ├── docker-compose.prod.yml      # 프로덕션 환경
│   │   └── docker-compose.monitoring.yml # 모니터링 (선택)
│   ├── env/
│   │   ├── .env.example
│   │   ├── .env.dev.example
│   │   └── .env.prod.example
│   ├── init/postgres/init-db.sh
│   ├── nginx/nginx.prod.conf
│   └── scripts/
│       ├── dev.sh
│       ├── prod.sh
│       └── monitoring.sh
└── README.docker.md
```

## 🎯 주요 개선 사항

### 1. 환경 분리
- ✅ 개발/프로덕션 환경 완전 분리
- ✅ 환경별 설정 오버라이드 패턴 사용
- ✅ 모니터링 스택 선택적 추가

### 2. 보안 강화
- ✅ 네트워크 3단 분리 (frontend-net, backend-net, database-net)
- ✅ 프로덕션에서 DB 포트 외부 노출 차단
- ✅ 리소스 제한 (CPU, Memory)
- ✅ Security options (no-new-privileges, cap_drop)
- ✅ 환경변수 파일 .gitignore 추가

### 3. 개발 편의성
- ✅ Shell 스크립트로 직관적인 명령어 제공
- ✅ 프론트엔드 HMR 지원
- ✅ 로깅 설정 표준화
- ✅ Health check 강화

### 4. 운영 편의성
- ✅ 자동 재시작 정책
- ✅ 백업 스크립트 제공
- ✅ 로그 순환 설정
- ✅ 상태 모니터링 도구

---

## 🔄 마이그레이션 절차

### Step 1: 기존 환경 중지 및 백업

```bash
# 1. 기존 컨테이너 중지
docker-compose down
# 또는
docker compose -f docker-compose.yaml down

# 2. 데이터 백업 (중요!)
# PostgreSQL 백업
docker exec wealist-postgres pg_dumpall -U postgres > backup_$(date +%Y%m%d).sql

# Redis 백업
docker exec wealist-redis redis-cli SAVE
docker cp wealist-redis:/data/dump.rdb redis_backup_$(date +%Y%m%d).rdb

# 3. 기존 환경변수 백업
cp .env .env.backup.$(date +%Y%m%d)

# 4. 현재 볼륨 확인 (선택사항 - 재사용할 경우)
docker volume ls | grep wealist
```

### Step 2: 새 환경변수 설정

```bash
# 1. 개발 환경변수 생성
cp docker/env/.env.dev.example docker/env/.env.dev

# 2. 기존 .env.backup에서 값 복사
# 다음 항목들을 특히 주의해서 복사:
# - 데이터베이스 패스워드들
# - JWT_SECRET
# - REDIS_PASSWORD
# - GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET

# 3. 에디터로 열어서 확인
vim docker/env/.env.dev
# 또는
code docker/env/.env.dev
```

**복사해야 할 주요 값들:**
```bash
# From .env.backup to docker/env/.env.dev
POSTGRES_SUPERUSER_PASSWORD=기존값
USER_DB_PASSWORD=기존값
BOARD_DB_PASSWORD=기존값
REDIS_PASSWORD=기존값
JWT_SECRET=기존값
GOOGLE_CLIENT_ID=기존값
GOOGLE_CLIENT_SECRET=기존값
```

### Step 3: 새 구조로 실행

```bash
# 1. 개발 환경 시작
./docker/scripts/dev.sh up-d

# 2. 서비스 상태 확인
./docker/scripts/dev.sh ps

# 3. 로그 확인 (문제 있는지 체크)
./docker/scripts/dev.sh logs

# 4. 서비스 접속 테스트
curl http://localhost:8080/actuator/health  # User Service
curl http://localhost:8000/health           # Board Service
curl http://localhost:3000                  # Frontend
```

### Step 4: 데이터 복원 (백업한 경우)

```bash
# PostgreSQL 복원
cat backup_YYYYMMDD.sql | docker exec -i wealist-postgres psql -U postgres

# Redis 복원
docker cp redis_backup_YYYYMMDD.rdb wealist-redis:/data/dump.rdb
docker restart wealist-redis

# 데이터 확인
docker exec -it wealist-postgres psql -U postgres -c "\l"
docker exec -it wealist-redis redis-cli -a $(grep REDIS_PASSWORD docker/env/.env.dev | cut -d= -f2) KEYS '*'
```

### Step 5: 기존 파일 정리 (선택사항)

```bash
# ⚠️ 주의: 새 구조가 정상 작동하는지 충분히 테스트한 후 진행!

# 1. 기존 compose 파일들 백업 디렉토리로 이동
mkdir -p .backup/old-docker-setup
mv docker-compose.yaml .backup/old-docker-setup/
mv docker-compose.base.yml .backup/old-docker-setup/
mv docker-compose.local.yml .backup/old-docker-setup/
mv init-db.sh .backup/old-docker-setup/

# 2. 기존 frontend compose도 백업
mv frontend/docker-compose.yml .backup/old-docker-setup/

# 3. 기존 환경변수 파일 백업
mv .env.example .backup/old-docker-setup/
mv board-service/.env.example .backup/old-docker-setup/
mv frontend/.env.example .backup/old-docker-setup/
```

---

## 🔍 검증 체크리스트

마이그레이션 후 다음 사항들을 확인하세요:

### 기능 테스트
- [ ] 모든 서비스가 정상적으로 시작되는가?
- [ ] Frontend에서 API 호출이 정상 작동하는가?
- [ ] 사용자 로그인/회원가입이 작동하는가?
- [ ] OAuth 로그인이 정상 작동하는가?
- [ ] 보드/카드 CRUD가 정상 작동하는가?

### 데이터 확인
- [ ] 기존 사용자 데이터가 유지되었는가?
- [ ] 기존 보드/카드 데이터가 유지되었는가?
- [ ] Redis 세션이 정상 작동하는가?

### 개발 환경
- [ ] Hot Reload가 작동하는가? (프론트엔드)
- [ ] 데이터베이스에 직접 접속 가능한가?
- [ ] 로그가 정상 출력되는가?
- [ ] 디버깅이 가능한가?

### 보안
- [ ] 환경변수 파일이 .gitignore에 포함되었는가?
- [ ] 프로덕션 환경에서 DB 포트가 외부 노출되지 않는가?
- [ ] 모든 패스워드가 안전하게 관리되는가?

---

## 🚨 문제 해결

### 문제: 서비스가 시작되지 않음

```bash
# 로그 확인
./docker/scripts/dev.sh logs

# 네트워크 확인
docker network ls | grep wealist

# 볼륨 확인
docker volume ls | grep wealist

# 모든 것 정리 후 재시작
./docker/scripts/dev.sh clean
./docker/scripts/dev.sh up-d
```

### 문제: 데이터베이스 연결 실패

```bash
# 데이터베이스 상태 확인
docker exec wealist-postgres pg_isready -U postgres

# 연결 테스트
docker exec -it wealist-postgres psql -U postgres -c "SELECT version();"

# 환경변수 확인
docker exec wealist-user-service env | grep DATABASE
```

### 문제: 포트 충돌

```bash
# 사용 중인 포트 확인
sudo lsof -i :8080
sudo lsof -i :5432

# docker/env/.env.dev에서 포트 변경
USER_HOST_PORT=8081
POSTGRES_HOST_PORT=5433
```

### 문제: 권한 오류

```bash
# 스크립트 실행 권한
chmod +x docker/scripts/*.sh

# 볼륨 권한
sudo chown -R $USER:$USER docker/
```

---

## 📊 명령어 비교표

| 작업 | 이전 | 새로운 |
|------|------|--------|
| 개발 환경 시작 | `docker-compose up` | `./docker/scripts/dev.sh up` |
| 프로덕션 시작 | `docker-compose -f docker-compose.yaml up -d` | `./docker/scripts/prod.sh up` |
| 로그 확인 | `docker-compose logs -f` | `./docker/scripts/dev.sh logs` |
| 서비스 재시작 | `docker-compose restart` | `./docker/scripts/dev.sh restart` |
| 중지 | `docker-compose down` | `./docker/scripts/dev.sh down` |
| 백업 | 수동 | `./docker/scripts/prod.sh backup` |

---

## 🎓 팀원 교육 포인트

팀원들에게 다음 사항을 공유하세요:

### 1. 새로운 디렉토리 구조
- `docker/` 아래에 모든 Docker 관련 파일이 정리됨
- 환경변수는 `docker/env/` 아래에 위치
- 절대 `.env.dev`, `.env.prod` 파일을 Git에 커밋하지 말 것

### 2. 새로운 명령어
- Makefile 대신 shell 스크립트 사용
- `./docker/scripts/dev.sh [command]` 형태로 실행
- `dev.sh help` 또는 인자 없이 실행하면 사용법 확인 가능

### 3. 환경변수 관리
- 개인별로 `docker/env/.env.dev` 파일 생성
- OAuth 키는 개발용 별도 생성 권장
- 패스워드 생성: `openssl rand -base64 32`

### 4. 문서 확인
- `README.docker.md`: 전체 사용 가이드
- `MIGRATION.md`: 이 파일 (마이그레이션 가이드)

---

## ✅ 마이그레이션 완료 후

마이그레이션이 성공적으로 완료되면:

1. **팀원들에게 공지**
   - 새로운 구조 설명
   - 환경변수 설정 방법
   - 새로운 명령어 사용법

2. **문서 업데이트**
   - README.md에 Quick Start 섹션 업데이트
   - 개발 가이드 문서 갱신

3. **CI/CD 업데이트** (필요시)
   - GitHub Actions 워크플로우 수정
   - 배포 스크립트 업데이트

4. **모니터링 설정** (선택사항)
   ```bash
   ./docker/scripts/monitoring.sh up dev
   ```

---

## 📞 지원

마이그레이션 중 문제가 발생하면:

1. `README.docker.md`의 트러블슈팅 섹션 확인
2. 기존 백업에서 롤백 가능 (`.backup/` 디렉토리)
3. 팀 슬랙 채널에 문의
4. GitHub Issues 생성

---

**작성일**: 2025-01-10
**작성자**: Claude AI Assistant
**버전**: 1.0.0
