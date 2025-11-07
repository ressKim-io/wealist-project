package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.exception.UserNotFoundException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class UserProfileService {

    private final UserProfileRepository userProfileRepository;

    /**
     * 사용자 프로필 생성 (회원가입 시)
     */
    public UserProfile createProfile(UUID userId, String name) {
        log.debug("Creating profile for user: {}", userId);

        UserProfile profile = UserProfile.builder()
                .userId(userId)
                .name(name)
                .build();

        UserProfile savedProfile = userProfileRepository.save(profile);
        log.debug("Profile created successfully for user: {}", userId);

        return savedProfile;
    }

    /**
     * 사용자 프로필 조회 (캐시됨)
     */
    @Cacheable(value = "userProfile", key = "#userId")
    public UserProfile getProfile(UUID userId) {
        log.debug("Fetching profile for user: {}", userId);

        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> {
                    log.warn("Profile not found for user: {}", userId);
                    return new UserNotFoundException("프로필을 찾을 수 없습니다.");
                });

        log.debug("Profile retrieved for user: {}", userId);
        return profile;
    }

    /**
     * 프로필 사진 URL 업데이트
     */
    @CacheEvict(value = "userProfile", key = "#userId")
    public UserProfile updateProfileImageUrl(UUID userId, String imageUrl) {
        log.debug("Updating profile image for user: {}", userId);

        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> {
                    log.warn("Profile not found for user: {}", userId);
                    return new UserNotFoundException("프로필을 찾을 수 없습니다.");
                });

        profile.updateProfileImageUrl(imageUrl);
        UserProfile updatedProfile = userProfileRepository.save(profile);

        log.debug("Profile image updated for user: {}", userId);
        return updatedProfile;
    }
}