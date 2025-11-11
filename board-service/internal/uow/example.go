package uow

// ==================== Unit of Work 사용 예제 ====================
//
// UnitOfWork 패턴을 사용하면 여러 repository 작업을 하나의 트랜잭션으로 묶을 수 있습니다.
//
// Example 1: 간단한 트랜잭션
//
//   func (s *boardService) DeleteBoardAndComments(boardID uuid.UUID) error {
//       return s.uow.Do(func(repos *uow.Repositories) error {
//           // 1. 보드 삭제
//           board, err := repos.Board.FindByID(boardID)
//           if err != nil {
//               return err
//           }
//           board.MarkAsDeleted()
//           if err := repos.Board.Update(board); err != nil {
//               return err
//           }
//
//           // 2. 관련 댓글 모두 삭제
//           comments, err := repos.Comment.FindByBoard(boardID)
//           if err != nil {
//               return err
//           }
//           for _, comment := range comments {
//               if err := repos.Comment.Delete(comment.ID); err != nil {
//                   return err
//               }
//           }
//
//           return nil // 성공 시 커밋, 에러 시 롤백
//       })
//   }
//
// Example 2: Rich Domain Model과 함께 사용
//
//   func (s *boardService) UpdateBoardWithValidation(boardID uuid.UUID, req *UpdateBoardRequest) error {
//       return s.uow.Do(func(repos *uow.Repositories) error {
//           // 1. 보드 조회
//           board, err := repos.Board.FindByID(boardID)
//           if err != nil {
//               return err
//           }
//
//           // 2. Domain 메서드로 비즈니스 로직 실행 (검증 포함)
//           if err := board.UpdateTitle(req.Title); err != nil {
//               return err
//           }
//           board.UpdateDescription(req.Description)
//
//           if req.AssigneeID != nil {
//               board.Assign(*req.AssigneeID)
//           } else {
//               board.Unassign()
//           }
//
//           if req.DueDate != nil {
//               board.SetDueDate(*req.DueDate)
//           } else {
//               board.ClearDueDate()
//           }
//
//           // 3. 변경사항 저장
//           return repos.Board.Update(board)
//       })
//   }
//
// Example 3: 복잡한 비즈니스 로직
//
//   func (s *boardService) MoveBoard(boardID uuid.UUID, targetProjectID uuid.UUID) error {
//       return s.uow.Do(func(repos *uow.Repositories) error {
//           // 1. 보드 조회
//           board, err := repos.Board.FindByID(boardID)
//           if err != nil {
//               return err
//           }
//
//           // 2. 대상 프로젝트 확인
//           targetProject, err := repos.Project.FindByID(targetProjectID)
//           if err != nil {
//               return err
//           }
//
//           // 3. 보드 이동 (프로젝트 변경)
//           board.ProjectID = targetProject.ID
//
//           // 4. 이전 프로젝트의 커스텀 필드 값 삭제
//           if err := repos.Field.DeleteFieldValue(board.ID, uuid.Nil); err != nil {
//               return err
//           }
//
//           // 5. 보드 저장
//           return repos.Board.Update(board)
//       })
//   }
//
// ==================== 장점 ====================
//
// 1. 트랜잭션 관리 중앙화: 비즈니스 로직에서 트랜잭션 시작/커밋/롤백을 명시적으로 관리할 필요 없음
// 2. 일관성 보장: 여러 repository 작업이 모두 성공하거나 모두 실패함
// 3. 테스트 용이: UnitOfWork를 mock으로 교체하여 테스트 가능
// 4. 코드 가독성: 비즈니스 로직이 트랜잭션 경계 안에서 명확하게 표현됨
//
// ==================== 주의사항 ====================
//
// 1. Do 함수 내에서는 panic을 사용하지 말 것 (트랜잭션이 롤백되지 않을 수 있음)
// 2. Do 함수는 가능한 짧게 유지 (긴 트랜잭션은 DB 락을 오래 잡음)
// 3. Do 함수 내에서 외부 API 호출은 피할 것 (실패 시 롤백이 어려움)
// 4. 중첩된 UnitOfWork.Do는 지원되지 않음 (GORM의 제약)
//
