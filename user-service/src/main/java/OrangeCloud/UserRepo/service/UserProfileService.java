package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.exception.UserNotFoundException; // ✅ UserNotFoundException을 사용
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
// import org.springframework.data.crossstore.ChangeSetPersister.NotFoundException; // 🚫 불필요한 임포트 제거
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserProfileService {

    private final UserProfileRepository userProfileRepository;

    /**
     * 사용자 프로필을 조회하고 DTO로 반환합니다. (Redis 캐시 적용)
     * 캐시 이름: "userProfile", 키: userId
     */
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfile", key = "#userId")
    // 💡 수정: 반환 타입을 UserProfileResponse DTO로 변경
    public UserProfileResponse getProfile(UUID userId) { 
        log.info("[Cacheable] Attempting to retrieve profile from DB for user: {}", userId);
        
        // DB 조회 (UserProfile 엔티티)
         UserProfile profile = userProfileRepository.findByUserId(userId)
                                           // 💡 수정: 정의된 UserNotFoundException을 사용
                                           .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
    
        // 💡 수정: DTO를 반환하도록 로직을 유지
        return UserProfileResponse.from(profile);
    }

    /**
     * 사용자 프로필 닉네임, 이메일 및 이미지 URL을 통합 업데이트하고 캐시를 무효화합니다.
     * @param userId 사용자 ID (UUID)
     * @return 업데이트된 UserProfile 엔티티 (Service 내부에서 사용되므로 엔티티 반환 유지)
     */
    @Transactional
    @CacheEvict(value = "userProfile", key = "#userId") 
    // 💡 주의: Service 내부의 CRUD 메서드는 엔티티를 반환하도록 유지
    public UserProfile updateProfile(UUID userId, String nickName, String email, String profileImageUrl) {
        log.info("[CacheEvict] Updating profile for user: userId={}, nickName={}, email={}, imageUrl={}", userId, nickName, email, profileImageUrl);

        // 1. UserProfile 조회
        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> new UserNotFoundException("프로필 업데이트 대상 사용자를 찾을 수 없습니다."));

        // 2. 닉네임 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (nickName != null && !nickName.trim().isEmpty()) {
            profile.updateNickName(nickName.trim());
            log.debug("Profile nickName updated to: {}", nickName.trim());
        }

        // 3. 이메일 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (email != null && !email.trim().isEmpty()) {
            profile.updateEmail(email.trim());
            log.debug("Profile email updated to: {}", email.trim());
        }

        // 4. 이미지 URL 업데이트
        if (profileImageUrl != null) {
            String urlToSave = profileImageUrl.trim().isEmpty() ? null : profileImageUrl.trim();
            profile.updateProfileImageUrl(urlToSave);
            log.debug("Profile image URL updated to: {}", urlToSave);
        }

        // 5. 변경된 프로필 저장
        return userProfileRepository.save(profile);
    }
}