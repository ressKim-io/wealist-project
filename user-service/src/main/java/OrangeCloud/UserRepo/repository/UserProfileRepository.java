package OrangeCloud.UserRepo.repository;

import OrangeCloud.UserRepo.entity.UserProfile;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;
import java.util.UUID;

/**
 * UserProfile 엔티티에 대한 데이터 접근 계층 (DAO)
 */
public interface UserProfileRepository extends JpaRepository<UserProfile, UUID> {
    /**
     * User ID를 기반으로 UserProfile을 조회합니다.
     * @param userId User 엔티티의 ID (FK)
     * @return UserProfile 객체 (Optional)
     */
    Optional<UserProfile> findByUserId(UUID userId);
}