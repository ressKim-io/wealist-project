## 그라파나 사용법
localhost:9092 admin/admin 으로 접속   

1. Home 에서 Add your first data source 클릭
2. 프로메테우스 선택
3. URL에 host.docker.internal:9090 입력 (Prometheus)

## 📊 Board Service 대시보드 (신규)
**역할별 맞춤 대시보드 4종 제공**

### 1. ⭐ DevOps 운영 대시보드 (운영팀/SRE용) - 필수!
- 파일: `dashboards/board-service-devops.json`
- **24/7 운영 모니터링용 종합 대시보드**
- 서비스 가용성, SLA, 에러 추이, 리소스 사용률
- 장애 대응 1차 진단 도구
- 알림 설정 권장

### 2. 비즈니스 대시보드 (PM/운영팀용)
- 파일: `dashboards/board-service-business.json`
- 프로젝트/보드 생성 현황, 에러율, 응답 시간
- 비즈니스 성장 추이 분석

### 3. 개발자 대시보드 (백엔드 개발자용)
- 파일: `dashboards/board-service-developer.json`
- DB 쿼리 성능, API 디버깅, External API 모니터링
- 코드 최적화 및 성능 튜닝

### 4. 성능 테스트 대시보드 (QA/성능팀용)
- 파일: `dashboards/board-service-performance.json`
- 부하 테스트 실시간 모니터링
- 병목 지점 분석, SLA 준수 확인

**자세한 사용법**: `dashboards/METRICS_GUIDE.md` 참고

---

## 기존 서비스 대시보드

## user-service 추가
1. 대시보드 설정 왼쪽 대시보드에서 New 버튼에서 import 클릭 후 Upload Json file에 18812_rev4.json 넣기
2. Prometheus에서  3에서 설정한 프로메테우스 클릭후 완료
3. test-api 실행시키면 cpu 증가하는것 확인 가능  

## redis 추가 
1. 대쉬보드에서 import 11692.json 파일 추가 - 데이터 소스 프로메테우스 선택

## postgresql 추가
1. 대쉬보드에서 import 9628.json 파일 추가 - 데이터 소스 프로메테우스 선택

## go 추가 
1. 대쉬보드에서 import -> 6671.json 파일 추가 - 데이터 소스 프로메테우스 선택