package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.Workspace;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface WorkspaceRepository extends JpaRepository<Workspace, UUID> {

    // ============================================================================
    // 소프트 삭제/복구
    // ============================================================================

    /**
     * Workspace 소프트 삭제
     */
    @Modifying
    @Transactional
    @Query("UPDATE Workspace w SET w.isActive = false, w.deletedAt = CURRENT_TIMESTAMP WHERE w.groupId = :workspaceId")
    int softDeleteById(@Param("workspaceId") UUID workspaceId);

    /**
     * Workspace 재활성화
     */
    @Modifying
    @Transactional
    @Query("UPDATE Workspace w SET w.isActive = true, w.deletedAt = null WHERE w.groupId = :workspaceId")
    int reactivateById(@Param("workspaceId") UUID workspaceId);

    // ============================================================================
    // Workspace 조회
    // ============================================================================

    /**
     * 모든 활성화된 Workspace 조회
     */
    @Query("SELECT w FROM Workspace w WHERE w.isActive = true ORDER BY w.createdAt DESC")
    List<Workspace> findAllActiveWorkspaces();

    /**
     * ID로 활성화된 Workspace 조회
     */
    @Query("SELECT w FROM Workspace w WHERE w.groupId = :workspaceId AND w.isActive = true")
    Optional<Workspace> findActiveById(@Param("workspaceId") UUID workspaceId);

    /**
     * 이름으로 활성화된 Workspace 검색
     */
    @Query("SELECT w FROM Workspace w WHERE w.name LIKE %:name% AND w.isActive = true ORDER BY w.createdAt DESC")
    List<Workspace> findActiveByNameContaining(@Param("name") String name);

    /**
     * 회사명으로 활성화된 Workspace 조회
     */
    @Query("SELECT w FROM Workspace w WHERE w.companyName = :companyName AND w.isActive = true ORDER BY w.createdAt ASC")
    List<Workspace> findActiveByCompanyName(@Param("companyName") String companyName);

    /**
     * 회사명과 이름으로 활성화된 Workspace 조회
     */
    @Query("SELECT w FROM Workspace w WHERE w.companyName = :companyName AND w.name = :name AND w.isActive = true")
    Optional<Workspace> findActiveByCompanyNameAndName(@Param("companyName") String companyName, @Param("name") String name);

    // ============================================================================
    // 중복 체크
    // ============================================================================

    /**
     * 회사명 중복 체크
     */
    @Query("SELECT COUNT(w) > 0 FROM Workspace w WHERE w.companyName = :companyName AND w.isActive = true")
    boolean existsActiveByCompanyName(@Param("companyName") String companyName);

    // ============================================================================
    // 통계
    // ============================================================================

    /**
     * 활성화된 Workspace 수
     */
    @Query("SELECT COUNT(w) FROM Workspace w WHERE w.isActive = true")
    long countActiveWorkspaces();

    /**
     * 회사별 Workspace 수
     */
    @Query("SELECT COUNT(w) FROM Workspace w WHERE w.companyName = :companyName AND w.isActive = true")
    long countActiveByCompanyName(@Param("companyName") String companyName);

    /**
     * 비활성화된 Workspace 조회 (관리자용)
     */
    @Query("SELECT w FROM Workspace w WHERE w.isActive = false ORDER BY w.deletedAt DESC")
    List<Workspace> findInactiveWorkspaces();
}