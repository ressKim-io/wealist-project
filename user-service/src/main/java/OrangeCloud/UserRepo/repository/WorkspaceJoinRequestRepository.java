package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.WorkspaceJoinRequest;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface WorkspaceJoinRequestRepository extends JpaRepository<WorkspaceJoinRequest, UUID> {

    // 워크스페이스의 모든 가입 신청 조회
    @Query("SELECT wjr FROM WorkspaceJoinRequest wjr WHERE wjr.workspaceId = :workspaceId ORDER BY wjr.requestedAt DESC")
    List<WorkspaceJoinRequest> findByWorkspaceId(@Param("workspaceId") UUID workspaceId);

    // 워크스페이스의 상태별 가입 신청 조회
    @Query("SELECT wjr FROM WorkspaceJoinRequest wjr WHERE wjr.workspaceId = :workspaceId AND wjr.status = :status ORDER BY wjr.requestedAt DESC")
    List<WorkspaceJoinRequest> findByWorkspaceIdAndStatus(@Param("workspaceId") UUID workspaceId, @Param("status") WorkspaceJoinRequest.JoinRequestStatus status);

    // 사용자의 가입 신청 조회
    @Query("SELECT wjr FROM WorkspaceJoinRequest wjr WHERE wjr.userId = :userId ORDER BY wjr.requestedAt DESC")
    List<WorkspaceJoinRequest> findByUserId(@Param("userId") UUID userId);

    // 특정 워크스페이스에 대한 사용자의 PENDING 신청 조회
    @Query("SELECT wjr FROM WorkspaceJoinRequest wjr WHERE wjr.workspaceId = :workspaceId AND wjr.userId = :userId AND wjr.status = 'PENDING'")
    Optional<WorkspaceJoinRequest> findPendingByWorkspaceIdAndUserId(@Param("workspaceId") UUID workspaceId, @Param("userId") UUID userId);

    // 특정 워크스페이스에 대한 사용자의 가입 신청 존재 여부 (PENDING)
    @Query("SELECT COUNT(wjr) > 0 FROM WorkspaceJoinRequest wjr WHERE wjr.workspaceId = :workspaceId AND wjr.userId = :userId AND wjr.status = 'PENDING'")
    boolean existsPendingByWorkspaceIdAndUserId(@Param("workspaceId") UUID workspaceId, @Param("userId") UUID userId);

    // 워크스페이스의 PENDING 신청 수
    @Query("SELECT COUNT(wjr) FROM WorkspaceJoinRequest wjr WHERE wjr.workspaceId = :workspaceId AND wjr.status = 'PENDING'")
    long countPendingByWorkspaceId(@Param("workspaceId") UUID workspaceId);
}