# weAlist 프로젝트 변경 이력

## 📋 최근 주요 변경사항 요약

이 문서는 weAlist 프로젝트의 최근 주요 변경사항을 정리한 문서입니다.

### 🎯 핵심 변경 사항

**프로젝트 통계:**
- 변경된 파일: 473개
- 추가된 코드: +91,784줄
- 삭제된 코드: -30,007줄
- 순증가: +61,777줄

---

## 🏗️ 1. CI/CD 파이프라인 구축

### GitHub Actions 워크플로우 추가

#### 개발 환경 CI/CD
- **Board Service CI/CD**
  - `ci-dev-board-service.yml` - 빌드 및 테스트 자동화
  - `cd-dev-board-service.yml` - EC2 자동 배포

- **User Service CI/CD**
  - `ci-dev-user-service.yml` - Java 빌드 및 테스트
  - `cd-dev-user-service.yml` - EC2 자동 배포

- **Frontend CI/CD**
  - `frontend-cicd.yml` - 프론트엔드 빌드 및 배포

- **모니터링**
  - `cd-dev-monitoring.yml` - Prometheus & Grafana 배포

#### 주요 기능
- 자동화된 빌드 및 테스트
- Docker 이미지 빌드 및 ECR 푸시
- EC2 인스턴스 자동 배포
- Parameter Store를 통한 보안 설정 관리

**관련 문서:**
- [GitHub Actions README](.github/workflows/README.md)
- [GitHub Secrets 설정](docs/GITHUB_SECRETS_SETUP.md)
- [Parameter Store 설정](docs/PARAMETER_STORE_SETUP.md)

---

## 🔧 2. Board Service (Go) 대규모 개선

### 2.1 새로운 기능 추가

#### Attachment (첨부파일) 시스템
- **파일 업로드/다운로드**: S3 기반 첨부파일 관리
- **엔티티**: `Attachment` 도메인 모델 추가
- **API 엔드포인트**:
  - 프로젝트 첨부파일 업로드/조회/삭제
  - 보드 첨부파일 업로드/조회/삭제
- **S3 클라이언트**: `internal/client/s3_client.go`
- **통합 테스트**: 전체 첨부파일 플로우 테스트 포함

**관련 파일:**
```
board-service/internal/domain/attachment.go
board-service/internal/handler/attachment_handler.go
board-service/internal/repository/attachment_repository.go
board-service/internal/service/attachment_helper.go
board-service/internal/client/s3_client.go
```

#### Comment (댓글) 시스템 개선
- **기능 강화**: 댓글 CRUD 완전 구현
- **엔티티 개선**: `Comment` 도메인 모델 업데이트
- **서비스 레이어**: 댓글 비즈니스 로직 강화
- **테스트**: 통합 테스트 및 단위 테스트 추가

**관련 파일:**
```
board-service/internal/domain/comment.go
board-service/internal/handler/comment_handler.go
board-service/internal/repository/comment_repository.go
board-service/internal/service/comment_service.go
```

#### WebSocket 실시간 통신
- **WebSocket 핸들러**: `internal/handler/ws_handler.go`
- **실시간 업데이트**: 보드/프로젝트 변경사항 실시간 동기화
- **연결 관리**: WebSocket 연결 및 메시지 브로드캐스팅

#### Project Member & Participant 관리
- **프로젝트 멤버 관리**: 멤버 추가/제거/권한 관리
- **참여자 관리**: 보드 참여자 관리
- **가입 요청**: 프로젝트 가입 요청 처리

**관련 파일:**
```
board-service/internal/handler/project_member_handler.go
board-service/internal/handler/participant_handler.go
board-service/internal/handler/project_join_request_handler.go
board-service/internal/repository/project_member_repository.go
board-service/internal/repository/participant_repository.go
```

### 2.2 아키텍처 개선

#### Custom Field → Field Option 리팩토링
- **구조 개선**: 커스텀 필드 시스템을 Field Option으로 재설계
- **유연성 향상**: 더 확장 가능한 필드 옵션 구조
- **타입 안전성**: 타입 안전한 필드 옵션 처리

**변경 파일:**
```diff
- internal/handler/custom_field_handler.go
- internal/service/custom_field_service.go
+ internal/handler/field_option_handler.go
+ internal/service/field_option_service.go
+ internal/converter/field_option_converter.go
```

#### Repository 패턴 개선
- **Base Repository**: 공통 CRUD 로직 추상화
- **인터페이스 정의**: `internal/repository/interfaces.go`
- **타입 안전성**: 제네릭을 활용한 타입 안전한 Repository

**관련 파일:**
```
board-service/internal/repository/base/base_repository.go
board-service/internal/repository/interfaces.go
```

#### Unit of Work 패턴 도입
- **트랜잭션 관리**: UoW 패턴으로 트랜잭션 일관성 보장
- **서비스 레이어**: `board_service_with_uow.go`

**관련 파일:**
```
board-service/internal/uow/unit_of_work.go
board-service/internal/uow/example.go
board-service/internal/service/board_service_with_uow.go
```

### 2.3 모니터링 & 메트릭스

#### Prometheus 메트릭스 수집
- **비즈니스 메트릭**: 프로젝트/보드 생성, 업데이트 추적
- **HTTP 메트릭**: 요청 지연시간, 상태 코드 분포
- **데이터베이스 메트릭**: 쿼리 실행 시간, 연결 풀 상태
- **외부 API 메트릭**: User Service 호출 추적

**메트릭 수집 모듈:**
```
board-service/internal/metrics/metrics.go
board-service/internal/metrics/business.go
board-service/internal/metrics/http.go
board-service/internal/metrics/database.go
board-service/internal/metrics/external.go
```

#### Grafana 대시보드
- **비즈니스 대시보드**: 비즈니스 KPI 모니터링
- **성능 대시보드**: 응답 시간 및 처리량
- **개발자 대시보드**: 에러율, 디버깅 정보
- **DevOps 대시보드**: 인프라 상태 모니터링

**대시보드 정의:**
```
prometheus/dashboards/board-service-business.json
prometheus/dashboards/board-service-performance.json
prometheus/dashboards/board-service-developer.json
prometheus/dashboards/board-service-devops.json
```

### 2.4 테스트 강화

#### 테스트 커버리지 대폭 확대
- **단위 테스트**: 모든 핸들러, 서비스, Repository 테스트
- **통합 테스트**: 전체 API 플로우 테스트
- **Property-based Testing**: 속성 기반 테스트 도입

**테스트 파일 (일부):**
```
board-service/internal/handler/*_test.go (17개 파일)
board-service/internal/service/*_test.go (12개 파일)
board-service/internal/repository/*_test.go (8개 파일)
board-service/internal/client/*_test.go (2개 파일)
```

#### 테스트 유틸리티
- **Test Fixtures**: 테스트 데이터 생성 헬퍼
- **Mock 객체**: 의존성 모킹
- **Database Helper**: 테스트 DB 설정

**관련 파일:**
```
board-service/internal/testutil/database.go
board-service/internal/testutil/fixtures.go
board-service/internal/testutil/mocks.go
```

### 2.5 설정 및 문서화

#### 설정 관리 개선
- **YAML 설정**: `configs/config.yaml.example`
- **환경별 설정**: local, dev, ec2-dev, prod
- **Viper 기반**: 계층적 설정 관리

**관련 문서:**
```
board-service/configs/README.md
board-service/docs/CONFIGURATION.md
```

#### 개발자 문서 추가
- **아키텍처 문서**: `ARCHITECTURE.md` - 전체 아키텍처 설명
- **캐시 전략**: `CACHE_STRATEGY.md` - 캐싱 전략
- **개발 가이드라인**: `docs/DEVELOPER_GUIDELINES.md`
- **CI/CD 통합**: `docs/CI_CD_INTEGRATION.md`
- **Swagger 가이드**: `docs/SWAGGER.md`

---

## 👤 3. User Service (Spring Boot) 개선

### 3.1 프로필 이미지 업로드 기능

#### Presigned URL 방식
- **2단계 업로드 프로세스**:
  1. Presigned URL 요청
  2. 클라이언트가 S3에 직접 업로드
  3. 업로드 완료 후 메타데이터 저장

**API 엔드포인트:**
```
POST /api/profile-images/presigned-url - Presigned URL 생성
POST /api/profile-images/save - 업로드 완료 후 메타데이터 저장
GET  /api/profile-images/{attachmentId} - 이미지 조회
DELETE /api/profile-images/{attachmentId} - 이미지 삭제
```

**관련 파일:**
```
user-service/src/main/java/OrangeCloud/UserRepo/controller/ProfileImageController.java
user-service/src/main/java/OrangeCloud/UserRepo/service/S3Service.java
user-service/src/main/java/OrangeCloud/UserRepo/service/AttachmentService.java
```

**문서:**
- [Presigned URL 가이드](user-service/docs/PRESIGNED_URL_PROFILE_IMAGE_GUIDE.md)
- [Attachment 임시 파일 가이드](user-service/docs/ATTACHMENT_TEMPORARY_FILE_GUIDE.md)

### 3.2 Attachment 엔티티 추가

#### 첨부파일 관리
- **엔티티**: `Attachment.java`
- **상태 관리**: `TEMPORARY`, `PERMANENT`, `DELETED`
- **자동 정리**: 임시 파일 자동 삭제 Job

**관련 파일:**
```
user-service/src/main/java/OrangeCloud/UserRepo/entity/Attachment.java
user-service/src/main/java/OrangeCloud/UserRepo/repository/AttachmentRepository.java
user-service/src/main/java/OrangeCloud/UserRepo/job/AttachmentCleanupJob.java
```

### 3.3 Board Service 클라이언트

#### 마이크로서비스 간 통신
- **RestTemplate 기반**: Board Service API 호출
- **DTO 정의**: 프로젝트, 보드, 댓글 생성 요청/응답
- **샘플 데이터 생성**: 개발 환경 초기 데이터 자동 생성

**관련 파일:**
```
user-service/src/main/java/OrangeCloud/UserRepo/client/BoardServiceClient.java
user-service/src/main/java/OrangeCloud/UserRepo/client/CreateProjectRequest.java
user-service/src/main/java/OrangeCloud/UserRepo/client/CreateBoardRequest.java
user-service/src/main/java/OrangeCloud/UserRepo/client/CreateCommentRequest.java
```

### 3.4 Sample Data Seeder

#### 개발 환경 초기 데이터
- **자동 생성**: 워크스페이스, 사용자, 프로젝트, 보드
- **설정 가능**: `SAMPLE_DATA_ENABLED` 환경변수로 제어
- **테스트 지원**: 통합 테스트용 데이터 생성

**관련 파일:**
```
user-service/src/main/java/OrangeCloud/UserRepo/service/SampleDataSeederService.java
user-service/src/main/java/OrangeCloud/UserRepo/util/SampleDataGenerator.java
```

### 3.5 Workspace 기능 강화

#### 워크스페이스 관리 개선
- **설정 관리**: 워크스페이스별 설정 추가
- **멤버 초대**: 멤버 초대 기능
- **유효성 검증**: 워크스페이스 접근 권한 검증

**관련 파일:**
```
user-service/src/main/java/OrangeCloud/UserRepo/service/WorkspaceService.java
user-service/src/main/java/OrangeCloud/UserRepo/dto/workspace/WorkspaceSettingsResponse.java
user-service/src/main/java/OrangeCloud/UserRepo/dto/workspace/UpdateWorkspaceSettingsRequest.java
```

### 3.6 S3 설정 개선

#### 환경별 S3 설정
- **로컬 개발**: MinIO 지원
- **AWS 환경**: S3 + IAM Role
- **이미지 URL**: 환경별 URL 생성 로직

**관련 파일:**
```
user-service/src/main/java/OrangeCloud/UserRepo/config/S3Config.java
```

---

## 🎨 4. Frontend (React) 대규모 리팩토링

### 4.1 컴포넌트 구조 개편

#### 모달 컴포넌트 재구성
**이전 구조:**
```
components/modals/
├── BoardDetailModal.tsx
├── CreateBoardModal.tsx
├── CreateProjectModal.tsx
├── CustomFieldManageModal.tsx
├── ProjectManageModal.tsx
├── TaskDetailModal.tsx
├── UserProfileModal.tsx
└── WorkspaceSettingsModal.tsx
```

**새로운 구조:**
```
components/modals/
├── board/
│   ├── BoardDetailModal.tsx (개선)
│   ├── BoardManageModal.tsx (신규)
│   ├── ProjectManageModal.tsx (개선)
│   ├── FieldManageModal.tsx
│   ├── FilterBar.tsx
│   └── customFields/
│       ├── CustomFieldManageModal.tsx (개선)
│       └── constants/colors.ts
├── user/
│   ├── UserProfileModal.tsx (개선)
│   └── wsManager/
│       ├── WorkspaceManagementModal.tsx
│       └── tabs/
│           ├── WorkspaceMembersTab.tsx
│           └── WorkspaceSettingsTab.tsx
```

#### 새로운 레이아웃 컴포넌트
- **MainLayout**: 메인 애플리케이션 레이아웃
- **ProjectContent**: 프로젝트 컨텐츠 영역
- **ProjectHeader**: 프로젝트 헤더

**관련 파일:**
```
frontend/src/components/layout/MainLayout.tsx
frontend/src/components/layout/ProjectContent.tsx
frontend/src/components/layout/ProjectHeader.tsx
```

### 4.2 새로운 기능 추가

#### Comment 시스템
- **CommentList 컴포넌트**: 댓글 목록 및 작성
- **실시간 업데이트**: WebSocket 연동

**관련 파일:**
```
frontend/src/components/comment/CommentList.tsx
```

#### 파일 업로드
- **FileUploader 컴포넌트**: 재사용 가능한 파일 업로더
- **useFileUpload Hook**: 파일 업로드 로직 훅
- **S3 직접 업로드**: Presigned URL 활용

**관련 파일:**
```
frontend/src/components/common/FileUploader.tsx
frontend/src/hooks/useFileUpload.ts
frontend/src/utils/uploadFileToS3.ts
```

#### WebSocket 통신
- **WebSocket 유틸**: 실시간 통신 유틸리티
- **연결 관리**: 자동 재연결 로직

**관련 파일:**
```
frontend/src/utils/websocket.ts
```

### 4.3 API 서비스 개선

#### Board Service API
- **타입 안전성 강화**: TypeScript 타입 정의 개선
- **에러 핸들링**: 일관된 에러 처리
- **API 엔드포인트 추가**: 새로운 기능 지원

**관련 파일:**
```
frontend/src/api/board/boardService.ts
frontend/src/types/board.ts
```

#### User Service API
- **프로필 관리**: 프로필 이미지 업로드/수정
- **워크스페이스 관리**: 워크스페이스 CRUD

**관련 파일:**
```
frontend/src/api/user/userService.ts
frontend/src/types/user.ts
```

### 4.4 UI/UX 개선

#### 공통 컴포넌트
- **AvatarStack**: 멤버 아바타 스택
- **LoadingSpinner**: 로딩 인디케이터
- **Portal**: 모달 포털 컴포넌트

**관련 파일:**
```
frontend/src/components/common/AvartarStack.tsx
frontend/src/components/common/LoadingSpinner.tsx
frontend/src/components/common/Portal.tsx
```

### 4.5 패키지 관리

#### npm → pnpm 마이그레이션
- **패키지 매니저**: npm에서 pnpm으로 전환
- **성능 개선**: 더 빠른 설치 속도
- **디스크 효율성**: 공유 패키지 저장소

**변경 사항:**
```diff
- package-lock.json (삭제)
+ pnpm-lock.yaml (추가)
```

---

## 🐳 5. Docker & DevOps 개선

### 5.1 Docker Compose 구조 개편

#### 환경별 Compose 파일 분리
**이전:**
```
docker-compose.yaml
docker-compose.base.yml
docker-compose.local.yml
```

**새로운 구조:**
```
docker/compose/
├── docker-compose.yml           # 기본 서비스 정의
├── docker-compose.dev.yml       # 로컬 개발 환경
├── docker-compose.ec2-dev.yml   # EC2 개발 환경
├── docker-compose.prod.yml      # 프로덕션 환경
└── docker-compose.monitoring.yml # Prometheus & Grafana
```

#### 환경 변수 관리
```
docker/env/
├── .env.example           # 기본 템플릿
├── .env.dev.example       # 로컬 개발
├── .env.ec2-dev.example   # EC2 개발
└── .env.prod.example      # 프로덕션
```

### 5.2 실행 스크립트

#### 환경별 스크립트
```bash
# 로컬 개발
./docker/scripts/dev.sh up
./docker/scripts/dev.sh down

# EC2 개발
./docker/scripts/ec2-dev.sh up
./docker/scripts/ec2-dev.sh down

# 프로덕션
./docker/scripts/prod.sh up
./docker/scripts/prod.sh down

# 모니터링
./docker/scripts/monitoring.sh up
./docker/scripts/monitoring.sh down
```

**관련 문서:**
- [Docker README](docker/README.md)

### 5.3 Nginx 설정

#### 환경별 Nginx 설정
- **개발 환경**: `docker/nginx/nginx.dev.conf`
- **프로덕션**: `docker/nginx/nginx.prod.conf`
- **리버스 프록시**: 서비스 라우팅
- **SSL 지원**: HTTPS 설정

### 5.4 Prometheus & Grafana

#### 모니터링 스택 추가
- **Prometheus**: 메트릭 수집
- **Grafana**: 대시보드 시각화
- **설정 파일**: `prometheus/config/prometheus-ec2.yml`

**Grafana 대시보드:**
- Board Service 비즈니스 메트릭
- Board Service 성능 메트릭
- Board Service 개발자 메트릭
- Board Service DevOps 메트릭

---

## 📚 6. 문서화 강화

### 6.1 배포 가이드

#### AWS 배포 문서
- **[ALB 라우팅 배포](docs/ALB_ROUTING_DEPLOYMENT.md)**: ALB 기반 라우팅 설정
- **[ALB 검증 가이드](docs/ALB_VERIFICATION_GUIDE.md)**: ALB 설정 검증 방법
- **[EC2 개발 배포](docs/EC2-DEV-DEPLOYMENT.md)**: EC2 환경 배포 가이드
- **[보안 배포 설정](docs/SECURE_DEPLOYMENT_SETUP.md)**: 보안 설정 가이드
- **[Parameter Store 설정](docs/PARAMETER_STORE_SETUP.md)**: AWS Parameter Store 설정
- **[Terraform Parameter Store](docs/TERRAFORM_PARAMETER_STORE_SETUP.md)**: Terraform으로 Parameter Store 관리

### 6.2 API 문서

#### API 레퍼런스
- **[Board Service API](docs/api/board-service-api.md)**: Board Service API 문서
- **[User Service API](docs/api/user-service-api.md)**: User Service API 문서

### 6.3 개발 가이드

#### 개발자 가이드
- **[Backend 최적화](docs/guides/backend-optimization.md)**: 백엔드 성능 최적화 가이드
- **[Frontend 구현](docs/guides/frontend-implementation.md)**: 프론트엔드 개발 가이드
- **[Frontend 마이그레이션](docs/guides/frontend-migration-checklist.md)**: 마이그레이션 체크리스트
- **[Coding Conventions](docs/guides/coding-conventions.md)**: 코딩 컨벤션

### 6.4 마이그레이션 가이드

#### 마이그레이션 문서
- **[API 마이그레이션](docs/migration/api-migration-guide.md)**: API 변경사항 마이그레이션 가이드
- **[마이그레이션 문서](docs/migration/migration.md)**: 일반 마이그레이션 가이드

---

## 🔍 7. 테스트 개선

### 7.1 Board Service 테스트

#### 통합 테스트 스크립트
```bash
# Board Service 통합 테스트
./scripts/tests/test-board-integration.sh

# Fractional Indexing 테스트
./scripts/tests/test-fractional-indexing.sh
```

### 7.2 User Service 테스트

#### User Service 테스트 스크립트
```bash
# User Service API 테스트
./scripts/tests/test-user-api.sh

# User Service 통합 테스트
./scripts/tests/test-user-service.sh
```

**관련 문서:**
- [테스트 가이드](scripts/tests/README.md)

---

## 🔐 8. 보안 개선

### 8.1 AWS IAM 정책

#### EC2 IAM 정책 정의
- **ECR 접근**: Docker 이미지 Pull 권한
- **Parameter Store**: 설정 값 읽기 권한
- **CloudWatch**: 로그 및 메트릭 전송 권한

**관련 파일:**
```
docs/EC2_IAM_POLICY.json
```

### 8.2 환경 변수 보안

#### Presigned URL 보안
- **임시 접근**: 시간 제한된 S3 접근
- **직접 업로드**: 서버 부하 감소
- **권한 관리**: 세밀한 권한 제어

---

## 🚀 9. 성능 최적화

### 9.1 캐싱 전략

#### Board Service 캐싱
- **Redis 캐싱**: 자주 조회되는 데이터 캐싱
- **User Info 캐싱**: User Service 호출 감소
- **Workspace 캐싱**: 워크스페이스 정보 캐싱

**관련 문서:**
- [캐시 전략](board-service/CACHE_STRATEGY.md)

### 9.2 데이터베이스 최적화

#### 인덱스 최적화
- **복합 인덱스**: 자주 사용되는 쿼리 최적화
- **Soft Delete 인덱스**: `is_deleted` 필터 최적화

### 9.3 API 최적화

#### 페이지네이션
- **Cursor-based**: 대용량 데이터 효율적 처리
- **Limit/Offset**: 기본 페이지네이션 지원

---

## 📊 10. 모니터링 및 로깅

### 10.1 구조화된 로깅

#### Zap Logger (Go)
- **구조화된 로그**: JSON 형식 로그
- **로그 레벨**: Debug, Info, Warn, Error
- **컨텍스트 정보**: Request ID, User ID 포함

#### Logback (Java)
- **패턴 기반 로깅**: 일관된 로그 포맷
- **MDC**: 컨텍스트 정보 추가

### 10.2 메트릭 수집

#### Prometheus 메트릭
- **요청 메트릭**: HTTP 요청 수, 응답 시간
- **비즈니스 메트릭**: 프로젝트/보드 생성 수
- **데이터베이스 메트릭**: 쿼리 실행 시간
- **외부 API 메트릭**: User Service 호출 추적

---

## 🛠️ 11. 개발 환경 개선

### 11.1 로컬 개발 환경

#### MinIO 지원
- **로컬 S3**: MinIO를 로컬 S3로 사용
- **개발 편의성**: AWS 없이 로컬 개발 가능

### 11.2 샘플 데이터

#### 자동 데이터 생성
- **개발 환경**: 초기 데이터 자동 생성
- **테스트 데이터**: 통합 테스트용 데이터

---

## 📝 12. Context Path 배포 전략

### 12.1 User Service Context Path

#### 환경별 Context Path 설정
- **로컬 개발**: Context path 없음
- **AWS 환경**: `/api/users` prefix

**관련 문서:**
- [Context Path 배포](user-service/CONTEXT_PATH_DEPLOYMENT.md)

### 12.2 Board Service Base Path

#### 환경 변수 기반 Base Path
- **로컬 개발**: `SERVER_BASE_PATH=""` (빈 문자열)
- **AWS 환경**: `SERVER_BASE_PATH="/api/boards"`

---

## 🎯 13. 주요 버그 수정

### 13.1 Board Service
- ✅ Attachment ID 관련 버그 수정
- ✅ Comment 저장 로직 개선
- ✅ UUID 생성 위치 변경으로 버그 해결
- ✅ Project Response 메타데이터 수정

### 13.2 User Service
- ✅ User image localhost 경로 문제 수정
- ✅ Workspace sample data seeding 중복 이메일 문제 해결
- ✅ Board Service URL 설정 버그 수정
- ✅ Java 빌드 에러 수정

### 13.3 Frontend
- ✅ API 호출 경로 수정
- ✅ 타입 정의 개선
- ✅ 모달 상태 관리 버그 수정

---

## 📌 14. Breaking Changes

### 14.1 API 변경사항

#### Custom Field → Field Option
- **API 엔드포인트 변경**: `/custom-fields` → `/field-options`
- **DTO 변경**: CustomFieldDTO → FieldOptionDTO

**마이그레이션 가이드:**
- [API 마이그레이션 가이드](docs/migration/api-migration-guide.md)

### 14.2 환경 변수 변경

#### 새로운 환경 변수
```bash
# Board Service
SERVER_BASE_PATH=/api/boards  # AWS 환경
BOARD_SERVICE_URL=http://board-service:8000

# User Service
SAMPLE_DATA_ENABLED=true  # 샘플 데이터 활성화
S3_REGION=ap-northeast-2
S3_BUCKET=wealist-bucket
```

---

## 🔄 15. 의존성 업데이트

### 15.1 Board Service (Go)
```go
// 주요 의존성 추가
github.com/prometheus/client_golang
github.com/aws/aws-sdk-go-v2
github.com/gorilla/websocket
```

### 15.2 User Service (Spring Boot)
```gradle
// 주요 의존성 추가
implementation 'software.amazon.awssdk:s3'
implementation 'org.springframework.boot:spring-boot-starter-validation'
```

### 15.3 Frontend
```json
// 패키지 매니저 변경
npm → pnpm

// 주요 패키지 업데이트
"react": "^18.x",
"typescript": "^5.x"
```

---

## 📅 16. 향후 계획

### 16.1 계획 중인 기능
- [ ] 알림 시스템 구현
- [ ] 검색 기능 강화
- [ ] 모바일 반응형 개선
- [ ] 성능 최적화 지속

### 16.2 기술 부채 해결
- [ ] 레거시 코드 리팩토링
- [ ] 테스트 커버리지 향상
- [ ] 문서화 지속 개선

---

## 🙏 17. 기여자

### 개발팀
- Backend Team: Board Service & User Service 개발
- Frontend Team: React 애플리케이션 개발
- DevOps Team: CI/CD 파이프라인 구축

---

## 📖 18. 추가 참고 자료

### 프로젝트 문서
- [메인 README](README.md)
- [Docker 가이드](docker/README.md)
- [Board Service README](board-service/README.md)
- [User Service README](user-service/README.md)
- [Frontend README](frontend/README.md)

### 아키텍처 문서
- [Board Service 아키텍처](board-service/ARCHITECTURE.md)
- [캐시 전략](board-service/CACHE_STRATEGY.md)
- [개발자 가이드라인](board-service/docs/DEVELOPER_GUIDELINES.md)

### API 문서
- [Board Service API](docs/api/board-service-api.md)
- [User Service API](docs/api/user-service-api.md)
- [Swagger 문서](board-service/docs/SWAGGER.md)

---

**마지막 업데이트**: 2025-01-23
**문서 버전**: 1.0.0
