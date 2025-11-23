# 🔒 보안 강화 배포 설정 가이드

Parameter Store 기반 안전한 배포 방식 설정 가이드

> **💡 Terraform 사용 권장**: Parameter Store를 **Infrastructure as Code**로 관리하려면 [Terraform 가이드](./TERRAFORM_PARAMETER_STORE_SETUP.md)를 참고하세요. (코드 리뷰, 변경 이력 관리, 환경별 관리 가능)
>
> 이 문서는 AWS CLI를 사용한 수동 설정 방법을 설명합니다.

## 🎯 보안 강화 포인트

### 기존 방식 (보안 취약)
```
❌ GitHub Secrets에 모든 비밀번호 저장
❌ GitHub 해킹 시 모든 정보 유출 위험
❌ Secrets 변경 시 GitHub에서 수동 업데이트
❌ 누가 언제 접근했는지 추적 어려움
```

### 개선된 방식 (보안 강화)
```
✅ AWS Parameter Store에 비밀 정보 저장
✅ GitHub는 배포 명령만 실행 (실제 비밀번호 모름)
✅ EC2가 Parameter Store에서 직접 값 읽음
✅ IAM으로 접근 제어
✅ Parameter Store에서 변경 이력 추적
✅ 자동 암호화 (KMS)
```

---

## 📝 설정 단계

### 1단계: Parameter Store에 비밀 정보 저장

로컬 터미널에서 실행:

```bash
# AWS CLI 설정 확인
aws sts get-caller-identity

# 리전 설정
REGION="ap-northeast-2"

# PostgreSQL Superuser Password
aws ssm put-parameter \
  --name "/wealist/dev/postgres/superuser-password" \
  --value "YOUR_POSTGRES_PASSWORD" \
  --type "SecureString" \
  --region ${REGION} \
  --description "PostgreSQL superuser password for wealist dev"

# User Service DB Password
aws ssm put-parameter \
  --name "/wealist/dev/db/user-password" \
  --value "YOUR_USER_DB_PASSWORD" \
  --type "SecureString" \
  --region ${REGION} \
  --description "User service database password"

# Board Service DB Password
aws ssm put-parameter \
  --name "/wealist/dev/db/board-password" \
  --value "YOUR_BOARD_DB_PASSWORD" \
  --type "SecureString" \
  --region ${REGION} \
  --description "Board service database password"

# Redis Password
aws ssm put-parameter \
  --name "/wealist/dev/redis/password" \
  --value "YOUR_REDIS_PASSWORD" \
  --type "SecureString" \
  --region ${REGION} \
  --description "Redis password"

# JWT Secret (64+ characters)
aws ssm put-parameter \
  --name "/wealist/dev/jwt/secret" \
  --value "YOUR_JWT_SECRET_AT_LEAST_64_BYTES_FOR_HS512" \
  --type "SecureString" \
  --region ${REGION} \
  --description "JWT signing secret"

# Google OAuth Client ID (민감하지 않으므로 String)
aws ssm put-parameter \
  --name "/wealist/dev/oauth/google-client-id" \
  --value "YOUR_GOOGLE_CLIENT_ID" \
  --type "String" \
  --region ${REGION} \
  --description "Google OAuth Client ID"

# Google OAuth Client Secret
aws ssm put-parameter \
  --name "/wealist/dev/oauth/google-client-secret" \
  --value "YOUR_GOOGLE_CLIENT_SECRET" \
  --type "SecureString" \
  --region ${REGION} \
  --description "Google OAuth Client Secret"

# Grafana Admin Password
aws ssm put-parameter \
  --name "/wealist/dev/grafana/admin-password" \
  --value "YOUR_GRAFANA_PASSWORD" \
  --type "SecureString" \
  --region ${REGION} \
  --description "Grafana admin password"

echo "✅ All parameters stored successfully!"
```

**확인:**
```bash
# 저장된 파라미터 목록 확인
aws ssm get-parameters-by-path \
  --path "/wealist/dev" \
  --recursive \
  --region ap-northeast-2 \
  --query 'Parameters[*].[Name,Type]' \
  --output table

# 특정 값 확인 (테스트)
aws ssm get-parameter \
  --name "/wealist/dev/postgres/superuser-password" \
  --with-decryption \
  --region ap-northeast-2 \
  --query 'Parameter.Value' \
  --output text
```

---

### 2단계: EC2 IAM Role 설정

#### A. IAM Policy 생성

```bash
# 정책 파일 내용 (docs/EC2_IAM_POLICY.json 참고)
aws iam create-policy \
  --policy-name WealistEC2DeployPolicy \
  --policy-document file://docs/EC2_IAM_POLICY.json \
  --description "Allow EC2 to access ECR, S3, and Parameter Store for wealist deployment"
```

#### B. IAM Role 생성 (없으면)

```bash
# Trust Policy 파일 생성
cat > /tmp/ec2-trust-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# Role 생성
aws iam create-role \
  --role-name WealistEC2Role \
  --assume-role-policy-document file:///tmp/ec2-trust-policy.json

# Policy 연결
aws iam attach-role-policy \
  --role-name WealistEC2Role \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore

aws iam attach-role-policy \
  --role-name WealistEC2Role \
  --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/WealistEC2DeployPolicy

# Instance Profile 생성 및 연결
aws iam create-instance-profile \
  --instance-profile-name WealistEC2InstanceProfile

aws iam add-role-to-instance-profile \
  --instance-profile-name WealistEC2InstanceProfile \
  --role-name WealistEC2Role
```

#### C. EC2에 Role 연결

**AWS Console 방식:**
```
1. AWS Console → EC2 → Instances
2. wealist-dev-ec2 선택
3. Actions → Security → Modify IAM role
4. IAM role: WealistEC2Role 선택
5. Update IAM role
```

**AWS CLI 방식:**
```bash
# EC2 인스턴스 ID 확인
INSTANCE_ID=$(aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=wealist-dev-ec2" \
  --query 'Reservations[0].Instances[0].InstanceId' \
  --output text)

# IAM Instance Profile 연결
aws ec2 associate-iam-instance-profile \
  --instance-id ${INSTANCE_ID} \
  --iam-instance-profile Name=WealistEC2InstanceProfile
```

---

### 3단계: GitHub Secrets 최소화

이제 GitHub Secrets에는 **AWS Credentials와 비민감 정보만** 저장:

```
GitHub Repository → Settings → Secrets and variables → Actions
```

**필요한 Secrets (최소화):**
```
✅ WEALIST_DEV_AWS_ACCESS_KEY_ID      - AWS Access Key
✅ WEALIST_DEV_AWS_SECRET_ACCESS_KEY  - AWS Secret Key
✅ WEALIST_DEV_AWS_ACCOUNT_ID         - AWS Account ID (12자리)
✅ EC2_INSTANCE_ID                    - EC2 Instance ID (i-xxxxxx)
```

**제거해도 되는 Secrets:**
```
❌ USER_DB_PASSWORD           → Parameter Store로 이동
❌ BOARD_DB_PASSWORD          → Parameter Store로 이동
❌ POSTGRES_SUPERUSER_PASSWORD → Parameter Store로 이동
❌ REDIS_PASSWORD             → Parameter Store로 이동
❌ JWT_SECRET                 → Parameter Store로 이동
❌ GOOGLE_CLIENT_SECRET       → Parameter Store로 이동
❌ GRAFANA_ADMIN_PASSWORD     → Parameter Store로 이동
❌ EC2_HOST                   → 불필요 (SSM 사용)
❌ EC2_SSH_PRIVATE_KEY        → 불필요 (SSM 사용)
```

---

### 4단계: 배포 워크플로우 활성화

```bash
# 기존 워크플로우 비활성화
git mv .github/workflows/dev-backend-deploy.yml \
       .github/workflows/_OLD_dev-backend-deploy-ssh.yml

git mv .github/workflows/dev-backend-deploy-ssm.yml \
       .github/workflows/_OLD_dev-backend-deploy-ssm.yml

# 새 보안 워크플로우 활성화
git add .github/workflows/dev-backend-deploy-secure.yml
git add docs/EC2_IAM_POLICY.json
git add docs/SECURE_DEPLOYMENT_SETUP.md

git commit -m "feat: implement secure deployment with Parameter Store"
git push origin deploy-dev
```

---

## 🧪 테스트

### 1. Parameter Store 접근 테스트 (EC2에서)

```bash
# SSM으로 EC2 접속
aws ssm start-session --target i-xxxxxxxxx

# Parameter Store 값 읽기 테스트
aws ssm get-parameter \
  --name "/wealist/dev/postgres/superuser-password" \
  --with-decryption \
  --region ap-northeast-2 \
  --query 'Parameter.Value' \
  --output text

# 성공하면 비밀번호가 출력됨
```

### 2. 배포 테스트

```
GitHub → Actions → Backend EC2 CD - Secure
→ Run workflow
→ Branch: deploy-dev
→ Run workflow
```

**확인 사항:**
- ✅ "Fetching secrets from Parameter Store..." 성공
- ✅ ECR 로그인 성공
- ✅ 이미지 Pull 성공
- ✅ 서비스 시작 성공
- ✅ Health Check 성공

---

## 🔐 보안 Best Practices

### 1. Parameter Store 값 변경

```bash
# 비밀번호 업데이트
aws ssm put-parameter \
  --name "/wealist/dev/db/user-password" \
  --value "NEW_PASSWORD" \
  --type "SecureString" \
  --region ap-northeast-2 \
  --overwrite

# 변경 후 재배포 (EC2가 자동으로 새 값 읽음)
```

### 2. 접근 제어

**Parameter Store 값은 다음만 접근 가능:**
- ✅ wealist-dev-ec2 (IAM Role)
- ✅ 관리자 IAM 사용자/Role
- ❌ GitHub Actions (접근 불가!)
- ❌ 다른 서비스 (접근 불가!)

### 3. 감사 로그

```bash
# Parameter Store 값 변경 이력 확인
aws ssm get-parameter-history \
  --name "/wealist/dev/db/user-password" \
  --region ap-northeast-2
```

### 4. 비밀번호 로테이션

**권장 주기:**
- JWT Secret: 3개월마다
- DB Passwords: 6개월마다
- OAuth Secrets: 변경 시

**로테이션 방법:**
1. Parameter Store에서 값 업데이트
2. 재배포 (자동으로 새 값 사용)
3. 이전 값 삭제 확인

---

## 🆚 보안 비교

| 항목 | SSH 방식 | SSM (Secrets in GitHub) | Parameter Store ✅ |
|------|----------|-------------------------|-------------------|
| SSH 키 필요 | ❌ 필요 | ✅ 불필요 | ✅ 불필요 |
| 포트 오픈 | ❌ 22번 필요 | ✅ 불필요 | ✅ 불필요 |
| GitHub에 비밀 저장 | ❌ 저장 | ❌ 저장 | ✅ 저장 안 함 |
| IAM 기반 접근 제어 | ❌ 없음 | ⚠️ 제한적 | ✅ 완전 제어 |
| 감사 로그 | ❌ 없음 | ⚠️ GitHub만 | ✅ CloudTrail |
| 비밀번호 로테이션 | ❌ 어려움 | ⚠️ 수동 | ✅ 쉬움 |
| KMS 암호화 | ❌ 없음 | ⚠️ GitHub | ✅ AWS KMS |
| 보안 점수 | ⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## 🚨 문제 해결

### Parameter Store 접근 실패

**증상:**
```
An error occurred (AccessDeniedException) when calling GetParameter
```

**해결:**
1. EC2 IAM Role 확인
2. Parameter Store 권한 확인
3. KMS 복호화 권한 확인

### 배포 시 값이 안 읽어짐

**증상:**
```
Parameter not found
```

**해결:**
```bash
# Parameter 이름 확인
aws ssm get-parameters-by-path \
  --path "/wealist/dev" \
  --recursive \
  --region ap-northeast-2
```

---

## 📚 참고 문서

- [AWS Parameter Store 공식 문서](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
- [AWS Secrets Manager vs Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-vs-secrets-manager.html)
- [IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
