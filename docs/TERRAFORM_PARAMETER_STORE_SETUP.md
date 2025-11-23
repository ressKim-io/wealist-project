# 🔐 Terraform으로 Parameter Store 관리하기

**Infrastructure as Code 방식**으로 안전하게 Parameter Store를 관리하는 가이드입니다.

## 📌 왜 Terraform으로 관리하나?

### AWS CLI 방식의 문제점
```bash
❌ 수동으로 8개 파라미터 입력 (실수 가능)
❌ 누가 언제 생성했는지 추적 어려움
❌ 환경별 일관성 보장 어려움
❌ 변경 이력 관리 안됨
❌ 코드 리뷰 불가능
```

### Terraform 방식의 장점
```bash
✅ Infrastructure as Code (코드로 관리)
✅ Git으로 변경 이력 추적
✅ terraform plan으로 변경 사항 미리 확인
✅ 환경별 변수 분리 관리 (dev/staging/prod)
✅ 팀원과 코드 리뷰 가능
✅ 일관성 있는 배포
✅ 롤백 가능
```

---

## 🚨 보안 주의사항

### ⚠️ 절대 하면 안되는 것

1. **terraform.tfvars를 Git에 커밋하지 마세요**
   ```bash
   # .gitignore에 이미 추가됨
   *.tfvars  # 모든 .tfvars 파일 제외
   !*.tfvars.example  # example 파일만 포함
   ```

2. **실제 비밀번호를 .tf 파일에 직접 쓰지 마세요**
   ```hcl
   # ❌ 절대 이렇게 하지 마세요
   resource "aws_ssm_parameter" "bad_example" {
     value = "my-actual-password-123"  # Git에 노출됨!
   }

   # ✅ 올바른 방법
   resource "aws_ssm_parameter" "good_example" {
     value = var.my_password  # 변수 사용
   }
   ```

3. **State 파일을 로컬에만 두지 마세요 (팀 작업 시)**
   - State 파일에 실제 비밀번호가 평문으로 저장됨
   - S3 백엔드 사용 + 암호화 필수

---

## 📝 설정 단계

### 1단계: Terraform 파일 확인

```bash
cd terraform/dev/parameter-store

# 파일 구조
terraform/dev/parameter-store/
├── parameter-store.tf          # Parameter Store 리소스 정의
└── terraform.tfvars.example    # 변수 예제 파일
```

### 2단계: 변수 파일 생성

```bash
# terraform.tfvars.example을 복사
cp terraform.tfvars.example terraform.tfvars

# 실제 값 입력 (에디터 사용)
vim terraform.tfvars  # 또는 nano, code 등
```

**terraform.tfvars 예시**:
```hcl
aws_region = "ap-northeast-2"

# PostgreSQL (강력한 비밀번호 사용)
postgres_superuser_password = "P0stgr3s!S3cur3P@ssw0rd#2025"

# User Service DB
user_db_password = "Us3rS3rv1c3!P@ssw0rd#2025"

# Board Service DB
board_db_password = "B0@rdS3rv1c3!P@ssw0rd#2025"

# Redis
redis_password = "R3d1s!S3cur3P@ssw0rd#2025"

# JWT Secret (64+ characters)
jwt_secret = "eyJhbGciOiJIUzUxMiJ9.S3cur3JWT.S1gn1ng.S3cr3t.F0r.HS512.2025.V3ry.L0ng"

# Google OAuth
google_client_id = "123456789-abcdefg.apps.googleusercontent.com"
google_client_secret = "GOCSPX-abcdefghijklmnop"

# Grafana
grafana_admin_password = "Gr@f@n@Adm1n!P@ss2025"
```

**비밀번호 생성 도우미**:
```bash
# 안전한 랜덤 비밀번호 생성 (macOS/Linux)
openssl rand -base64 32

# JWT Secret (64+ bytes)
openssl rand -base64 64 | tr -d '\n'
```

### 3단계: Terraform 초기화

```bash
cd terraform/dev/parameter-store

# Terraform 초기화 (providers 다운로드)
terraform init

# 출력 예시:
# Initializing the backend...
# Initializing provider plugins...
# - Finding hashicorp/aws versions matching "~> 5.0"...
# - Installing hashicorp/aws v5.x.x...
# Terraform has been successfully initialized!
```

### 4단계: 변경 사항 미리 확인

```bash
# 어떤 리소스가 생성될지 미리 확인
terraform plan

# 출력 예시:
# Terraform will perform the following actions:
#
#   # aws_ssm_parameter.postgres_superuser_password will be created
#   + resource "aws_ssm_parameter" "postgres_superuser_password" {
#       + arn   = (known after apply)
#       + name  = "/wealist/dev/postgres/superuser-password"
#       + type  = "SecureString"
#       + value = (sensitive value)
#     }
#   ... (총 8개 파라미터)
#
# Plan: 8 to add, 0 to change, 0 to destroy.
```

**plan 출력 확인사항**:
- ✅ `Plan: 8 to add` - 8개 파라미터 생성 예정
- ✅ `value = (sensitive value)` - 민감한 값 숨김 처리
- ✅ Parameter 이름이 올바른지 확인

### 5단계: Parameter Store에 배포

```bash
# Parameter Store에 실제 적용
terraform apply

# 확인 메시지
# Do you want to perform these actions?
#   Terraform will perform the actions described above.
#   Only 'yes' will be accepted to approve.
#
#   Enter a value: yes  # 'yes' 입력

# 출력:
# aws_ssm_parameter.postgres_superuser_password: Creating...
# aws_ssm_parameter.user_db_password: Creating...
# ... (8개 파라미터 생성)
#
# Apply complete! Resources: 8 added, 0 changed, 0 destroyed.
#
# Outputs:
# parameter_names = [
#   "/wealist/dev/postgres/superuser-password",
#   "/wealist/dev/db/user-password",
#   ...
# ]
```

### 6단계: 확인

```bash
# Parameter Store에서 확인
aws ssm get-parameters-by-path \
  --path "/wealist/dev" \
  --recursive \
  --region ap-northeast-2 \
  --query 'Parameters[*].[Name,Type]' \
  --output table

# 출력:
# -------------------------------------------------------
# |                 GetParametersByPath                |
# +---------------------------------------------------+--------------+
# |  /wealist/dev/postgres/superuser-password         | SecureString |
# |  /wealist/dev/db/user-password                    | SecureString |
# |  /wealist/dev/db/board-password                   | SecureString |
# |  /wealist/dev/redis/password                      | SecureString |
# |  /wealist/dev/jwt/secret                          | SecureString |
# |  /wealist/dev/oauth/google-client-id              | String       |
# |  /wealist/dev/oauth/google-client-secret          | SecureString |
# |  /wealist/dev/grafana/admin-password              | SecureString |
# +---------------------------------------------------+--------------+

# 특정 값 확인 (복호화)
aws ssm get-parameter \
  --name "/wealist/dev/postgres/superuser-password" \
  --with-decryption \
  --region ap-northeast-2 \
  --query 'Parameter.Value' \
  --output text
```

---

## 🔄 비밀번호 변경하기

### 방법 1: Terraform으로 변경 (권장)

```bash
# 1. terraform.tfvars 수정
vim terraform.tfvars
# postgres_superuser_password = "NEW_PASSWORD"

# 2. 변경 사항 확인
terraform plan
# Plan: 0 to add, 1 to change, 0 to destroy.

# 3. 적용
terraform apply

# 4. EC2 재배포 (새 비밀번호 사용)
# GitHub Actions에서 "Backend EC2 CD - Secure" 워크플로우 수동 실행
```

### 방법 2: AWS CLI로 변경 (비추천)

```bash
# Terraform 외부에서 변경하면 state와 불일치 발생
aws ssm put-parameter \
  --name "/wealist/dev/postgres/superuser-password" \
  --value "NEW_PASSWORD" \
  --type "SecureString" \
  --region ap-northeast-2 \
  --overwrite

# ⚠️ 이후 terraform apply 시 원래 값으로 되돌려질 수 있음
```

---

## 🗑️ Parameter Store 삭제하기

```bash
# 모든 파라미터 삭제
terraform destroy

# 확인 메시지
# Do you really want to destroy all resources?
#   Enter a value: yes

# Plan: 0 to add, 0 to change, 8 to destroy.
```

---

## 🏢 팀 작업 시 (S3 Backend 설정)

### State 파일을 S3에 저장하기

**문제**: 로컬 terraform.tfstate에 실제 비밀번호가 평문으로 저장됨

**해결**: S3 백엔드 사용 + 암호화

```hcl
# backend.tf 파일 생성
terraform {
  backend "s3" {
    bucket         = "wealist-terraform-state"
    key            = "parameter-store/dev/terraform.tfstate"
    region         = "ap-northeast-2"
    encrypt        = true  # 암호화 필수
    dynamodb_table = "wealist-terraform-locks"  # State 잠금
  }
}
```

**S3 버킷 생성**:
```bash
# State 파일 저장용 버킷
aws s3api create-bucket \
  --bucket wealist-terraform-state \
  --region ap-northeast-2 \
  --create-bucket-configuration LocationConstraint=ap-northeast-2

# 버저닝 활성화 (롤백 가능)
aws s3api put-bucket-versioning \
  --bucket wealist-terraform-state \
  --versioning-configuration Status=Enabled

# 암호화 활성화
aws s3api put-bucket-encryption \
  --bucket wealist-terraform-state \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'
```

**DynamoDB 테이블 생성** (State 잠금):
```bash
aws dynamodb create-table \
  --table-name wealist-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region ap-northeast-2
```

---

## 🌍 환경별 관리 (Dev/Staging/Prod)

### Workspace 사용

```bash
# Dev 환경
terraform workspace new dev
terraform workspace select dev
terraform apply -var-file="dev.tfvars"

# Staging 환경
terraform workspace new staging
terraform workspace select staging
terraform apply -var-file="staging.tfvars"

# Prod 환경
terraform workspace new prod
terraform workspace select prod
terraform apply -var-file="prod.tfvars"

# 현재 workspace 확인
terraform workspace show
```

**파일 구조**:
```
terraform/dev/parameter-store/
├── parameter-store.tf
├── dev.tfvars         # Dev 환경 변수
├── staging.tfvars     # Staging 환경 변수
├── prod.tfvars        # Prod 환경 변수
└── *.tfvars.example
```

---

## 🆚 AWS CLI vs Terraform 비교

| 항목 | AWS CLI | Terraform |
|------|---------|-----------|
| **설정 방법** | 수동 명령어 8번 실행 | `terraform apply` 1번 |
| **변경 추적** | CloudTrail만 | Git + CloudTrail |
| **변경 미리보기** | ❌ 불가능 | ✅ `terraform plan` |
| **일관성** | ⚠️ 실수 가능 | ✅ 코드로 보장 |
| **롤백** | ⚠️ 수동 | ✅ `git revert` + `apply` |
| **코드 리뷰** | ❌ 불가능 | ✅ PR 리뷰 가능 |
| **환경 관리** | ⚠️ 스크립트 필요 | ✅ Workspace |
| **팀 협업** | ⚠️ 어려움 | ✅ S3 Backend |
| **보안** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## 🚨 문제 해결

### 1. "Error: configuring Terraform AWS Provider: no valid credential sources"

**원인**: AWS 자격 증명 설정 안됨

**해결**:
```bash
# AWS CLI 설정
aws configure
# AWS Access Key ID: YOUR_ACCESS_KEY
# AWS Secret Access Key: YOUR_SECRET_KEY
# Default region name: ap-northeast-2
# Default output format: json

# 또는 환경 변수
export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_KEY"
export AWS_DEFAULT_REGION="ap-northeast-2"
```

### 2. "Error: Value for undeclared variable"

**원인**: terraform.tfvars 파일이 없거나 변수명 오타

**해결**:
```bash
# terraform.tfvars 파일 생성 확인
ls -la terraform.tfvars

# 변수명이 parameter-store.tf의 variable 블록과 일치하는지 확인
```

### 3. "Error: creating SSM Parameter: InvalidKeyId"

**원인**: KMS 키 권한 문제

**해결**:
```bash
# SecureString은 기본 KMS 키 사용
# 별도 설정 불필요, IAM 권한 확인
aws sts get-caller-identity
```

### 4. State 파일이 Git에 커밋됨

**해결**:
```bash
# Git에서 제거 (이력에서도 완전 삭제)
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch terraform/dev/parameter-store/terraform.tfstate" \
  --prune-empty --tag-name-filter cat -- --all

# .gitignore 확인
cat .gitignore | grep tfstate
# *.tfstate
# *.tfstate.*

# 강제 푸시 (주의!)
git push origin --force --all
```

---

## 📚 참고 자료

- [Terraform AWS Provider - SSM Parameter](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ssm_parameter)
- [Terraform S3 Backend](https://developer.hashicorp.com/terraform/language/settings/backends/s3)
- [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
- [Terraform Workspaces](https://developer.hashicorp.com/terraform/language/state/workspaces)

---

## ✅ 체크리스트

배포 전 확인사항:

- [ ] `terraform.tfvars` 파일 생성 및 실제 값 입력
- [ ] `.gitignore`에 `*.tfvars` 추가 확인
- [ ] `terraform plan` 실행 및 변경 사항 확인
- [ ] 8개 파라미터 모두 생성되는지 확인
- [ ] AWS 콘솔에서 Parameter Store 확인
- [ ] EC2 IAM Role에 Parameter Store 읽기 권한 확인 (`docs/EC2_IAM_POLICY.json`)
- [ ] GitHub Secrets에 AWS 자격 증명 설정 (`WEALIST_DEV_AWS_*`)
- [ ] GitHub Actions "Backend EC2 CD - Secure" 워크플로우 테스트

---

## 🎯 다음 단계

1. **Parameter Store 설정 완료** (이 문서)
2. **EC2 IAM Role 설정**: `docs/SECURE_DEPLOYMENT_SETUP.md` 참고
3. **배포 테스트**: GitHub Actions에서 수동 실행
4. **모니터링**: CloudTrail에서 Parameter 접근 로그 확인
