# EC2 Dev 환경 배포 가이드

weAlist 프로젝트의 백엔드 서비스를 단일 EC2 인스턴스에 배포하는 가이드입니다.

**중요**:
- **EC2**: User Service, Board Service, PostgreSQL, Redis, Monitoring
- **S3**: Frontend (정적 웹사이트로 별도 배포)

## 📋 목차

- [1. EC2 인스턴스 준비](#1-ec2-인스턴스-준비)
- [2. 초기 설정](#2-초기-설정)
- [3. 프로젝트 배포](#3-프로젝트-배포)
- [4. 서비스 확인](#4-서비스-확인)
- [5. 모니터링](#5-모니터링)
- [6. 문제 해결](#6-문제-해결)

## 1. EC2 인스턴스 준비

### 1.1 인스턴스 생성

**권장 사양**:
- **인스턴스 타입**: `t3.small` (2 vCPU, 2GB RAM) - Dev 환경 최적화 ⭐
- **OS**: Amazon Linux 2 또는 Ubuntu 22.04 LTS
- **스토리지**: 20GB 이상 (gp3)

**리소스 사용량 (예상)**:
- User Service: 512MB
- Board Service: 256MB
- PostgreSQL: 384MB
- Redis: 256MB
- Monitoring (Prometheus + Grafana + Exporters): ~500MB
- **총 메모리: ~1.9GB** (t3.small 2GB에 최적화)

**보안 그룹 설정**:

| 타입 | 프로토콜 | 포트 | 소스 | 용도 |
|------|---------|------|------|------|
| SSH | TCP | 22 | My IP | 관리용 |
| Custom TCP | TCP | 8080 | 0.0.0.0/0 | User Service API |
| Custom TCP | TCP | 8000 | 0.0.0.0/0 | Board Service API |
| Custom TCP | TCP | 9090 | My IP | Prometheus (모니터링) |
| Custom TCP | TCP | 3001 | My IP | Grafana (대시보드) |
| Custom TCP | TCP | 5432 | My IP | PostgreSQL (디버깅용, 선택) |
| Custom TCP | TCP | 6379 | My IP | Redis (디버깅용, 선택) |

**참고**: Frontend는 S3에 별도 배포되므로 EC2에서는 백엔드 서비스만 실행합니다.

### 1.2 Elastic IP 할당 (권장)

재시작 시 IP가 변경되지 않도록 Elastic IP를 할당하는 것을 권장합니다.

```bash
# AWS Console에서:
# EC2 > Elastic IP > Allocate Elastic IP address
# 생성된 EIP를 인스턴스에 Associate
```

## 2. 초기 설정

### 2.1 EC2 접속

```bash
# SSH 키로 접속 (로컬에서 실행)
ssh -i "your-key.pem" ec2-user@YOUR_EC2_PUBLIC_IP

# 또는 Ubuntu의 경우
ssh -i "your-key.pem" ubuntu@YOUR_EC2_PUBLIC_IP
```

### 2.2 시스템 업데이트 및 Docker 설치

**Amazon Linux 2**:
```bash
# 시스템 업데이트
sudo yum update -y

# Docker 설치
sudo amazon-linux-extras install docker -y
sudo service docker start
sudo usermod -a -G docker ec2-user

# Docker Compose 설치 (v2)
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.5/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Git 설치 (이미 설치되어 있음)
sudo yum install git -y

# 로그아웃 후 재접속 (docker 그룹 권한 적용)
exit
```

**Ubuntu 22.04**:
```bash
# 시스템 업데이트
sudo apt update && sudo apt upgrade -y

# Docker 설치
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker ubuntu

# Docker Compose 설치 (이미 포함됨)
# Git 설치
sudo apt install git -y

# 로그아웃 후 재접속
exit
```

재접속 후 Docker 확인:
```bash
docker --version
docker compose version
```

### 2.3 프로젝트 클론

```bash
# SSH 키 설정 (Private Repository인 경우)
ssh-keygen -t ed25519 -C "your_email@example.com"
cat ~/.ssh/id_ed25519.pub
# 출력된 공개키를 GitHub > Settings > SSH keys에 추가

# 프로젝트 클론
git clone git@github.com:YOUR_ORG/wealist-project.git
cd wealist-project
```

## 3. 프로젝트 배포

### 3.1 환경 변수 설정

```bash
# 초기 설정 스크립트 실행
./docker/scripts/ec2-dev.sh setup

# 또는 수동으로:
cp docker/env/.env.ec2-dev.example docker/env/.env.ec2-dev
vi docker/env/.env.ec2-dev
```

**필수 수정 항목**:

1. **JWT Secret 생성**:
```bash
openssl rand -base64 64 | tr -d '\n'
# 출력된 값을 JWT_SECRET에 입력
```

2. **비밀번호 변경**:
   - `POSTGRES_SUPERUSER_PASSWORD`
   - `USER_DB_PASSWORD`
   - `BOARD_DB_PASSWORD`
   - `REDIS_PASSWORD`

3. **Google OAuth 설정**:
   - [Google Cloud Console](https://console.cloud.google.com/apis/credentials)에서 OAuth 2.0 클라이언트 ID 생성
   - Authorized redirect URIs에 추가:
     - `https://your-s3-frontend-url.com/oauth/callback` (S3 Frontend URL)
     - `http://YOUR_EC2_PUBLIC_IP:8080/login/oauth2/code/google`
   - `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` 입력

4. **CORS 설정**:
```bash
# S3에서 호스팅되는 프론트엔드 URL을 CORS에 추가
# 예시: S3 Static Website URL이 https://your-bucket.s3-website.ap-northeast-2.amazonaws.com인 경우
CORS_ORIGINS=https://your-bucket.s3-website.ap-northeast-2.amazonaws.com

# 개발 중이라면 임시로 * 사용 가능 (보안 주의!)
CORS_ORIGINS=*
```

**참고**: Frontend는 S3에 별도 배포되며, 빌드 시 다음 환경 변수를 설정해야 합니다:
- `VITE_REACT_APP_JAVA_API_URL=http://YOUR_EC2_PUBLIC_IP:8080`
- `VITE_REACT_APP_GO_API_URL=http://YOUR_EC2_PUBLIC_IP:8000`

### 3.2 서비스 시작

```bash
# 전체 서비스 시작 (백그라운드)
./docker/scripts/ec2-dev.sh up

# 또는 로그를 보면서 시작 (포그라운드)
./docker/scripts/ec2-dev.sh up-fg
```

**초기 시작 시간**: 약 5-10분 (이미지 빌드 포함)

### 3.3 로그 확인

```bash
# 전체 로그 확인
./docker/scripts/ec2-dev.sh logs

# 특정 서비스 로그만 확인
./docker/scripts/ec2-dev.sh logs user-service
./docker/scripts/ec2-dev.sh logs board-service
./docker/scripts/ec2-dev.sh logs postgres
```

## 4. 서비스 확인

### 4.1 헬스체크

```bash
# 스크립트로 확인
./docker/scripts/ec2-dev.sh health

# 또는 수동으로 확인
curl http://localhost:8080/actuator/health  # User Service
curl http://localhost:8000/health           # Board Service
```

### 4.2 웹 브라우저 접속

EC2 퍼블릭 IP를 `YOUR_EC2_PUBLIC_IP`로 가정:

| 서비스 | URL | 설명 |
|--------|-----|------|
| **User Service Swagger** | http://YOUR_EC2_PUBLIC_IP:8080/swagger-ui/index.html | User API 문서 |
| **Board Service Swagger** | http://YOUR_EC2_PUBLIC_IP:8000/swagger/index.html | Board API 문서 |
| **Prometheus** | http://YOUR_EC2_PUBLIC_IP:9090 | 메트릭 수집 |
| **Grafana** | http://YOUR_EC2_PUBLIC_IP:3001 | 대시보드 (admin/admin) |

**참고**: Frontend는 S3에서 별도로 호스팅됩니다.

### 4.3 컨테이너 상태 확인

```bash
# 실행 중인 컨테이너 확인
./docker/scripts/ec2-dev.sh ps

# 또는
docker ps
```

예상 컨테이너 목록:
- wealist-user-service
- wealist-board-service
- wealist-postgres
- wealist-redis
- wealist-prometheus
- wealist-grafana
- wealist-redis-exporter
- wealist-postgres-exporter
- wealist-node-exporter

**총 9개 컨테이너** (Frontend는 S3에 별도 배포)

## 5. 모니터링

### 5.1 Grafana 대시보드 설정

1. Grafana 접속: http://YOUR_EC2_PUBLIC_IP:3001
2. 로그인: admin / admin
3. Data Source 추가:
   - Configuration > Data sources > Add data source
   - Prometheus 선택
   - URL: `http://prometheus:9090`
   - Save & Test

4. 대시보드 추가:
   - Create > Import
   - 추천 대시보드 ID:
     - PostgreSQL: 9628
     - Redis: 11835
     - Node Exporter: 1860

### 5.2 리소스 모니터링

```bash
# 컨테이너 리소스 사용량 확인
docker stats

# 디스크 사용량 확인
df -h

# 메모리 사용량 확인
free -h
```

## 6. 문제 해결

### 6.1 서비스 시작 실패

**증상**: 컨테이너가 계속 재시작됨

**해결**:
```bash
# 로그 확인
./docker/scripts/ec2-dev.sh logs [service-name]

# 일반적인 원인:
# 1. 환경 변수 오류 (.env.ec2-dev 파일 확인)
# 2. JWT_SECRET 불일치 (User Service와 Board Service)
# 3. 메모리 부족 (인스턴스 타입 업그레이드)
```

### 6.2 메모리 부족

**증상**: 서비스가 OOM(Out of Memory)으로 종료됨

**해결**:

t3.small에서 메모리 부족이 발생하면:

```bash
# 옵션 1: 스왑 메모리 추가 (1GB 추가)
sudo dd if=/dev/zero of=/swapfile bs=1M count=1024
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 영구 적용
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# 옵션 2: 모니터링 스택 제거 (Prometheus, Grafana 등)
# docker-compose.ec2-dev.yml에서 monitoring 섹션 주석 처리

# 옵션 3: 인스턴스 타입 업그레이드
# t3.small → t3.medium (월 ~$15 추가)
```

**참고**: 개발 환경에서 트래픽이 적다면 t3.small + 스왑으로도 충분합니다.

### 6.3 환경 변수 변경

```bash
# 환경 변수 수정
vi docker/env/.env.ec2-dev

# 서비스 재시작 (환경 변수 반영)
./docker/scripts/ec2-dev.sh restart
```

### 6.4 포트 충돌

**증상**: "Port already in use" 에러

**해결**:
```bash
# 사용 중인 포트 확인
sudo netstat -tlnp | grep :8080

# 프로세스 종료
sudo kill -9 <PID>
```

### 6.5 디스크 공간 부족

**증상**: "No space left on device"

**해결**:
```bash
# Docker 정리
docker system prune -a --volumes

# 불필요한 이미지 삭제
docker image prune -a

# 로그 파일 정리
docker logs 명령어는 이미 10MB * 3개 파일로 제한됨
```

### 6.6 데이터베이스 초기화

**모든 데이터를 삭제하고 재시작**:
```bash
# 주의: 모든 데이터가 삭제됩니다!
./docker/scripts/ec2-dev.sh clean
./docker/scripts/ec2-dev.sh up
```

### 6.7 서비스 재시작

```bash
# 전체 재시작
./docker/scripts/ec2-dev.sh restart

# 특정 서비스만 재시작
./docker/scripts/ec2-dev.sh restart user-service
```

## 7. 일반적인 작업

### 7.1 코드 업데이트 배포

```bash
# Git pull
cd ~/wealist-project
git pull origin main

# 이미지 재빌드 및 재시작
./docker/scripts/ec2-dev.sh rebuild
```

### 7.2 데이터베이스 백업

```bash
# PostgreSQL 백업
docker exec wealist-postgres pg_dumpall -U postgres > backup_$(date +%Y%m%d).sql

# 복원
cat backup_20250114.sql | docker exec -i wealist-postgres psql -U postgres
```

### 7.3 Redis 데이터 확인

```bash
# Redis CLI 접속
docker exec -it wealist-redis redis-cli -a YOUR_REDIS_PASSWORD

# 키 확인
KEYS *

# 특정 값 조회
GET key_name
```

### 7.4 PostgreSQL 데이터베이스 접속

```bash
# PostgreSQL 접속
docker exec -it wealist-postgres psql -U postgres

# 데이터베이스 목록
\l

# 특정 데이터베이스 접속
\c wealist_user_db

# 테이블 목록
\dt

# 쿼리 실행
SELECT * FROM users LIMIT 10;

# 종료
\q
```

## 8. 보안 권장 사항

### 8.1 SSH 키 기반 인증만 사용

```bash
# /etc/ssh/sshd_config 수정
sudo vi /etc/ssh/sshd_config

# 다음 설정 확인:
PasswordAuthentication no
PubkeyAuthentication yes

# SSH 재시작
sudo systemctl restart sshd
```

### 8.2 방화벽 설정

```bash
# 필요한 포트만 오픈 (AWS 보안 그룹으로 관리하는 것을 권장)
# iptables 대신 AWS Security Group 사용
```

### 8.3 정기적인 업데이트

```bash
# 시스템 업데이트 (주 1회)
sudo yum update -y  # Amazon Linux
sudo apt update && sudo apt upgrade -y  # Ubuntu

# Docker 이미지 업데이트 (월 1회)
cd ~/wealist-project
git pull
./docker/scripts/ec2-dev.sh rebuild
```

## 9. 비용 최적화

### 9.1 인스턴스 중지

개발하지 않을 때는 인스턴스를 중지하여 비용 절감:

```bash
# 로컬에서 AWS CLI로 중지
aws ec2 stop-instances --instance-ids i-1234567890abcdef0

# 재시작
aws ec2 start-instances --instance-ids i-1234567890abcdef0
```

**참고**: Elastic IP는 인스턴스가 중지된 상태에서도 요금이 부과됩니다.

### 9.2 예약 인스턴스

장기간 사용 시 예약 인스턴스 고려 (최대 72% 할인)

## 10. Frontend S3 배포

Frontend는 별도로 S3에 정적 웹사이트로 배포해야 합니다.

### 10.1 Frontend 빌드

```bash
cd frontend

# 환경 변수 설정 (.env.production)
cat > .env.production << EOF
VITE_REACT_APP_JAVA_API_URL=http://YOUR_EC2_PUBLIC_IP:8080
VITE_REACT_APP_GO_API_URL=http://YOUR_EC2_PUBLIC_IP:8000
VITE_GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
EOF

# 빌드
pnpm build
```

### 10.2 S3 배포

```bash
# S3 버킷 생성
aws s3 mb s3://your-frontend-bucket

# 정적 웹사이트 호스팅 활성화
aws s3 website s3://your-frontend-bucket \
  --index-document index.html \
  --error-document index.html

# 빌드 파일 업로드
aws s3 sync dist/ s3://your-frontend-bucket --delete

# 퍼블릭 읽기 권한 설정
aws s3api put-bucket-policy --bucket your-frontend-bucket --policy file://bucket-policy.json
```

### 10.3 CORS 업데이트

S3 배포 후 EC2의 CORS 설정을 업데이트:

```bash
# .env.ec2-dev 수정
CORS_ORIGINS=https://your-frontend-bucket.s3-website.ap-northeast-2.amazonaws.com

# 서비스 재시작
./docker/scripts/ec2-dev.sh restart
```

## 11. 다음 단계

- [ ] CloudFront CDN 추가 (S3 앞단)
- [ ] Let's Encrypt SSL 인증서 적용
- [ ] 자동 백업 스크립트 설정
- [ ] CloudWatch 로그 연동
- [ ] GitHub Actions CI/CD 구성
- [ ] Terraform으로 인프라 자동화

---

**참고 문서**:
- [프로젝트 README](../README.md)
- [Docker 가이드](../docker/README.md)
- [CLAUDE.md](../CLAUDE.md)
