# Project Board Management API

프로젝트 관리 도구의 Board 관리 시스템 RESTful API

<!-- Updated: 2025-11-16 - Docker Compose v1 compatibility -->

## 목차

- [소개](#소개)
- [기술 스택](#기술-스택)
- [주요 기능](#주요-기능)
- [시작하기](#시작하기)
  - [사전 요구사항](#사전-요구사항)
  - [설치](#설치)
  - [실행](#실행)
- [API 엔드포인트](#api-엔드포인트)
- [환경 설정](#환경-설정)
- [개발](#개발)
- [Docker](#docker)
- [프로젝트 구조](#프로젝트-구조)
- [아키텍처](#아키텍처)
- [라이선스](#라이선스)

## 소개

Project Board Management API는 프로젝트 관리 도구의 핵심 기능을 제공하는 RESTful API입니다. Workspace 내의 Project에 속한 Board들을 관리하며, 각 Board는 Stage(진행 상태), Importance(중요도), Role(담당자 역할) 속성을 가지고 참여자 및 댓글 기능을 제공합니다.

### 특징

- Clean Architecture 기반 설계로 유지보수성과 테스트 용이성 확보
- RESTful API 설계 원칙 준수
- Soft Delete 정책으로 데이터 복구 가능
- 구조화된 로깅 (Zap)으로 운영 모니터링 용이
- JWT 기반 인증 지원
- Docker 및 Docker Compose 지원으로 간편한 배포

## 기술 스택

- **Language**: Go 1.21+
- **Framework**: [Gin](https://gin-gonic.com/) - 고성능 HTTP 웹 프레임워크
- **ORM**: [GORM](https://gorm.io/) - Go용 ORM 라이브러리
- **Database**: PostgreSQL 14+
- **Logging**: [Zap](https://github.com/uber-go/zap) - 구조화된 고성능 로깅
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) - 구조체 검증
- **Configuration**: [Viper](https://github.com/spf13/viper) - 설정 관리

## 주요 기능

### Project 관리
- Workspace별 프로젝트 생성 및 조회
- Default 프로젝트 지원

### Board 관리
- Board 생성, 조회, 수정, 삭제 (CRUD)
- Stage 관리: 진행중, 승인대기, 승인, 재검토
- Importance 설정: 긴급, 보통
- Role 지정: 개발자, 기획자

### 참여자 관리
- Board에 참여자 추가/제거
- 참여자 목록 조회
- 중복 참여 방지

### 댓글 기능
- Board에 댓글 작성, 수정, 삭제
- 댓글 목록 조회 (작성 시간 순)

## 시작하기

### 사전 요구사항

- **Go**: 1.21 이상
- **PostgreSQL**: 14 이상
- **Make**: (선택사항, 편의 명령어 사용 시)
- **Docker**: (선택사항, Docker 사용 시)

### 설치

#### 1. 저장소 클론

```bash
git clone <repository-url>
cd project-board-api
```

#### 2. 의존성 설치

```bash
go mod download
```

또는 Make를 사용하는 경우:

```bash
make deps
```

#### 3. 환경 설정

애플리케이션 설정은 두 가지 방법으로 가능합니다:

**방법 1: 환경 변수 사용 (권장)**

```bash
# .env 파일 생성
cp .env.example .env

# .env 파일을 편집하여 환경변수 설정
# 최소한 다음 항목들을 설정하세요:
# - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
# - JWT_SECRET (프로덕션에서는 반드시 변경!)
```

**방법 2: YAML 설정 파일 사용**

```bash
# config.yaml 파일 생성
cp configs/config.yaml.example configs/config.yaml

# config.yaml 파일을 편집하여 설정
```

> **참고**: 환경 변수가 YAML 설정보다 우선순위가 높습니다.

자세한 설정 가이드는 [docs/CONFIGURATION.md](docs/CONFIGURATION.md)를 참조하세요.

#### 4. 데이터베이스 설정

**데이터베이스 생성**

```bash
# PostgreSQL에 접속하여 데이터베이스 생성
make db-create

# 또는 직접 실행
createdb -U postgres project_board
```

**마이그레이션 실행**

```bash
make migrate-up
```

마이그레이션에 대한 자세한 내용은 [migrations/README.md](migrations/README.md)를 참조하세요.

### 실행
#### 개발 모드

```bash
# Go로 직접 실행
make run

# 또는
go run cmd/api/main.go
```

#### 프로덕션 모드

```bash
# 빌드
make build

# 실행
./bin/main
```

#### Hot Reload 개발 모드 (air 사용)

```bash
# air가 설치되어 있지 않으면 자동으로 설치됩니다
make dev
```

서버가 정상적으로 시작되면 `http://localhost:8000`에서 API에 접근할 수 있습니다.

#### Health Check

```bash
curl http://localhost:8000/health
```

## API 문서

### Swagger UI

API 문서는 Swagger UI를 통해 인터랙티브하게 확인할 수 있습니다.

서버 실행 후 다음 URL에 접속하세요:

```
http://localhost:8000/swagger/index.html
```

Swagger UI에서는 다음 기능을 제공합니다:
- 모든 API 엔드포인트 목록 및 상세 설명
- 요청/응답 스키마 확인
- API 직접 테스트 (Try it out)
- 예시 요청/응답 확인

### Presigned URL 기반 파일 업로드 API

Board, Comment, Project에 파일을 첨부할 수 있는 Presigned URL 기반 업로드 API를 제공합니다.

**주요 특징:**
- 클라이언트가 S3에 직접 업로드하여 서버 부하 최소화
- 이미지 및 문서 파일 지원 (최대 20MB)
- 임시 파일 자동 정리 (1시간 후)

**지원 파일 형식:**
- 이미지: jpg, jpeg, png, gif, webp
- 문서: pdf, txt, doc, docx, xls, xlsx, ppt, pptx

**API 엔드포인트:**
- `POST /api/attachments/presigned-url` - Presigned URL 생성
- `POST /api/attachments` - 첨부파일 메타데이터 저장

**상세 가이드:** [docs/PRESIGNED_URL_API_GUIDE.md](docs/PRESIGNED_URL_API_GUIDE.md)를 참조하세요.

### API 마이그레이션 가이드

API 표준화 작업으로 인해 엔드포인트와 필드명이 변경되었습니다. 자세한 내용은 [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)를 참조하세요.

### Swagger 문서 재생성

코드 변경 후 Swagger 문서를 업데이트하려면:

```bash
# Swagger 문서 생성 (의존성 및 내부 패키지 파싱 포함)
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# 또는 Make 명령어 사용
make swagger
```

**중요**: `--parseDependency`와 `--parseInternal` 플래그를 반드시 포함해야 모든 DTO 스키마가 올바르게 생성됩니다.

### Swagger 문서 검증

Swagger 문서가 최신 상태인지 확인하려면:

```bash
# 엔드포인트 커버리지 검증
./scripts/validate-swagger.sh

# DTO 스키마 검증
./scripts/validate-dto-schemas.sh

# Godoc 주석 품질 검증
./scripts/validate-godoc.sh
```

모든 검증 스크립트는 100% 커버리지를 목표로 하며, 누락된 항목이 있으면 오류와 함께 종료됩니다.

### Swagger 문서 유지보수 가이드

API 개발 시 Swagger 문서를 최신 상태로 유지하는 방법:

1. **새 핸들러 추가 시**: 반드시 godoc 주석 추가
2. **DTO 변경 시**: 변경 후 `swag init` 실행
3. **라우터 변경 시**: 핸들러의 `@Router` 주석 업데이트
4. **커밋 전**: 검증 스크립트 실행하여 문서 완전성 확인

#### Godoc 주석 표준

모든 핸들러 함수는 다음 형식의 godoc 주석을 포함해야 합니다:

```go
// HandlerName godoc
// @Summary      간단한 요약 (한 줄)
// @Description  상세한 설명
// @Tags         tag-name
// @Accept       json
// @Produce      json
// @Param        paramName paramType dataType required "description"
// @Success      200 {object} response.SuccessResponse{data=dto.ResponseType}
// @Failure      400 {object} response.ErrorResponse "error message"
// @Failure      404 {object} response.ErrorResponse "not found"
// @Failure      500 {object} response.ErrorResponse "internal error"
// @Router       /api/path [method]
func (h *Handler) HandlerName(c *gin.Context) {
    // ...
}
```

**필수 태그:**
- `@Summary`: 엔드포인트 요약 (필수)
- `@Tags`: API 그룹화 태그 (필수)
- `@Router`: 라우트 경로 및 메서드 (필수)

**권장 태그:**
- `@Description`: 상세 설명
- `@Param`: 모든 파라미터 문서화
- `@Success`: 성공 응답 (상태 코드별)
- `@Failure`: 에러 응답 (모든 가능한 상태 코드)

자세한 가이드는 다음 문서를 참조하세요:
- [Swagger 문서 가이드](docs/SWAGGER.md) - Swagger 생성 및 검증
- [개발자 가이드](docs/DEVELOPER_GUIDELINES.md) - 전체 개발 워크플로우
- [CI/CD 통합 가이드](docs/CI_CD_INTEGRATION.md) - CI/CD 파이프라인 및 자동화

## API 엔드포인트

### Base URL

```
http://localhost:8000/api
```

> **참고**: API 표준화 작업으로 base path가 `/api/v1`에서 `/api`로 변경되었습니다. 자세한 내용은 [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)를 참조하세요.

### Projects

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/projects` | 프로젝트 생성 | `CreateProjectRequest` | `ProjectResponse` |
| GET | `/projects/workspace/:workspaceId` | Workspace의 프로젝트 목록 조회 | - | `[]ProjectResponse` |
| GET | `/projects/workspace/:workspaceId/default` | Workspace의 default 프로젝트 조회 | - | `ProjectResponse` |

**CreateProjectRequest 예시:**
```json
{
  "workspaceId": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Project",
  "description": "Project description",
  "isDefault": false
}
```

### Boards

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/boards` | Board 생성 | `CreateBoardRequest` | `BoardResponse` |
| GET | `/boards/:boardId` | Board 상세 조회 | - | `BoardDetailResponse` |
| GET | `/boards/project/:projectId` | Project의 Board 목록 조회 | - | `[]BoardResponse` |
| PUT | `/boards/:boardId` | Board 수정 | `UpdateBoardRequest` | `BoardResponse` |
| DELETE | `/boards/:boardId` | Board 삭제 | - | Success message |

**CreateBoardRequest 예시:**
```json
{
  "projectId": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Implement User Authentication",
  "content": "Add JWT-based authentication system",
  "stage": "in_progress",
  "importance": "urgent",
  "role": "developer"
}
```

**Stage 값**: `in_progress`, `pending`, `approved`, `review`
**Importance 값**: `urgent`, `normal`
**Role 값**: `developer`, `planner`

### Participants

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/participants` | 참여자 추가 | `AddParticipantRequest` | Success message |
| GET | `/participants/board/:boardId` | Board 참여자 목록 조회 | - | `[]ParticipantResponse` |
| DELETE | `/participants/board/:boardId/user/:userId` | 참여자 제거 | - | Success message |

**AddParticipantRequest 예시:**
```json
{
  "boardId": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "660e8400-e29b-41d4-a716-446655440000"
}
```

### Comments

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/comments` | 댓글 작성 | `CreateCommentRequest` | `CommentResponse` |
| GET | `/comments/board/:boardId` | Board 댓글 목록 조회 | - | `[]CommentResponse` |
| PUT | `/comments/:commentId` | 댓글 수정 | `UpdateCommentRequest` | `CommentResponse` |
| DELETE | `/comments/:commentId` | 댓글 삭제 | - | Success message |

**CreateCommentRequest 예시:**
```json
{
  "boardId": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "660e8400-e29b-41d4-a716-446655440000",
  "content": "This looks good to me!"
}
```

### 응답 형식

**성공 응답:**
```json
{
  "data": { ... },
  "message": "Success message"
}
```

**에러 응답:**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description",
    "details": "Additional details"
  },
  "message": "User-friendly error message"
}
```

**HTTP 상태 코드:**
- `200 OK`: 성공
- `201 Created`: 리소스 생성 성공
- `400 Bad Request`: 잘못된 요청 (검증 실패)
- `404 Not Found`: 리소스를 찾을 수 없음
- `409 Conflict`: 충돌 (예: 중복 참여자)
- `500 Internal Server Error`: 서버 내부 오류

## 환경 설정

애플리케이션은 두 가지 방법으로 설정할 수 있습니다:

1. **환경 변수** (`.env` 파일 또는 시스템 환경 변수)
2. **YAML 설정 파일** (`configs/config.yaml`)

환경 변수가 YAML 설정보다 우선순위가 높습니다.

### 환경 변수 형식 지원

Board-service는 wealist-project 환경과의 호환성을 위해 두 가지 환경 변수 형식을 지원합니다:

**원본 형식 (wealist-project 호환):**
- `DATABASE_URL`: PostgreSQL 연결 문자열
- `SECRET_KEY`: JWT 서명 키
- `USER_SERVICE_URL`: 사용자 서비스 엔드포인트
- `ENV`: 환경 모드 (dev/prod)
- `CORS_ORIGINS`: 허용된 CORS 출처

**현재 형식:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: 개별 데이터베이스 설정
- `JWT_SECRET`: JWT 서명 키
- `USER_API_BASE_URL`: 사용자 서비스 엔드포인트
- `SERVER_MODE`: 서버 모드 (debug/release)
- `CORS_ALLOWED_ORIGINS`: 허용된 CORS 출처

두 형식 모두 완전히 지원되며, 두 형식이 모두 제공되는 경우 원본 형식이 우선합니다.

### 주요 환경 변수

**원본 형식 (권장):**
```bash
# Server Configuration
SERVER_PORT=8000              # 서버 포트 (기본값: 8000)
ENV=dev                       # 환경 모드: dev 또는 prod

# Database Configuration (연결 문자열 형식)
DATABASE_URL=postgresql://postgres:password@localhost:5432/project_board?sslmode=disable

# JWT Configuration
SECRET_KEY=your-secret-key    # JWT 서명 키 (프로덕션에서 반드시 변경!)

# User Service Configuration
USER_SERVICE_URL=http://user-service:8080

# CORS Configuration
CORS_ORIGINS=http://localhost:3000

# Logger Configuration
LOG_LEVEL=info                # 로그 레벨: debug, info, warn, error

# S3 Configuration
S3_BUCKET=wealist-dev-files   # S3 버킷 이름
S3_REGION=ap-northeast-2      # S3 리전
# S3_ENDPOINT=http://localhost:9000  # MinIO 사용 시에만 설정
# S3_ACCESS_KEY=minioadmin           # MinIO 사용 시에만 설정
# S3_SECRET_KEY=minioadmin           # MinIO 사용 시에만 설정
```

**현재 형식 (하위 호환성):**
```bash
# Server Configuration
SERVER_PORT=8000              # 서버 포트 (기본값: 8000)
SERVER_MODE=debug             # 서버 모드: debug 또는 release

# Database Configuration (개별 변수)
DB_HOST=localhost             # PostgreSQL 호스트
DB_PORT=5432                  # PostgreSQL 포트
DB_USER=postgres              # 데이터베이스 사용자
DB_PASSWORD=password          # 데이터베이스 비밀번호
DB_NAME=project_board         # 데이터베이스 이름
DB_MAX_OPEN_CONNS=25          # 최대 오픈 커넥션 수
DB_MAX_IDLE_CONNS=5           # 최대 유휴 커넥션 수

# Logger Configuration
LOG_LEVEL=info                # 로그 레벨: debug, info, warn, error
LOG_OUTPUT_PATH=stdout        # 로그 출력: stdout, stderr, 또는 파일 경로

# JWT Configuration
JWT_SECRET=your-secret-key    # JWT 서명 키 (프로덕션에서 반드시 변경!)
JWT_EXPIRE_TIME=24h           # 토큰 만료 시간 (예: 24h, 7d)

# User Service Configuration
USER_API_BASE_URL=http://user-service:8080

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000

# S3 Configuration
S3_BUCKET=wealist-dev-files   # S3 버킷 이름
S3_REGION=ap-northeast-2      # S3 리전
# S3_ENDPOINT=http://localhost:9000  # MinIO 사용 시에만 설정
# S3_ACCESS_KEY=minioadmin           # MinIO 사용 시에만 설정
# S3_SECRET_KEY=minioadmin           # MinIO 사용 시에만 설정
```

### S3 자격증명 설정

**AWS 환경 (EC2):**
- IAM 역할 사용 (자격증명 불필요)
- EC2 인스턴스에 S3 접근 권한이 있는 IAM 역할 할당

**로컬 환경:**
- `~/.aws/credentials` 파일 사용 (권장)
- 또는 환경 변수 사용: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`

```bash
# ~/.aws/credentials 파일 설정
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY

# ~/.aws/config 파일 설정
[default]
region = ap-northeast-2
```

**MinIO 사용 (로컬 테스트):**
- `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` 환경 변수 설정 필요

```bash
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
```

### 설정 우선순위

1. 환경 변수 (최우선)
2. `.env` 파일
3. `configs/config.yaml` 파일
4. 기본값

### User Service 연동 설정 검증

board-service는 user-service와 통신하기 위해 올바른 base URL 설정이 필요합니다.

**설정 확인 방법:**

1. **애플리케이션 시작 로그 확인**
   ```
   User API Configuration validated:
     - Base URL: http://user-service:8080
     - Scheme: http
     - Host: user-service:8080
     - Timeout: 5s
     - Source: USER_SERVICE_URL environment variable
   ```

2. **엔드포인트 예시 로그 확인**
   ```
   User API endpoint examples (for debugging):
     - validate_member: http://user-service:8080/api/workspaces/{workspaceId}/validate-member/{userId}
     - get_user: http://user-service:8080/api/users/{userId}
     - get_workspace_profile: http://user-service:8080/api/profiles/workspace/{workspaceId}
     - get_workspace: http://user-service:8080/api/workspaces/{workspaceId}
   ```

**일반적인 설정 오류:**

1. **Trailing slash 문제**
   - ❌ 잘못된 설정: `USER_SERVICE_URL=http://user-service:8080/`
   - ✅ 올바른 설정: `USER_SERVICE_URL=http://user-service:8080`
   - 시스템이 자동으로 trailing slash를 제거하고 경고를 표시합니다.

2. **Docker 환경에서 localhost 사용**
   - ❌ 잘못된 설정: `USER_SERVICE_URL=http://localhost:8080` (Docker 컨테이너 내부)
   - ✅ 올바른 설정: `USER_SERVICE_URL=http://user-service:8080` (Docker service name)
   - ✅ 로컬 개발: `USER_SERVICE_URL=http://localhost:8080` (호스트에서 직접 실행 시)

3. **포트 번호 누락**
   - ❌ 잘못된 설정: `USER_SERVICE_URL=http://user-service`
   - ✅ 올바른 설정: `USER_SERVICE_URL=http://user-service:8080`

**환경별 권장 설정:**

- **로컬 개발 (호스트에서 직접 실행):**
  ```bash
  USER_SERVICE_URL=http://localhost:8080
  ```

- **Docker Compose:**
  ```bash
  USER_SERVICE_URL=http://user-service:8080
  ```

- **Kubernetes:**
  ```bash
  USER_SERVICE_URL=http://user-service.default.svc.cluster.local:8080
  ```

### 프로덕션 환경 설정

프로덕션 환경에서는 다음 사항을 반드시 확인하세요:

- `SERVER_MODE=release` 설정
- `JWT_SECRET`을 강력한 랜덤 값으로 변경
- `LOG_LEVEL=info` 또는 `warn` 설정
- 데이터베이스 비밀번호를 안전하게 관리
- HTTPS 사용 (리버스 프록시 설정)
- `USER_SERVICE_URL`이 올바른 내부 서비스 주소를 가리키는지 확인

**자세한 설정 가이드**: [docs/CONFIGURATION.md](docs/CONFIGURATION.md)를 참조하세요.


## 개발

### 사용 가능한 Make 명령어

```bash
make help              # 사용 가능한 모든 명령어 표시
```

### 빌드

```bash
make build             # 애플리케이션 빌드 (bin/main)
make build-linux       # Linux용 빌드 (Docker용)
```

### 테스트

```bash
make test              # 모든 테스트 실행
make test-short        # 빠른 테스트 (race detector 없이)
make test-coverage     # 테스트 커버리지 리포트 (HTML)
make test-coverage-text # 테스트 커버리지 (텍스트)
```

### 통합 테스트

실제 user-service와 board-service를 사용한 통합 테스트:

```bash
# 전체 통합 테스트 실행 (user-service 로그인 포함)
./board-service/scripts/integration-test.sh

# 빠른 단일 엔드포인트 테스트
./board-service/scripts/quick-test.sh projects
```


### 코드 품질

```bash
make fmt               # 코드 포맷팅
make vet               # go vet 실행
make lint              # golangci-lint 실행
make check             # fmt + vet + lint 실행
```

### 데이터베이스

```bash
make db-create         # 데이터베이스 생성
make db-drop           # 데이터베이스 삭제 (주의!)
make db-reset          # 데이터베이스 재설정 (drop + create + migrate)
make migrate-up        # 마이그레이션 실행
make migrate-down      # 마이그레이션 롤백
make migrate-status    # 마이그레이션 상태 확인
```

### 의존성 관리

```bash
make deps              # 의존성 다운로드 및 정리
make deps-upgrade      # 모든 의존성 업그레이드
```

### 개발 도구 설치

```bash
make install-tools     # 개발 도구 설치 (air, golangci-lint)
```

## Docker

### Docker로 실행하기

#### 방법 1: Docker Compose 사용 (권장)

가장 간단한 방법으로, PostgreSQL과 API 서버를 함께 실행합니다.

```bash
# 서비스 시작
make docker-compose-up

# 또는
docker-compose up -d

# 로그 확인
make docker-compose-logs

# 서비스 중지
make docker-compose-down
```

API는 `http://localhost:8000`에서 접근 가능합니다.

#### 방법 2: Docker 단독 사용

```bash
# Docker 이미지 빌드
make docker-build

# Docker 컨테이너 실행
make docker-run

# 또는 인터랙티브 모드로 실행
make docker-run-interactive

# 로그 확인
make docker-logs

# 컨테이너 중지
make docker-stop
```

### Docker 이미지 정보

- **Base Image**: Alpine Linux (경량)
- **Multi-stage Build**: 최적화된 이미지 크기
- **Non-root User**: 보안 강화
- **Health Check**: 자동 헬스 체크 포함

### Docker 환경 변수

Docker Compose를 사용할 때는 `.env` 파일의 환경 변수가 자동으로 적용됩니다.

```bash
# .env 파일 생성
cp .env.example .env

# 필요한 값 수정 후 실행
docker-compose up -d
```

## 프로젝트 구조

```
project-board-api/
├── cmd/
│   └── api/
│       └── main.go                 # 애플리케이션 진입점
├── internal/
│   ├── config/                     # 설정 관리
│   │   └── config.go
│   ├── database/                   # 데이터베이스 연결
│   │   └── database.go
│   ├── domain/                     # 도메인 모델
│   │   ├── base.go
│   │   ├── project.go
│   │   ├── board.go
│   │   ├── participant.go
│   │   └── comment.go
│   ├── dto/                        # 요청/응답 DTO
│   │   ├── project_dto.go
│   │   ├── board_dto.go
│   │   ├── participant_dto.go
│   │   └── comment_dto.go
│   ├── repository/                 # 데이터 접근 계층
│   │   ├── project_repository.go
│   │   ├── board_repository.go
│   │   ├── participant_repository.go
│   │   └── comment_repository.go
│   ├── service/                    # 비즈니스 로직 계층
│   │   ├── project_service.go
│   │   ├── board_service.go
│   │   ├── participant_service.go
│   │   └── comment_service.go
│   ├── handler/                    # HTTP 핸들러
│   │   ├── project_handler.go
│   │   ├── board_handler.go
│   │   ├── participant_handler.go
│   │   ├── comment_handler.go
│   │   └── error_handler.go
│   ├── middleware/                 # 미들웨어
│   │   ├── auth.go                # JWT 인증
│   │   ├── logger.go              # 요청 로깅
│   │   ├── recovery.go            # Panic 복구
│   │   └── cors.go                # CORS 설정
│   ├── logger/                     # 로거 설정
│   │   └── logger.go
│   └── response/                   # 공통 응답 헬퍼
│       ├── response.go
│       └── error.go
├── configs/
│   ├── config.yaml                 # 설정 파일
│   ├── config.yaml.example         # 설정 파일 예시
│   └── README.md                   # 설정 가이드
├── migrations/                     # DB 마이그레이션
│   ├── 001_init_schema.sql
│   ├── 001_init_schema_down.sql
│   └── README.md
├── docs/
│   └── CONFIGURATION.md            # 상세 설정 가이드
├── .env.example                    # 환경 변수 예시
├── .gitignore
├── docker-compose.yml              # Docker Compose 설정
├── Dockerfile                      # Docker 이미지 빌드
├── Makefile                        # Make 명령어
├── go.mod                          # Go 모듈 정의
├── go.sum                          # Go 모듈 체크섬
└── README.md                       # 이 파일
```

## 아키텍처

### Clean Architecture

이 프로젝트는 Clean Architecture 패턴을 따릅니다:

```
Handler → Service → Repository → Domain
   ↓         ↓          ↓
  DTO    Interface   GORM
```

**계층별 책임:**

- **Handler**: HTTP 요청/응답 처리, DTO 바인딩 및 검증
- **Service**: 비즈니스 로직 구현, 트랜잭션 관리
- **Repository**: 데이터 접근, GORM 쿼리 실행
- **Domain**: 도메인 모델 정의, 비즈니스 규칙

### 의존성 방향

- Handler는 Service 인터페이스에 의존
- Service는 Repository 인터페이스에 의존
- Repository는 Domain 모델에 의존
- Domain은 어떤 계층도 import하지 않음

### 주요 설계 결정

1. **UUID 사용**: 모든 엔티티 ID를 UUID로 사용하여 분산 시스템 확장 가능
2. **Soft Delete**: 모든 삭제 작업을 소프트 삭제로 처리하여 데이터 복구 가능
3. **Interface 기반**: Repository와 Service를 인터페이스로 정의하여 테스트 용이성 확보
4. **Context 전파**: 모든 계층에 context.Context를 전달하여 timeout 제어 및 추적 가능

### 데이터베이스 스키마

주요 테이블:
- `projects`: 프로젝트 정보
- `boards`: Board 정보 (Stage, Importance, Role 포함)
- `participants`: Board 참여자 (board_id, user_id unique constraint)
- `comments`: Board 댓글

모든 테이블은 soft delete를 위한 `deleted_at` 컬럼을 포함합니다.

## 트러블슈팅

### 데이터베이스 연결 실패

```bash
# PostgreSQL이 실행 중인지 확인
pg_isready -h localhost -p 5432

# 데이터베이스가 존재하는지 확인
psql -U postgres -l | grep project_board

# 데이터베이스 재생성
make db-reset
```

### 포트 충돌

```bash
# 8000 포트를 사용 중인 프로세스 확인
lsof -i :8000

# 다른 포트 사용
SERVER_PORT=8001 make run
```

### 마이그레이션 실패

```bash
# 마이그레이션 상태 확인
make migrate-status

# 마이그레이션 롤백 후 재실행
make migrate-down
make migrate-up
```

## 기여하기

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 라이선스

MIT License

## 문의

프로젝트에 대한 문의사항이나 버그 리포트는 GitHub Issues를 이용해주세요.


