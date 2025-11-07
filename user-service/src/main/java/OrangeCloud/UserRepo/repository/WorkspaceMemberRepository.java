package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.WorkspaceMember;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface WorkspaceMemberRepository extends JpaRepository<WorkspaceMember, UUID> {

    // 💡 [수정] WorkspaceService에서 사용하기 위해 findAllByWorkspaceId 추가 (활성/비활성 모두 조회)
    // 이 메서드가 WorkspaceService의 getWorkspaceMembers에서 필요합니다.
    List<WorkspaceMember> findAllByWorkspaceId(UUID workspaceId); 

    // 워크스페이스의 모든 활성 멤버 조회
    @Query("SELECT wm FROM WorkspaceMember wm WHERE wm.workspaceId = :workspaceId AND wm.isActive = true")
    List<WorkspaceMember> findActiveByWorkspaceId(@Param("workspaceId") UUID workspaceId);

    // 사용자의 모든 활성 워크스페이스 멤버십 조회
    @Query("SELECT wm FROM WorkspaceMember wm WHERE wm.userId = :userId AND wm.isActive = true")
    List<WorkspaceMember> findActiveByUserId(@Param("userId") UUID userId);

    // 특정 워크스페이스의 특정 사용자 멤버십 조회
    @Query("SELECT wm FROM WorkspaceMember wm WHERE wm.workspaceId = :workspaceId AND wm.userId = :userId AND wm.isActive = true")
    Optional<WorkspaceMember> findByWorkspaceIdAndUserId(@Param("workspaceId") UUID workspaceId, @Param("userId") UUID userId);

    // 사용자의 기본 워크스페이스 조회
    @Query("SELECT wm FROM WorkspaceMember wm WHERE wm.userId = :userId AND wm.isDefault = true AND wm.isActive = true")
    Optional<WorkspaceMember> findDefaultByUserId(@Param("userId") UUID userId);

    // 워크스페이스의 OWNER 조회
    @Query("SELECT wm FROM WorkspaceMember wm WHERE wm.workspaceId = :workspaceId AND wm.role = 'OWNER' AND wm.isActive = true")
    Optional<WorkspaceMember> findOwnerByWorkspaceId(@Param("workspaceId") UUID workspaceId);

    // 워크스페이스의 활성 멤버 수
    @Query("SELECT COUNT(wm) FROM WorkspaceMember wm WHERE wm.workspaceId = :workspaceId AND wm.isActive = true")
    long countActiveByWorkspaceId(@Param("workspaceId") UUID workspaceId);

    // 특정 역할의 모든 멤버 조회
    @Query("SELECT wm FROM WorkspaceMember wm WHERE wm.workspaceId = :workspaceId AND wm.role = :role AND wm.isActive = true")
    List<WorkspaceMember> findByWorkspaceIdAndRole(@Param("workspaceId") UUID workspaceId, @Param("role") WorkspaceMember.WorkspaceRole role);

    // 멤버 존재 여부 확인
    @Query("SELECT COUNT(wm) > 0 FROM WorkspaceMember wm WHERE wm.workspaceId = :workspaceId AND wm.userId = :userId AND wm.isActive = true")
    boolean existsByWorkspaceIdAndUserId(@Param("workspaceId") UUID workspaceId, @Param("userId") UUID userId);
}