package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.exception.UserNotFoundException; // ⚠️ UserNotFoundException 클래스가 필요합니다.
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserProfileService {

    private final UserProfileRepository userProfileRepository;

    /**
     * 사용자 프로필을 조회합니다. (Redis 캐시 적용)
     * 캐시 이름: "userProfile", 키: userId
     */
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfile", key = "#userId")
    public UserProfile getProfile(UUID userId) {
        log.info("[Cacheable] Attempting to retrieve profile from DB for user: {}", userId);
        
        // userId를 FK로 사용하여 조회
        return userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> new UserNotFoundException("프로필 조회 대상 사용자를 찾을 수 없습니다: " + userId));
    }

    /**
     * 사용자 프로필 이름, 이메일 및 이미지 URL을 통합 업데이트하고 캐시를 무효화합니다.
     * @param userId 사용자 ID (UUID)
     * @param name 업데이트할 이름 (null 가능)
     * @param email 업데이트할 이메일 (null 가능)
     * @param profileImageUrl 업데이트할 이미지 URL (null 또는 빈 문자열 가능)
     * @return 업데이트된 UserProfile 엔티티
     */
    @Transactional
    @CacheEvict(value = "userProfile", key = "#userId") // 💡 캐시 무효화: 다음 조회 시 최신 DB 데이터 로드
    public UserProfile updateProfile(UUID userId, String name, String email, String profileImageUrl) {
        log.info("[CacheEvict] Updating profile for user: userId={}, name={}, email={}, imageUrl={}", userId, name, email, profileImageUrl);

        // 1. UserProfile 조회
        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> new UserNotFoundException("프로필 업데이트 대상 사용자를 찾을 수 없습니다."));

        // 2. 이름 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (name != null && !name.trim().isEmpty()) {
            profile.updateName(name.trim());
            log.debug("Profile name updated to: {}", name.trim());
        }

        // 3. 이메일 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (email != null && !email.trim().isEmpty()) {
            profile.updateEmail(email.trim());
            log.debug("Profile email updated to: {}", email.trim());
        }

        // 4. 이미지 URL 업데이트 
        // 클라이언트에서 명시적으로 업데이트 요청이 들어왔을 때만 처리합니다.
        // 클라이언트에서 빈 문자열("")을 보내면 URL을 null로 저장하여 기본 이미지를 사용하도록 처리합니다.
        if (profileImageUrl != null) {
            String urlToSave = profileImageUrl.trim().isEmpty() ? null : profileImageUrl.trim();
            profile.updateProfileImageUrl(urlToSave);
            log.debug("Profile image URL updated to: {}", urlToSave);
        }

        // 5. 변경된 프로필 저장 (save 메서드가 이미 @Transactional 안에서 호출되므로, 변경 감지 후 자동 커밋됩니다.)
        return userProfileRepository.save(profile);
    }
}