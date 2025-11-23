# CI/CD Integration Guide

## 개요

Board Service는 GitHub Actions를 사용하여 CI/CD 파이프라인을 구현합니다. 이 문서는 Swagger 문서 검증 및 자동 생성이 CI/CD 파이프라인에 통합된 방법을 설명합니다.

## CI Pipeline (Continuous Integration)

### 워크플로우 파일

`.github/workflows/ci-dev-board-service.yml`

### 트리거 조건

- `deploy-dev` 브랜치에 push
- `board-service/` 디렉토리 또는 CI 워크플로우 파일 변경
- 수동 실행 (workflow_dispatch)

### CI 단계

#### 1. 코드 체크아웃
```yaml
- name: 📥 Checkout Code
  uses: actions/checkout@v4
```

#### 2. Go 환경 설정
```yaml
- name: 🔧 Setup Go Environment
  uses: actions/setup-go@v5
  with:
    go-version: "1.25"
    cache: true
    cache-dependency-path: board-service/go.sum
```

#### 3. Go 의존성 다운로드
```yaml
- name: 📦 Download Go Dependencies
  run: |
    cd board-service
    go mod tidy
    go mod download
```

#### 4. Swagger 도구 설치
```yaml
- name: 📚 Install Swagger Tools
  run: |
    go install github.com/swaggo/swag/cmd/swag@latest
```

#### 5. Swagger 문서 생성
```yaml
- name: 📝 Generate Swagger Documentation
  run: |
    cd board-service
    swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

**중요**: `--parseDependency`와 `--parseInternal` 플래그를 반드시 포함해야 모든 DTO 스키마가 올바르게 생성됩니다.

#### 6. Swagger 문서 검증 ⭐
```yaml
- name: ✅ Validate Swagger Documentation
  run: |
    cd board-service
    
    # 엔드포인트 커버리지 검증
    ./scripts/validate-swagger.sh
    
    # DTO 스키마 검증
    ./scripts/validate-dto-schemas.sh
    
    # Godoc 주석 품질 검증
    ./scripts/validate-godoc.sh
```

이 단계에서 다음을 검증합니다:
- **엔드포인트 커버리지**: router.go의 모든 라우트가 swagger.yaml에 문서화되어 있는지 확인
- **DTO 스키마**: 핸들러에서 사용하는 모든 DTO가 swagger definitions에 정의되어 있는지 확인
- **Godoc 주석 품질**: 모든 핸들러에 필수 godoc 주석이 있는지 확인

**검증 실패 시**: CI 파이프라인이 실패하고 빌드가 중단됩니다.

#### 7. Go 테스트 실행
```yaml
- name: 🧪 Run Go Tests
  continue-on-error: true
  run: |
    cd board-service
    go test -v -race -cover ./...
```

#### 8. Go 빌드 검증
```yaml
- name: 🔨 Verify Go Build
  run: |
    cd board-service
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o board-api ./cmd/api
```

#### 9-12. Docker 이미지 빌드 및 ECR 푸시
- AWS 자격 증명 설정
- 이미지 태그 생성
- Docker 이미지 빌드
- Amazon ECR에 푸시

### CI 성공 조건

CI 파이프라인이 성공하려면:
1. ✅ Go 의존성 다운로드 성공
2. ✅ Swagger 문서 생성 성공
3. ✅ **모든 Swagger 검증 통과** (100% 커버리지)
4. ✅ Go 빌드 성공
5. ✅ Docker 이미지 빌드 및 푸시 성공

## CD Pipeline (Continuous Deployment)

### 워크플로우 파일

`.github/workflows/cd-dev-board-service.yml`

### 트리거 조건

- CI 워크플로우 성공 시 자동 실행
- `deploy-dev` 브랜치에 인프라 파일 변경 시
- 수동 실행 (workflow_dispatch)

### CD 단계

#### Swagger 문서 자동 재생성 ⭐

배포 스크립트에 Swagger 문서 재생성 단계가 포함되어 있습니다:

```bash
# Swagger 문서 재생성 (배포 전)
echo "📚 Regenerating Swagger documentation..."
cd /home/ubuntu/wealist/board-service

# swag 도구 설치 확인
if ! command -v swag &> /dev/null; then
  echo "  ⏳ Installing swag tool..."
  go install github.com/swaggo/swag/cmd/swag@latest
  export PATH=$PATH:$(go env GOPATH)/bin
fi

# Swagger 문서 생성
echo "  ⏳ Generating Swagger files..."
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

if [ -f "docs/swagger.yaml" ] && [ -f "docs/swagger.json" ]; then
  echo "  ✅ Swagger documentation regenerated successfully"
else
  echo "  ⚠️  Swagger generation may have failed, but continuing deployment..."
fi
```

**배포 시 Swagger 재생성 이유**:
- 최신 코드 변경사항 반영
- 배포 환경에서 문서 일관성 보장
- Swagger UI에서 최신 API 문서 제공

#### 나머지 배포 단계

1. Parameter Store에서 환경변수 로드
2. ECR 로그인
3. 최신 이미지 Pull
4. 인프라 서비스 확인 및 시작
5. 데이터베이스 마이그레이션 실행
6. Board Service 재시작
7. Health Check
8. 리소스 정리

## 검증 스크립트 상세

### 1. validate-swagger.sh

**목적**: 엔드포인트 커버리지 검증

**동작**:
1. `router.go`에서 모든 라우트 추출
2. `swagger.yaml`에서 모든 경로 추출
3. 비교 및 커버리지 계산
4. 100% 미만 시 오류 반환

**사용법**:
```bash
cd board-service
./scripts/validate-swagger.sh
```

**출력 예시**:
```
=== Router Routes ===
GET /api/boards/:boardId
POST /api/boards
...

=== Swagger Paths ===
/api/boards/{boardId}
/api/boards
...

Router endpoints: 25
Swagger endpoints: 25
Coverage: 100%
✅ Full coverage achieved
```

### 2. validate-dto-schemas.sh

**목적**: DTO 스키마 완전성 검증

**동작**:
1. 핸들러 파일에서 사용된 모든 DTO 타입 추출
2. `swagger.yaml`의 definitions 섹션에서 정의된 DTO 추출
3. 비교 및 누락된 스키마 리포트
4. 누락 시 오류 반환

**사용법**:
```bash
cd board-service
./scripts/validate-dto-schemas.sh
```

**출력 예시**:
```
=== Handler DTOs ===
dto.BoardResponse
dto.CreateBoardRequest
...

=== Swagger Definitions ===
dto.BoardResponse
dto.CreateBoardRequest
...

✅ All DTOs are defined in Swagger
```

### 3. validate-godoc.sh

**목적**: Godoc 주석 품질 검증

**동작**:
1. 모든 핸들러 파일 스캔
2. 각 핸들러의 godoc 주석 확인
3. 필수 태그 존재 여부 검증 (@Summary, @Router, @Tags)
4. 불완전한 주석 리포트

**사용법**:
```bash
cd board-service
./scripts/validate-godoc.sh
```

**출력 예시**:
```
Checking godoc annotations...

✅ All handlers have @Summary annotations
✅ All handlers have @Router annotations
✅ All handlers have @Tags annotations

✅ All godoc annotations are complete
```

## 로컬 개발 워크플로우

### 코드 변경 후 체크리스트

1. **Swagger 문서 재생성**
   ```bash
   cd board-service
   make swagger
   ```

2. **검증 스크립트 실행**
   ```bash
   ./scripts/validate-swagger.sh
   ./scripts/validate-dto-schemas.sh
   ./scripts/validate-godoc.sh
   ```

3. **Swagger UI 확인**
   ```bash
   make run
   # 브라우저에서 http://localhost:8000/swagger/index.html 접속
   ```

4. **테스트 실행**
   ```bash
   make test
   ```

5. **커밋 및 푸시**
   ```bash
   git add .
   git commit -m "feat: add new endpoint"
   git push origin feature/new-endpoint
   ```

### Pull Request 생성 시

1. PR 생성 시 CI 파이프라인 자동 실행
2. Swagger 검증 포함한 모든 검사 통과 확인
3. 검증 실패 시 수정 후 재푸시
4. 모든 검사 통과 후 리뷰 요청

## CI/CD 파이프라인 모니터링

### GitHub Actions UI

1. GitHub 저장소 → Actions 탭
2. 워크플로우 실행 목록 확인
3. 각 단계별 로그 확인

### Swagger 검증 실패 시 대응

#### 엔드포인트 커버리지 실패

**증상**: `validate-swagger.sh` 실패

**원인**:
- 새 엔드포인트가 swagger에 문서화되지 않음
- @Router 주석 누락 또는 경로 불일치

**해결**:
1. 누락된 엔드포인트 확인
2. 해당 핸들러에 godoc 주석 추가
3. `make swagger` 실행
4. 재검증

#### DTO 스키마 검증 실패

**증상**: `validate-dto-schemas.sh` 실패

**원인**:
- 새 DTO가 swagger definitions에 정의되지 않음
- `--parseDependency` 또는 `--parseInternal` 플래그 누락

**해결**:
1. 누락된 DTO 확인
2. 올바른 플래그로 swagger 재생성
   ```bash
   swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
   ```
3. 재검증

#### Godoc 주석 품질 검증 실패

**증상**: `validate-godoc.sh` 실패

**원인**:
- 핸들러에 필수 godoc 주석 누락
- @Summary, @Router, @Tags 중 하나 이상 누락

**해결**:
1. 불완전한 핸들러 확인
2. 누락된 주석 추가
3. 재검증

## 모범 사례

### 1. 커밋 전 로컬 검증

```bash
# 한 번에 모든 검증 실행
cd board-service
make swagger && \
./scripts/validate-swagger.sh && \
./scripts/validate-dto-schemas.sh && \
./scripts/validate-godoc.sh && \
echo "✅ All validations passed"
```

### 2. Pre-commit Hook 설정 (선택사항)

`.git/hooks/pre-commit` 파일 생성:

```bash
#!/bin/bash

echo "🔍 Running pre-commit checks..."

cd board-service

# Swagger 문서 재생성
echo "📝 Regenerating Swagger documentation..."
make swagger

# 검증 실행
echo "✅ Validating Swagger documentation..."
if ! ./scripts/validate-swagger.sh; then
  echo "❌ Swagger validation failed"
  exit 1
fi

if ! ./scripts/validate-dto-schemas.sh; then
  echo "❌ DTO schema validation failed"
  exit 1
fi

if ! ./scripts/validate-godoc.sh; then
  echo "❌ Godoc annotation validation failed"
  exit 1
fi

echo "✅ All pre-commit checks passed"
```

실행 권한 부여:
```bash
chmod +x .git/hooks/pre-commit
```

### 3. 문서 변경사항 커밋

Swagger 문서 변경사항은 항상 커밋에 포함:

```bash
git add docs/swagger.yaml docs/swagger.json docs/docs.go
git commit -m "docs(swagger): update API documentation"
```

### 4. CI 실패 시 빠른 대응

1. GitHub Actions 로그 확인
2. 실패한 검증 스크립트 식별
3. 로컬에서 동일한 검증 실행
4. 문제 수정 후 재푸시

## 참고 자료

- [GitHub Actions 문서](https://docs.github.com/en/actions)
- [Swagger 문서 가이드](SWAGGER.md)
- [개발자 가이드](DEVELOPER_GUIDELINES.md)

## 문의

CI/CD 통합 관련 문의사항이나 개선 제안은 GitHub Issues를 이용해주세요.
