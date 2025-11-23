# Board Service 리팩토링 계획 (Temporary)

> ⚠️ **이 문서는 임시 계획 문서입니다.**
> 모든 작업이 완료되면 삭제 예정입니다.
> 작성일: 2025-11-23
> 브랜치: claude/refactor-board-service-016bKhVdwZ8DxmFh5Wvtc69D

---

## 📊 분석 요약

### 코드 현황
- **Service 파일**: board_service.go (661줄)
- **테스트 파일**: 총 3,368줄
  - board_service_test.go: 1,716줄
  - board_service_unit_test.go: 737줄
  - board_service_property_test.go: 650줄
  - board_service_property_partial_test.go: 265줄
- **도메인 모델**: 8개 파일
- **Repository**: 18개 파일

### 의존성
boardServiceImpl이 의존하는 컴포넌트: **9개**
```go
type boardServiceImpl struct {
    boardRepo            repository.BoardRepository
    projectRepo          repository.ProjectRepository
    fieldOptionRepo      repository.FieldOptionRepository
    participantRepo      repository.ParticipantRepository
    attachmentRepo       repository.AttachmentRepository
    s3Client             S3Client
    fieldOptionConverter FieldOptionConverter
    metrics              *metrics.Metrics
    logger               *zap.Logger
}
```

---

## 🔴 발견된 주요 문제점

### 1. BaseModel IsDeleted 필드 불일치 (Critical)
**위치**: `board-service/internal/domain/base.go`

**문제**:
- GetIsDeleted(), SetIsDeleted() 메서드는 존재
- 하지만 실제 IsDeleted 필드는 없음
- DeletedAt (GORM soft delete)만 존재

**영향도**: 컴파일 에러는 없지만 메서드 호출 시 패닉 가능

**해결**:
- 옵션 1: 메서드 삭제 (현재 DeletedAt 사용중)
- 옵션 2: IsDeleted 필드 추가

**우선순위**: 🔥 High

---

### 2. 수동 트랜잭션 롤백 (High)
**위치**: `board-service/internal/service/board_service.go:145-167`

**문제**:
```go
// CreateBoard에서
if err := s.attachmentRepo.ConfirmAttachments(...); err != nil {
    // 이미 커밋된 board를 수동으로 삭제
    if deleteErr := s.boardRepo.Delete(ctx, board.ID); deleteErr != nil {
        s.logger.Error("Failed to rollback board...")
    }
    return nil, response.NewAppError(...)
}
```

**리스크**:
- Board 생성 성공 → Attachment 실패 → Board 삭제 실패 = 고아 레코드
- Participants는 Board에 FK로 연결되어 있어서 함께 남음

**해결**: DB 트랜잭션 사용
```go
return s.db.Transaction(func(tx *gorm.DB) error {
    // 모든 작업을 tx 안에서 수행
    // 실패 시 자동 롤백
})
```

**우선순위**: 🔥 High

---

### 3. N+1 쿼리 문제 (Medium)
**위치**: `board-service/internal/service/board_service.go:264-270`

**문제**:
```go
// GetBoardsByProject - 10개 board = 11번 쿼리
for _, board := range boards {
    attachments, err := s.attachmentRepo.FindByEntityID(ctx, domain.EntityTypeBoard, board.ID)
    board.Attachments = toDomainAttachments(attachments)
}
```

**성능 영향**:
- Boards 1번 조회 + 각 Board마다 Attachments 조회
- 100개 board = 101번 쿼리

**해결**: 배치 조회
```go
// 1. 모든 board ID 수집
boardIDs := extractBoardIDs(boards)

// 2. 한 번에 조회
allAttachments := s.attachmentRepo.FindByEntityIDs(ctx, domain.EntityTypeBoard, boardIDs)

// 3. Map으로 그룹핑 후 할당
attachmentMap := groupByEntityID(allAttachments)
for _, board := range boards {
    board.Attachments = attachmentMap[board.ID]
}
```

**우선순위**: 🟡 Medium

---

### 4. 테스트 코드 비대화 (Medium)
**문제**:
- board_service_test.go가 1,716줄로 너무 큼
- Mock 코드 중복 (100줄 이상)
- 유지보수 어려움

**해결**: 파일 분할
```
board-service/internal/service/
├── board_service_test/
│   ├── setup_test.go              # 공통 setup, mock
│   ├── create_board_test.go       # CreateBoard 테스트
│   ├── get_board_test.go          # GetBoard 테스트
│   ├── update_board_test.go       # UpdateBoard 테스트
│   ├── delete_board_test.go       # DeleteBoard 테스트
│   └── mocks_test.go              # Mock 구현체
```

**우선순위**: 🟡 Medium

---

### 5. Foreign Key 구조 (Low - 검토 필요)
**현재 상태**:
- Board → Project (FK)
- Board → Participants (FK, CASCADE)
- Board → Comments (FK, CASCADE)
- Participant → Board (FK, CASCADE)
- Comment → Board (FK, CASCADE)
- Attachment → 없음 (다형성 관계)

**FK 제거 고려사항**:

#### 제거 가능한 FK (외부 서비스 참조):
- Board.ProjectID → Project (다른 서비스로 분리 가능)
- Board.AuthorID, AssigneeID → User Service 참조
- Project.WorkspaceID, OwnerID → User Service 참조

#### 유지해야 할 FK (같은 Aggregate):
- Board ↔ Participants (같은 비즈니스 컨텍스트)
- Board ↔ Comments (같은 비즈니스 컨텍스트)
- Project ↔ Members (같은 비즈니스 컨텍스트)

#### FK 제거 시 필요한 작업:
1. CASCADE 삭제 로직을 Service 레이어에서 구현
2. 참조 무결성 검증 코드 추가
3. Orphaned records 정리 배치 작업
4. Domain Event로 서비스 간 동기화

**장점**:
- 마이크로서비스 독립성 향상
- DB Sharding 가능
- 서비스 간 느슨한 결합

**단점**:
- 데이터 무결성 보장 어려움
- 애플리케이션 복잡도 증가
- Orphaned records 발생 가능

**권장**: 하이브리드 접근
- 외부 서비스 참조 FK만 제거
- 내부 Aggregate FK는 유지
- Eventually Consistent 모델 도입

**우선순위**: 🟢 Low (추후 검토)

---

### 6. 사용되지 않는 코드 (Low)
**파일**: `board-service/internal/service/board_service_with_uow.go`

**문제**:
- Unit of Work 패턴 예제 코드만 있음
- 실제로 사용되지 않음 (모두 주석 처리)
- 290줄의 미사용 코드

**해결**:
- 옵션 1: 삭제
- 옵션 2: 실제 적용 (복잡한 트랜잭션에만)
- 옵션 3: 예제로 유지 (docs/examples로 이동)

**우선순위**: 🟢 Low

---

## 🎯 작업 계획 (브랜치별 분리)

### ✅ 현재 브랜치 (claude/refactor-board-service-016bKhVdwZ8DxmFh5Wvtc69D)
**작업 시간**: 1시간 이내

1. **BaseModel 메서드 삭제** (5분)
   - GetIsDeleted(), SetIsDeleted() 제거
   - 파일: `board-service/internal/domain/base.go`

2. **이 문서 작성** (완료)

3. **커밋 & 푸시**

---

### 🔜 다음 브랜치 #1: 트랜잭션 개선
**예상 작업 시간**: 2-3일

**작업 내용**:
1. Repository 인터페이스에 WithTx 메서드 추가
2. CreateBoard 트랜잭션 적용
3. UpdateBoard 트랜잭션 적용
4. DeleteBoard 트랜잭션 적용 (CASCADE 로직)
5. 통합 테스트

**파일**:
- `internal/repository/interfaces.go`
- `internal/repository/*_repository.go` (5개 파일)
- `internal/service/board_service.go`
- `internal/service/board_service_test.go`

**성공 기준**:
- 중간 실패 시 완전 롤백 검증
- 테스트 커버리지 유지

---

### 🔜 다음 브랜치 #2: 성능 최적화
**예상 작업 시간**: 2-3일

**작업 내용**:
1. AttachmentRepository.FindByEntityIDs 배치 메서드 추가
2. GetBoardsByProject N+1 쿼리 해결
3. Attachment 변환 로직 통합
4. 성능 측정 (Before/After)

**파일**:
- `internal/repository/attachment_repository.go`
- `internal/service/board_service.go`

**성능 목표**:
- GetBoardsByProject 쿼리 수: N+1 → 2-3개
- 응답 시간 50% 이상 개선

---

### 🔜 다음 브랜치 #3: 테스트 리팩토링
**예상 작업 시간**: 3-4일

**작업 내용**:
1. 테스트 파일 분할
   - create_board_test.go
   - get_board_test.go
   - update_board_test.go
   - delete_board_test.go
2. Mock 코드 공통화 (mocks_test.go)
3. 테스트 헬퍼 함수 정리
4. Table-driven tests 적용

**파일**:
- `internal/service/board_service_test/*.go` (신규)

**목표**:
- 각 테스트 파일 300줄 이하
- Mock 중복 제거
- 테스트 가독성 향상

---

### 🔜 다음 브랜치 #4: FK 제거 검토 (선택적)
**예상 작업 시간**: 1주

**작업 내용**:
1. FK 제거 최종 결정 (하이브리드 방식)
2. 외부 서비스 참조 FK 제거
   - Board.ProjectID
   - Board.AuthorID, AssigneeID
   - Project.WorkspaceID, OwnerID
3. CASCADE 로직 Service 구현
4. 참조 무결성 검증 추가
5. Migration 스크립트
6. Orphaned records 정리 배치 작업

**파일**:
- `internal/domain/*.go` (FK 제거)
- `internal/service/*.go` (CASCADE 구현)
- `migrations/*.sql` (신규)

**리스크**:
- 데이터 무결성 저하
- 복잡도 증가
- 철저한 테스트 필요

**Go/No-Go 결정 기준**:
- 마이크로서비스 완전 분리 일정
- DB Sharding 필요성
- 현재 FK 관련 성능 문제 여부

---

### 🔜 다음 브랜치 #5: 캐싱 추가
**예상 작업 시간**: 2-3일

**작업 내용**:
- `docs/planning/board-service-optimization.md` 참고
- Workspace 멤버십 캐싱
- User 정보 캐싱
- 캐시 무효화 전략

---

## 📈 예상 효과

### 트랜잭션 개선 후:
- ✅ 데이터 일관성 보장
- ✅ 고아 레코드 제거
- ✅ 롤백 자동화

### 성능 최적화 후:
- ✅ 쿼리 수 90% 감소 (N+1 해결)
- ✅ 응답 속도 50% 개선
- ✅ DB 부하 감소

### 테스트 리팩토링 후:
- ✅ 테스트 유지보수성 향상
- ✅ 코드 중복 제거
- ✅ 신규 테스트 추가 용이

### FK 제거 후 (선택적):
- ✅ 마이크로서비스 독립성 향상
- ✅ DB Sharding 가능
- ⚠️ 복잡도 증가
- ⚠️ 데이터 무결성 리스크

---

## 📌 참고 문서

- **성능 최적화 계획**: `docs/planning/board-service-optimization.md`
- **Board Service 코드**: `board-service/internal/service/board_service.go`
- **도메인 모델**: `board-service/internal/domain/*.go`
- **Repository**: `board-service/internal/repository/*.go`

---

## ✅ 체크리스트

### 현재 브랜치
- [ ] BaseModel 메서드 삭제
- [ ] 테스트 확인
- [ ] 커밋 & 푸시
- [ ] PR 생성

### 완료 후
- [ ] 이 문서 삭제
- [ ] 다음 브랜치 이슈 생성

---

## 💭 메모

### FK 제거 관련 추가 고려사항
1. **Soft Delete와의 조합**
   - FK가 없으면 DeletedAt 기반 조회 시 참조 무결성 체크 불가
   - WHERE deleted_at IS NULL 조건 필수

2. **트랜잭션 범위**
   - FK CASCADE가 없으면 트랜잭션이 길어짐
   - Distributed Transaction 고려 필요

3. **동시성 제어**
   - FK 없이 참조 무결성 보장하려면 락 필요
   - SELECT FOR UPDATE 활용

### 테스트 전략
1. **단위 테스트**: Mock 사용, 비즈니스 로직 검증
2. **통합 테스트**: 실제 DB, 트랜잭션 검증
3. **Property-based 테스트**: 경계 조건, 랜덤 데이터

### 성능 모니터링
- Prometheus 메트릭 추가
- 쿼리 수 추적
- 응답 시간 P50, P95, P99
- 캐시 히트율

---

**마지막 업데이트**: 2025-11-23
**다음 리뷰**: 각 브랜치 완료 시
