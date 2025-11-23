package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/common/parser"
	"board-service/internal/dto"
)

// ==================== Unit of Work 적용 예제 ====================
//
// 이 파일은 복잡한 트랜잭션에 Unit of Work 패턴을 적용하는 실제 예제를 보여줍니다.
// BoardService에 UnitOfWork를 추가하고 복잡한 비즈니스 로직에 적용합니다.

// ==================== 예제 1: 보드 삭제 시 관련 댓글도 함께 삭제 ====================

// DeleteBoardWithComments는 보드와 모든 관련 댓글을 하나의 트랜잭션으로 삭제합니다
// UnitOfWork를 사용하여 원자성을 보장합니다
//
// Before (UoW 없이):
//   - repo.Delete(boardID) 성공 후 comment 삭제 실패 시 데이터 불일치
//   - 수동으로 트랜잭션 관리 필요
//
// After (UoW 사용):
//   - 모두 성공하거나 모두 실패 (원자성 보장)
//   - 트랜잭션 관리 자동화
//
// 사용법:
//   err := boardService.DeleteBoardWithComments(boardID, userID)
//
func (s *boardService) DeleteBoardWithComments(boardID, userID string) error {
	// Parse UUIDs
	boardUUID, err := parser.ParseBoardID(boardID)
	if err != nil {
		return err
	}

	userUUID, err := parser.ParseUserID(userID)
	if err != nil {
		return err
	}

	// UnitOfWork를 사용한 트랜잭션
	// 만약 s.uow가 없다면 uow.NewUnitOfWork(s.db)로 생성 필요
	// return s.uow.Do(func(repos *uow.Repositories) error {
	// 	// 1. 보드 조회
	// 	board, err := repos.Board.FindByID(boardUUID)
	// 	if err != nil {
	// 		return apperrors.Wrap(err, apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
	// 	}
	//
	// 	// 2. 권한 확인
	// 	canDelete, err := s.authorizer.CanDelete(userUUID, board.ProjectID, board.CreatedBy)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !canDelete {
	// 		return apperrors.New(apperrors.ErrCodeForbidden, "삭제 권한이 없습니다", 403)
	// 	}
	//
	// 	// 3. 보드 삭제 (Domain 메서드 사용)
	// 	board.MarkAsDeleted()
	// 	if err := repos.Board.Update(board); err != nil {
	// 		return err
	// 	}
	//
	// 	// 4. 관련 댓글 모두 조회 및 삭제
	// 	comments, err := repos.Comment.FindByBoard(boardUUID)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	for _, comment := range comments {
	// 		if err := repos.Comment.Delete(comment.ID); err != nil {
	// 			return err
	// 		}
	// 	}
	//
	// 	// 모든 작업 성공 시 자동 커밋, 에러 발생 시 자동 롤백
	// 	return nil
	// })

	// 현재는 UoW가 Service에 주입되지 않았으므로 주석 처리
	// 실제 사용 시에는 위 주석을 해제하고 아래 코드를 삭제
	_, _ = boardUUID, userUUID
	return apperrors.New(apperrors.ErrCodeNotImplemented, "UoW 미구현", 501)
}

// ==================== 예제 2: 보드 이동 (프로젝트 간 이동) ====================

// MoveBoardToProject는 보드를 다른 프로젝트로 이동합니다
// 이동 시 다음 작업들이 하나의 트랜잭션으로 처리됩니다:
//   1. 보드의 ProjectID 변경
//   2. 이전 프로젝트의 커스텀 필드 값 삭제
//   3. 담당자가 새 프로젝트 멤버가 아니면 할당 해제
//   4. 보드 순서 초기화
//
// UnitOfWork의 이점:
//   - 복잡한 다단계 작업의 원자성 보장
//   - 실패 시 모든 변경사항 자동 롤백
//   - 코드 가독성 향상 (비즈니스 로직에 집중)
//
func (s *boardService) MoveBoardToProject(boardID, targetProjectID, userID string) (*dto.BoardResponse, error) {
	// Parse UUIDs
	boardUUID, err := parser.ParseBoardID(boardID)
	if err != nil {
		return nil, err
	}

	targetProjectUUID, err := parser.ParseProjectID(targetProjectID)
	if err != nil {
		return nil, err
	}

	userUUID, err := parser.ParseUserID(userID)
	if err != nil {
		return nil, err
	}

	// return s.uow.Do(func(repos *uow.Repositories) error {
	// 	// 1. 보드 조회
	// 	board, err := repos.Board.FindByID(boardUUID)
	// 	if err != nil {
	// 		return apperrors.Wrap(err, apperrors.ErrCodeNotFound, "보드를 찾을 수 없습니다", 404)
	// 	}
	//
	// 	// 2. 현재 프로젝트에서 이동 권한 확인
	// 	canEdit, err := s.authorizer.CanEdit(userUUID, board.ProjectID, board.CreatedBy)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !canEdit {
	// 		return apperrors.New(apperrors.ErrCodeForbidden, "이동 권한이 없습니다", 403)
	// 	}
	//
	// 	// 3. 대상 프로젝트 존재 확인 및 멤버십 확인
	// 	targetProject, err := repos.Project.FindByID(targetProjectUUID)
	// 	if err != nil {
	// 		return apperrors.Wrap(err, apperrors.ErrCodeNotFound, "대상 프로젝트를 찾을 수 없습니다", 404)
	// 	}
	//
	// 	_, err = repos.Project.FindMemberByUserAndProject(userUUID, targetProjectUUID)
	// 	if err != nil {
	// 		return apperrors.New(apperrors.ErrCodeForbidden, "대상 프로젝트의 멤버가 아닙니다", 403)
	// 	}
	//
	// 	// 4. 보드 이동 (프로젝트 변경)
	// 	board.ProjectID = targetProject.ID
	//
	// 	// 5. 담당자가 새 프로젝트 멤버인지 확인
	// 	if board.AssigneeID != nil {
	// 		_, err := repos.Project.FindMemberByUserAndProject(*board.AssigneeID, targetProjectUUID)
	// 		if err != nil {
	// 			// 새 프로젝트 멤버가 아니면 할당 해제 (Domain 메서드)
	// 			board.Unassign()
	// 		}
	// 	}
	//
	// 	// 6. 이전 프로젝트의 커스텀 필드 값 모두 삭제
	// 	fieldValues, err := repos.Field.FindFieldValuesByBoard(boardUUID)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	for _, fv := range fieldValues {
	// 		if err := repos.Field.DeleteFieldValue(boardUUID, fv.FieldID); err != nil {
	// 			return err
	// 		}
	// 	}
	//
	// 	// 7. 보드 저장
	// 	if err := repos.Board.Update(board); err != nil {
	// 		return err
	// 	}
	//
	// 	return nil
	// })

	// 현재는 UoW가 Service에 주입되지 않았으므로 주석 처리
	_, _, _ = boardUUID, targetProjectUUID, userUUID
	return nil, apperrors.New(apperrors.ErrCodeNotImplemented, "UoW 미구현", 501)
}

// ==================== 예제 3: 프로젝트 삭제 시 모든 관련 데이터 삭제 ====================

// DeleteProjectWithAllData는 프로젝트와 모든 관련 데이터를 삭제합니다
// 다음 작업들이 하나의 트랜잭션으로 처리됩니다:
//   1. 프로젝트의 모든 보드 조회 및 삭제
//   2. 각 보드의 모든 댓글 삭제
//   3. 프로젝트의 모든 커스텀 필드 삭제
//   4. 프로젝트의 모든 멤버 관계 삭제
//   5. 프로젝트 자체 삭제
//
// 이런 복잡한 cascade 삭제는 UnitOfWork 없이 구현하기 매우 어렵습니다.
// UoW를 사용하면 중간에 실패해도 모든 변경사항이 자동으로 롤백됩니다.
//
func DeleteProjectWithAllData(projectID, userID string) error {
	// 구현은 위 패턴과 동일
	// return s.uow.Do(func(repos *uow.Repositories) error {
	//     // 1. 프로젝트 조회 및 권한 확인
	//     // 2. 모든 보드 조회 및 삭제
	//     // 3. 각 보드의 댓글 삭제
	//     // 4. 커스텀 필드 삭제
	//     // 5. 멤버 관계 삭제
	//     // 6. 프로젝트 삭제
	//     return nil
	// })
	_, _ = projectID, userID
	return apperrors.New(apperrors.ErrCodeNotImplemented, "예제 함수", 501)
}

// ==================== UnitOfWork 적용 시 주의사항 ====================
//
// 1. 트랜잭션 길이
//    - 가능한 짧게 유지 (long-running transaction 피하기)
//    - 외부 API 호출은 트랜잭션 밖에서 수행
//
// 2. 데드락 방지
//    - 테이블 접근 순서 일관성 유지
//    - 필요한 경우 SELECT FOR UPDATE 사용
//
// 3. 에러 처리
//    - Domain 레벨 에러와 Infrastructure 레벨 에러 구분
//    - 비즈니스 규칙 위반 시 명확한 에러 메시지
//
// 4. 성능
//    - 대량 데이터 처리 시 배치 작업 고려
//    - 불필요한 조회 최소화 (N+1 문제 주의)
//
// 5. 테스트
//    - UnitOfWork를 Mock으로 교체하여 단위 테스트 가능
//    - 트랜잭션 롤백 테스트 필수
//

// ==================== 3단계 설명: UnitOfWork를 언제 사용해야 하나? ====================
//
// ## 사용해야 하는 경우:
//
// 1. **여러 Aggregate를 수정하는 경우**
//    - 예: 보드 삭제 시 댓글도 함께 삭제
//    - 예: 프로젝트 이동 시 관련 데이터 모두 업데이트
//
// 2. **복잡한 비즈니스 규칙이 여러 엔티티에 걸쳐 있는 경우**
//    - 예: 담당자 변경 시 알림 생성 + 히스토리 기록
//    - 예: 상태 변경 시 여러 엔티티 업데이트
//
// 3. **일관성이 중요한 작업**
//    - 예: 결제 처리 (주문 생성 + 재고 감소)
//    - 예: 권한 변경 (역할 변경 + 권한 재계산)
//
// ## 사용하지 말아야 하는 경우:
//
// 1. **단순 CRUD 작업**
//    - 예: 보드 제목만 수정
//    - 예: 댓글 하나만 삭제
//    - 이런 경우는 Repository 직접 사용
//
// 2. **읽기 전용 작업**
//    - 예: 보드 목록 조회
//    - 예: 댓글 조회
//    - 트랜잭션 불필요
//
// 3. **외부 API 호출이 포함된 경우**
//    - 외부 API는 트랜잭션 밖에서 호출
//    - Saga 패턴 또는 보상 트랜잭션 고려
//
// ## 실제 적용 방법:
//
// 1. **Service에 UnitOfWork 주입**
//    ```go
//    type boardService struct {
//        uow uow.UnitOfWork
//        // ... 기타 의존성
//    }
//    ```
//
// 2. **복잡한 비즈니스 로직에만 적용**
//    ```go
//    func (s *boardService) ComplexOperation() error {
//        return s.uow.Do(func(repos *uow.Repositories) error {
//            // 여기서 여러 repository 사용
//            return nil
//        })
//    }
//    ```
//
// 3. **Domain 메서드와 함께 사용**
//    - UoW로 트랜잭션 관리
//    - Domain 메서드로 비즈니스 로직
//    - Service는 조율만 담당
//
