# CI/CD 워크플로우 가이드

wealist 프로젝트의 GitHub Actions CI/CD 파이프라인 구조와 사용 방법을 안내합니다.

## 📋 목차
1. [전체 구조](#전체-구조)
2. [워크플로우 설명](#워크플로우-설명)
3. [실행 흐름](#실행-흐름)
4. [설정 방법](#설정-방법)
5. [문제 해결](#문제-해결)

---

## 🏗️ 전체 구조

```
.github/workflows/
├── dev/
│   ├── ci/
│   │   ├── user-service-ci-ecr.yml      ✅ NEW - User Service ECR CI
│   │   ├── board-service-ci-ecr.yml     ✅ NEW - Board Service ECR CI
│   │   ├── frontend-cicd.yml            (기존 Frontend CI/CD)
│   │   ├── user-service-ci.yml          ⚠️ OLD - Docker Hub 사용
│   │   └── board-service-ci.yml         ⚠️ OLD - Docker Hub 사용
│   └── cd/
│       ├── backend-ec2-cd-ecr.yml       ✅ NEW - ECR 이미지 기반 EC2 배포
│       ├── backend-ec2-cd.yml           ⚠️ OLD - Docker Hub 사용
│       ├── infra-dev-apply.yml          (인프라 배포)
│       └── _frontend-deploy.yml         (Frontend S3 배포)
└── README.md                            (이 파일)
```

**✅ 권장:** `-ecr.yml` 파일 사용 (AWS ECR 기반)
**⚠️ 레거시:** 기존 파일은 Docker Hub 사용 (deprecated)

---

## 🔄 워크플로우 설명

### 1. User Service CI - ECR
**파일:** `dev/ci/user-service-ci-ecr.yml`

**트리거:**
- `user-service/` 폴더 변경 시
- 워크플로우 파일 자체 변경 시
- 수동 실행 (`workflow_dispatch`)

**작업 내용:**
1. ☕ Java 21 환경 설정
2. 🔨 Gradle 빌드 & 테스트
3. 📊 JaCoCo 커버리지 리포트 업로드
4. 🔐 AWS Credentials 설정
5. 🔑 Amazon ECR 로그인
6. 🐳 Docker 이미지 빌드
7. 📤 ECR에 이미지 푸시
   - Tag: `{COMMIT_SHA:0:7}` (7자리 commit hash)
   - Tag: `latest`

**결과물:**
- ECR 이미지: `{AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com/wealist-dev-user-service:latest`

---

### 2. Board Service CI - ECR
**파일:** `dev/ci/board-service-ci-ecr.yml`

**트리거:**
- `board-service/` 폴더 변경 시
- 워크플로우 파일 자체 변경 시
- 수동 실행 (`workflow_dispatch`)

**작업 내용:**
1. 🔧 Go 1.25 환경 설정
2. 📦 Go Module 다운로드
3. 🧪 Go 테스트 실행
4. 🔨 Go 빌드 검증
5. 🔐 AWS Credentials 설정
6. 🔑 Amazon ECR 로그인
7. 🐳 Docker 이미지 빌드
8. 📤 ECR에 이미지 푸시
   - Tag: `{COMMIT_SHA:0:7}`
   - Tag: `latest`

**결과물:**
- ECR 이미지: `{AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com/wealist-dev-board-service:latest`

---

### 3. Backend EC2 CD - ECR
**파일:** `dev/cd/backend-ec2-cd-ecr.yml`

**트리거:**
- `User Service CI - ECR` 성공 시
- `Board Service CI - ECR` 성공 시
- 수동 실행 (`workflow_dispatch`)

**작업 내용:**
1. 📥 Docker Compose 파일 체크아웃
2. 🔑 SSH 접속 키 설정
3. 📤 Docker Compose & 스크립트 EC2 전송
4. 🔧 환경변수 파일 생성 (GitHub Secrets 기반)
5. 🚀 EC2 배포:
   - ECR 로그인
   - 최신 이미지 Pull
   - Docker Compose로 서비스 재시작
   - 구 이미지 정리
6. 🏥 헬스체크:
   - User Service: `http://localhost:8080/actuator/health`
   - Board Service: `http://localhost:8000/health`

**배포 대상:**
- EC2 Dev 환경 (`docker-compose.ec2-dev.yml` 사용)
- User Service, Board Service, PostgreSQL, Redis, Monitoring 스택

---

## 🔄 실행 흐름

### 시나리오 1: User Service 코드 변경

```mermaid
graph LR
    A[user-service/ 코드 수정] --> B[Git Push]
    B --> C[User Service CI - ECR]
    C --> D{빌드 성공?}
    D -->|Yes| E[ECR에 이미지 푸시]
    E --> F[Backend EC2 CD 트리거]
    F --> G[EC2에서 이미지 Pull]
    G --> H[Docker Compose 재시작]
    H --> I[헬스체크]
    I --> J{정상?}
    J -->|Yes| K[배포 완료 ✅]
    J -->|No| L[배포 실패 ❌]
    D -->|No| M[CI 실패 ❌]
```

### 시나리오 2: Board Service 코드 변경

```mermaid
graph LR
    A[board-service/ 코드 수정] --> B[Git Push]
    B --> C[Board Service CI - ECR]
    C --> D{빌드 성공?}
    D -->|Yes| E[ECR에 이미지 푸시]
    E --> F[Backend EC2 CD 트리거]
    F --> G[EC2에서 이미지 Pull]
    G --> H[Docker Compose 재시작]
    H --> I[헬스체크]
    I --> J{정상?}
    J -->|Yes| K[배포 완료 ✅]
    J -->|No| L[배포 실패 ❌]
    D -->|No| M[CI 실패 ❌]
```

### 시나리오 3: 수동 배포

```bash
# GitHub UI에서 Actions 탭 이동
1. User Service CI - ECR 선택
2. Run workflow 클릭
3. 브랜치 선택 (feature/cicd-dev-ec2-deploy 또는 main)
4. Run workflow 실행

# 성공 시 자동으로 Backend EC2 CD 실행됨
```

---

## ⚙️ 설정 방법

### 1. GitHub Secrets 설정

**필수 Secrets (총 18개):**

GitHub Repository → Settings → Secrets and variables → Actions → New repository secret

```
AWS 관련 (3개):
- AWS_ACCOUNT_ID
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY

EC2 배포 (2개):
- EC2_HOST
- EC2_SSH_PRIVATE_KEY

Database (9개):
- USER_DB_NAME, USER_DB_USER, USER_DB_PASSWORD
- BOARD_DB_NAME, BOARD_DB_USER, BOARD_DB_PASSWORD
- POSTGRES_SUPERUSER, POSTGRES_SUPERUSER_PASSWORD
- REDIS_PASSWORD

JWT & OAuth (3개):
- JWT_SECRET
- GOOGLE_CLIENT_ID
- GOOGLE_CLIENT_SECRET

Monitoring (1개):
- GRAFANA_ADMIN_PASSWORD
```

**상세 설정 가이드:** [docs/GITHUB_SECRETS_SETUP.md](../../docs/GITHUB_SECRETS_SETUP.md)

### 2. EC2 준비사항

EC2 인스턴스에 다음 도구가 설치되어 있어야 합니다:

```bash
# Docker & Docker Compose
docker --version
docker compose version

# AWS CLI
aws --version

# Git (docker-compose 파일 다운로드용, 선택사항)
git --version
```

**설치 스크립트:**
```bash
# EC2에 SSH 접속
ssh ec2-user@{EC2_HOST}

# Docker 설치
sudo yum update -y
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker ec2-user

# Docker Compose 설치
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# AWS CLI 설치
sudo yum install -y aws-cli

# EC2 IAM Role에 ECR 권한 추가 (AWS Console에서)
# - AmazonEC2ContainerRegistryReadOnly 정책 연결
```

### 3. 브랜치 전략

**개발 환경 (feature/cicd-dev-ec2-deploy):**
- 현재 작업 브랜치
- CI/CD 테스트 및 검증용
- EC2 Dev 환경에 배포

**프로덕션 환경 (main):**
- 검증 완료된 코드만 병합
- 프로덕션 배포 시 별도 워크플로우 사용 권장

---

## 🔍 모니터링 & 로깅

### GitHub Actions 로그 확인

```
GitHub Repository → Actions → 워크플로우 선택 → 실행 결과 클릭
```

**확인 사항:**
- ✅ 모든 Step이 초록색 체크
- ⚠️ 경고 메시지 확인
- ❌ 실패 시 로그 상세 확인

### EC2 배포 상태 확인

```bash
# EC2 SSH 접속
ssh ec2-user@{EC2_HOST}

# 실행 중인 컨테이너 확인
cd ~/wealist-deploy
docker compose --env-file .env ps

# 로그 확인
docker compose --env-file .env logs -f user-service
docker compose --env-file .env logs -f board-service

# 헬스체크
curl http://localhost:8080/actuator/health
curl http://localhost:8000/health
```

### Grafana 모니터링

```
브라우저에서 접속:
http://{EC2_HOST}:3001

로그인:
- Username: admin
- Password: {GRAFANA_ADMIN_PASSWORD}
```

---

## 🐛 문제 해결

### Q1. CI 실패: "ECR login failed"

**원인:** AWS Credentials 잘못됨 또는 IAM 권한 부족

**해결:**
```bash
# AWS Credentials 확인
aws sts get-caller-identity

# IAM 사용자에 ECR 권한 추가
# AmazonEC2ContainerRegistryPowerUser 정책 연결
```

### Q2. CD 실패: "Permission denied (publickey)"

**원인:** SSH Private Key 잘못됨 또는 EC2 Security Group 설정 오류

**해결:**
```bash
# SSH Key 형식 확인 (전체 복사했는지)
# -----BEGIN RSA PRIVATE KEY----- 부터
# -----END RSA PRIVATE KEY----- 까지

# EC2 Security Group에서 SSH(22) 포트 허용 확인
# GitHub Actions IP 대역: https://api.github.com/meta
```

### Q3. 배포 후 서비스 시작 실패

**원인:** 환경변수 누락 또는 Database 초기화 실패

**해결:**
```bash
# EC2에서 .env 파일 확인
ssh ec2-user@{EC2_HOST}
cat ~/wealist-deploy/.env

# 컨테이너 로그 확인
docker compose --env-file ~/wealist-deploy/.env logs postgres
docker compose --env-file ~/wealist-deploy/.env logs user-service

# 컨테이너 재시작
docker compose --env-file ~/wealist-deploy/.env restart user-service
```

### Q4. 이미지가 업데이트되지 않음

**원인:** Docker Compose가 기존 이미지 캐시 사용

**해결:**
```bash
# EC2에서 강제로 이미지 Pull
ssh ec2-user@{EC2_HOST}
cd ~/wealist-deploy

# 최신 이미지 강제 Pull
docker compose --env-file .env pull

# 컨테이너 재생성
docker compose --env-file .env up -d --force-recreate
```

### Q5. CD가 트리거되지 않음

**원인:** CI 워크플로우 이름 불일치 또는 브랜치 제한

**확인:**
```yaml
# backend-ec2-cd-ecr.yml 확인
workflow_run:
  workflows:
    - "User Service CI - ECR"  # 이름이 정확히 일치하는지 확인
    - "Board Service CI - ECR"
  branches:
    - feature/cicd-dev-ec2-deploy  # 현재 브랜치가 포함되어 있는지 확인
    - main
```

---

## 📚 참고 자료

- [GitHub Actions 공식 문서](https://docs.github.com/en/actions)
- [AWS ECR 사용 가이드](https://docs.aws.amazon.com/ecr/)
- [Docker Compose 공식 문서](https://docs.docker.com/compose/)
- [EC2 Dev 배포 가이드](../../docs/EC2-DEV-DEPLOYMENT.md)
- [GitHub Secrets 설정 가이드](../../docs/GITHUB_SECRETS_SETUP.md)

---

## 🔄 업데이트 히스토리

| 날짜 | 변경 사항 | 작성자 |
|------|-----------|--------|
| 2025-XX-XX | ECR 기반 CI/CD 파이프라인 구축 | - |
| 2025-XX-XX | Docker Hub에서 ECR로 마이그레이션 | - |

---

## 📞 문의

CI/CD 관련 문제 발생 시:
1. GitHub Actions 로그 확인
2. EC2 컨테이너 로그 확인
3. 이 문서의 "문제 해결" 섹션 참고
4. 팀 채널에 문의
