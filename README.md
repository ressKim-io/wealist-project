# weAlist BASIC Project

프로젝트 관리 플랫폼 - 마이크로서비스 아키텍처


## 🏗️ 서비스 구조

| 서비스 | 기술 스택 | 포트 | 상태 | 설명 |
|--------|----------|------|------|------|
| **User Service** | Spring Boot (Java) | 8080 | ✅ Active | 사용자 인증 및 관리 |
| **Board Service** | Gin (Go) | 8000 | ✅ Active | 보드/칸반 관리, 커스텀 필드 |
| **Frontend** | React (TypeScript) | 3000 | 🚧 Dev | 프론트엔드 애플리케이션 |

## 🌐 ALB Path-Based Routing

AWS 환경에서는 Application Load Balancer(ALB)를 통해 서비스별 경로 기반 라우팅을 제공합니다.

### 라우팅 구조

```
Client Request
    ↓
ALB (https://api.wealist.co.kr)
    ├─ /api/users/*   → User Service (port 8080)
    └─ /api/boards/*  → Board Service (port 8000)
```

### 환경별 설정

#### 로컬 개발 환경
- ALB 없이 직접 서비스 접근
- Context/Base Path 설정 없음

```bash
# User Service
curl http://localhost:8080/api/workspaces/all

# Board Service
curl http://localhost:8000/health
```

#### AWS 환경 (Production)
- ALB를 통한 통합 엔드포인트
- 서비스별 prefix 자동 제거

```bash
# User Service (ALB → /api/users 제거 → User Service)
curl https://api.wealist.co.kr/api/users/api/workspaces/all

# Board Service (ALB → /api/boards 제거 → Board Service)
curl https://api.wealist.co.kr/api/boards/health
```

### API 엔드포인트 예시

| 환경 | User Service | Board Service |
|------|-------------|---------------|
| **로컬** | `http://localhost:8080/api/workspaces/all` | `http://localhost:8000/health` |
| **AWS** | `https://api.wealist.co.kr/api/users/api/workspaces/all` | `https://api.wealist.co.kr/api/boards/health` |

### 서비스별 Path 설정

**User Service (Spring Boot)**
- AWS 환경: `server.servlet.context-path=/api/users` (application-aws.yml)
- 로컬 환경: Context path 없음 (application-local.yml)
- Profile 전환: `SPRING_PROFILES_ACTIVE` 환경 변수

**Board Service (Go)**
- AWS 환경: `SERVER_BASE_PATH=/api/boards` 환경 변수
- 로컬 환경: `SERVER_BASE_PATH=""` (빈 문자열)
- 환경 전환: `ENV` 환경 변수

자세한 배포 가이드는 [docs/ALB_ROUTING_DEPLOYMENT.md](docs/ALB_ROUTING_DEPLOYMENT.md)를 참조하세요.

### ALB 설정 검증

ALB 설정이 올바르게 구성되었는지 확인하는 스크립트를 제공합니다.

#### 1. API 엔드포인트 검증
```bash
# 기본 검증 (Health check만)
./scripts/verify-alb-setup.sh

# 상세 검증 (응답 내용 포함)
VERBOSE=true ./scripts/verify-alb-setup.sh

# 커스텀 ALB URL 사용
ALB_URL=https://your-alb-url.com ./scripts/verify-alb-setup.sh
```

#### 2. Target Group Health 확인 (AWS CLI 필요)
```bash
# Target Group health 상태 및 Listener Rules 확인
./scripts/check-alb-health.sh
```

**참고**: `check-alb-health.sh` 스크립트는 AWS CLI가 설치되어 있고 적절한 권한이 설정되어 있어야 합니다.

#### 3. 상세 검증 가이드
AWS Console에서 직접 확인하는 방법을 포함한 전체 검증 절차는 [docs/ALB_VERIFICATION_GUIDE.md](docs/ALB_VERIFICATION_GUIDE.md)를 참조하세요.

## 🚀 주요 기능

- ✅ 워크스페이스 & 프로젝트 관리
- ✅ 커스텀 보드 (역할, 진행단계, 중요도 기반)
- ✅ 드래그 앤 드롭 기능 (사용자별 순서 저장)
- ✅ 멤버 관리 및 역할 기반 접근 제어
- ✅ JWT 기반 인증
- ✅ 소프트 삭제 (복구 가능)
- ✅ RESTful API with Swagger

## 📋 실행 방법

### 1. 환경 변수 설정

개발 환경용 환경변수 파일을 생성합니다:

```bash
# 개발 환경 템플릿 복사
cp docker/env/.env.dev.example docker/env/.env.dev

# .env.dev 파일을 열어 필요한 값 수정 (특히 OAuth 관련)
vi docker/env/.env.dev
```

### 2. 개발 환경 시작

**방법 1: 스크립트 사용 (권장)**

```bash
# 서비스 시작 (포그라운드)
./docker/scripts/dev.sh up

# 서비스 시작 (백그라운드)
./docker/scripts/dev.sh up-d

# 로그 확인
./docker/scripts/dev.sh logs

# 서비스 종료
./docker/scripts/dev.sh down
```

**방법 2: Docker Compose 직접 사용**

```bash
# --env-file 옵션 필수!
docker compose --env-file docker/env/.env.dev \
  -f docker/compose/docker-compose.yml \
  -f docker/compose/docker-compose.dev.yml \
  up -d
```

> **중요**: `--env-file` 옵션을 빼먹으면 환경변수 인식 오류가 발생합니다. 스크립트 사용을 권장합니다.

### 3. 서비스 확인

개발 환경에서 접속 가능한 서비스:

- **Frontend**: http://localhost:3000
- **User Service**: http://localhost:8080/health
- **User Service Swagger**: http://localhost:8080/swagger-ui/index.html
- **Board Service**: http://localhost:8000/health
- **Board Service Swagger**: http://localhost:8000/swagger/index.html
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

### 4. 테스트

Board Service 통합 테스트:
```bash
./scripts/tests/test-board-integration.sh
```

User Service 테스트:
```bash
./scripts/tests/test-user-service.sh
```

자세한 내용은 [테스트 가이드](./scripts/tests/README.md)를 참고하세요.

### 5. 추가 명령어

```bash
# 서비스 재시작
./docker/scripts/dev.sh restart

# 실행 중인 서비스 확인
./docker/scripts/dev.sh ps

# 이미지 다시 빌드
./docker/scripts/dev.sh build

# 컨테이너 접속
./docker/scripts/dev.sh exec user-service

# 모든 서비스 및 볼륨 삭제
./docker/scripts/dev.sh clean
```

## 🔧 문제 해결

### 환경변수 인식 문제

만약 다음과 같은 경고가 나온다면:

```
WARN[0000] The "POSTGRES_SUPERUSER" variable is not set. Defaulting to a blank string.
```

**해결 방법**:

1. 환경변수 파일이 있는지 확인: `ls -la docker/env/.env.dev`
2. 스크립트를 사용하거나 `--env-file` 옵션을 추가하세요

자세한 내용은 [Docker 가이드](./docker/README.md)를 참조하세요.

## 🛠️ 개발 가이드

### 디렉토리 구조

```
wealist-project/
├── user-service/       # User Service (Spring Boot)
├── board-service/      # Board Service (Go)
├── frontend/           # Frontend (React)
├── docker/             # Docker 관련 파일
│   ├── compose/        # Docker Compose 파일
│   ├── env/            # 환경변수 파일
│   ├── scripts/        # 실행 스크립트 (dev.sh, prod.sh)
│   └── README.md       # Docker 가이드
├── docs/               # 프로젝트 문서
│   ├── api/            # API 레퍼런스
│   ├── guides/         # 개발 가이드
│   ├── planning/       # 계획 문서
│   └── migration/      # 마이그레이션 가이드
├── scripts/            # 유틸리티 스크립트
│   └── tests/          # 테스트 스크립트
├── CHANGELOG.md        # 변경 이력
└── README.md           # 이 파일
```

### 개발 시 주의사항

- **Board Service (Go)** 사용 권장
- JWT 토큰은 User Service와 Board Service 간 공유 (`SECRET_KEY` 일치 필요)
- 모든 ID는 UUID 타입 사용
- Foreign Key 없음 (샤딩 대비, 애플리케이션 레벨에서 관계 관리)
- Soft Delete 방식 (`is_deleted` 플래그)

### 추가 문서

- **API 레퍼런스**: [docs/api/](./docs/api/)
  - [Board Service API](./docs/api/board-service-api.md)
  - [User Service API](./docs/api/user-service-api.md)
- **개발 가이드**: [docs/guides/](./docs/guides/)
- **Docker 가이드**: [docker/README.md](./docker/README.md)
- **테스트 가이드**: [scripts/tests/README.md](./scripts/tests/README.md)
- **전체 문서 목록**: [docs/README.md](./docs/README.md)

## 📦 기술 스택

### Backend
- **User Service**: Spring Boot 3.x, Java 17, Spring Security, JWT
- **Board Service**: Go 1.21+, Gin, GORM, Viper, Zap Logger

### Database & Cache
- **PostgreSQL 17**: 각 서비스별 독립 DB
- **Redis 7**: 캐싱 및 세션 관리

### Frontend
- **React 18**: TypeScript, Tailwind CSS

### DevOps
- **Docker & Docker Compose**: 컨테이너 오케스트레이션
- **Git**: 모노레포 구조

