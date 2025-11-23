# ALB 설정 검증 가이드

이 문서는 ALB path-based routing 설정이 올바르게 구성되었는지 검증하는 절차를 설명합니다.

## 1. AWS Console에서 Listener Rules 확인

### 1.1 ALB 찾기
1. AWS Console에 로그인
2. EC2 서비스로 이동
3. 왼쪽 메뉴에서 "Load Balancers" 선택
4. 해당 ALB 선택 (예: `wealist-alb`)

### 1.2 Listener 확인
1. ALB 상세 페이지에서 "Listeners" 탭 선택
2. HTTPS:443 리스너 선택
3. "View/edit rules" 클릭

### 1.3 Rules 검증 체크리스트

#### Rule 1: User Service
- [ ] **Path pattern**: `/api/users/*`
- [ ] **Priority**: 1 (가장 높은 우선순위)
- [ ] **Action**: Forward to `user-service-tg`
- [ ] **Status**: Active

#### Rule 2: Board Service
- [ ] **Path pattern**: `/api/boards/*`
- [ ] **Priority**: 2
- [ ] **Action**: Forward to `board-service-tg`
- [ ] **Status**: Active

#### Default Rule
- [ ] **Condition**: Default (no conditions)
- [ ] **Action**: 설정된 기본 동작 확인

### 1.4 스크린샷 예시
```
Priority | Conditions              | Actions
---------|------------------------|------------------
1        | Path: /api/users/*     | Forward to user-service-tg
2        | Path: /api/boards/*    | Forward to board-service-tg
Default  | -                      | [기본 동작]
```

## 2. Target Group Health 상태 확인

### 2.1 User Service Target Group

1. EC2 콘솔에서 "Target Groups" 선택
2. `user-service-tg` 선택
3. "Targets" 탭에서 상태 확인

#### 검증 항목
- [ ] **Health check path**: `/api/users/actuator/health`
- [ ] **Health check interval**: 30초
- [ ] **Healthy threshold**: 2
- [ ] **Unhealthy threshold**: 3
- [ ] **Timeout**: 5초
- [ ] **Success codes**: 200

#### Target 상태
- [ ] 모든 등록된 타겟이 `healthy` 상태
- [ ] Health check 실패가 없음

### 2.2 Board Service Target Group

1. EC2 콘솔에서 "Target Groups" 선택
2. `board-service-tg` 선택
3. "Targets" 탭에서 상태 확인

#### 검증 항목
- [ ] **Health check path**: `/api/boards/health`
- [ ] **Health check interval**: 30초
- [ ] **Healthy threshold**: 2
- [ ] **Unhealthy threshold**: 3
- [ ] **Timeout**: 5초
- [ ] **Success codes**: 200

#### Target 상태
- [ ] 모든 등록된 타겟이 `healthy` 상태
- [ ] Health check 실패가 없음

### 2.3 Health Check 문제 해결

만약 타겟이 `unhealthy` 상태라면:

1. **Health check 로그 확인**
   ```bash
   # EC2 인스턴스에 접속하여 서비스 로그 확인
   docker logs user-service
   docker logs board-service
   ```

2. **Health check 경로 직접 테스트**
   ```bash
   # User Service
   curl http://localhost:8080/api/users/actuator/health
   
   # Board Service
   curl http://localhost:8000/api/boards/health
   ```

3. **일반적인 문제**
   - Context path/Base path 설정 오류
   - 서비스가 시작되지 않음
   - 포트 바인딩 문제
   - 환경 변수 설정 누락

## 3. ALB Access Log 활성화 (선택사항)

ALB access log를 활성화하면 모든 요청을 추적하고 디버깅할 수 있습니다.

### 3.1 S3 버킷 준비

1. S3 콘솔에서 새 버킷 생성 (예: `wealist-alb-logs`)
2. 버킷 정책 설정:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::elb-account-id:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::wealist-alb-logs/AWSLogs/your-account-id/*"
    }
  ]
}
```

**참고**: `elb-account-id`는 리전별로 다릅니다. [AWS 문서](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/enable-access-logging.html) 참조

### 3.2 ALB에서 Access Log 활성화

1. ALB 선택
2. "Attributes" 탭 선택
3. "Edit attributes" 클릭
4. "Access logs" 섹션에서:
   - [ ] "Enable" 체크
   - [ ] S3 버킷 선택: `wealist-alb-logs`
   - [ ] Prefix 입력 (선택사항): `alb-logs`
5. "Save changes" 클릭

### 3.3 Access Log 확인

로그는 다음 형식으로 S3에 저장됩니다:
```
s3://wealist-alb-logs/alb-logs/AWSLogs/account-id/elasticloadbalancing/region/yyyy/mm/dd/
```

로그 파일 예시:
```
https 2024-01-15T10:30:00.123456Z app/wealist-alb/1234567890abcdef 
1.2.3.4:12345 10.0.1.100:8080 0.001 0.002 0.000 200 200 
123 456 "GET https://api.wealist.co.kr:443/api/users/api/workspaces/all HTTP/1.1" 
"Mozilla/5.0..." ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 
arn:aws:elasticloadbalancing:region:account-id:targetgroup/user-service-tg/1234567890abcdef 
"Root=1-..." "api.wealist.co.kr" "arn:aws:acm:..." 0 2024-01-15T10:30:00.123456Z 
"forward" "-" "-" "10.0.1.100:8080" "200" "-" "-"
```

## 4. 통합 검증 테스트

모든 설정이 완료되면 실제 API 호출로 검증합니다.

### 4.1 User Service API 테스트

```bash
# Health check
curl -v https://api.wealist.co.kr/api/users/actuator/health

# 실제 API 호출 (인증 필요)
curl -v https://api.wealist.co.kr/api/users/api/workspaces/all \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**예상 결과**:
- HTTP 200 OK
- 올바른 JSON 응답

### 4.2 Board Service API 테스트

```bash
# Health check
curl -v https://api.wealist.co.kr/api/boards/health

# 실제 API 호출 (인증 필요)
curl -v https://api.wealist.co.kr/api/boards/api/projects \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**예상 결과**:
- HTTP 200 OK
- 올바른 JSON 응답

### 4.3 에러 케이스 테스트

```bash
# 존재하지 않는 경로 (404 예상)
curl -v https://api.wealist.co.kr/api/unknown/path

# 잘못된 prefix (ALB 기본 동작 확인)
curl -v https://api.wealist.co.kr/invalid/path
```

## 5. 검증 완료 체크리스트

### ALB Listener Rules
- [ ] User Service rule이 priority 1로 설정됨
- [ ] Board Service rule이 priority 2로 설정됨
- [ ] Path pattern이 올바르게 설정됨
- [ ] 모든 rule이 active 상태

### Target Group Health
- [ ] User Service TG의 모든 타겟이 healthy
- [ ] Board Service TG의 모든 타겟이 healthy
- [ ] Health check path가 올바르게 설정됨
- [ ] Health check 설정이 적절함 (interval, timeout, threshold)

### API 동작 확인
- [ ] User Service API가 정상 응답
- [ ] Board Service API가 정상 응답
- [ ] Health check가 정상 동작
- [ ] 에러 케이스가 적절히 처리됨

### Access Log (선택사항)
- [ ] S3 버킷이 생성됨
- [ ] ALB access log가 활성화됨
- [ ] 로그가 S3에 저장되는 것을 확인

## 6. 문제 해결

### 문제: Target이 unhealthy 상태

**원인**:
- Health check 경로 오류
- 서비스가 시작되지 않음
- Context path/Base path 설정 오류

**해결**:
1. 서비스 로그 확인
2. Health check 경로 직접 테스트
3. 환경 변수 확인
4. 서비스 재시작

### 문제: 404 Not Found

**원인**:
- Listener rule이 설정되지 않음
- Path pattern이 잘못됨
- Target group이 연결되지 않음

**해결**:
1. Listener rules 확인
2. Path pattern 수정
3. Target group 연결 확인

### 문제: 503 Service Unavailable

**원인**:
- 모든 타겟이 unhealthy
- Target group에 타겟이 없음

**해결**:
1. Target group health 확인
2. 타겟 등록 확인
3. 서비스 상태 확인

## 7. 모니터링 권장사항

### CloudWatch 메트릭
- `TargetResponseTime`: 타겟 응답 시간
- `HealthyHostCount`: Healthy 타겟 수
- `UnHealthyHostCount`: Unhealthy 타겟 수
- `RequestCount`: 요청 수
- `HTTPCode_Target_2XX_Count`: 성공 응답 수
- `HTTPCode_Target_4XX_Count`: 클라이언트 에러 수
- `HTTPCode_Target_5XX_Count`: 서버 에러 수

### 알람 설정 권장
- Unhealthy host count > 0
- Target response time > 1초
- 5XX error rate > 5%

## 참고 자료

- [AWS ALB 문서](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/)
- [Path-based Routing](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-listeners.html#path-conditions)
- [Health Checks](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
- [Access Logs](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/enable-access-logging.html)
