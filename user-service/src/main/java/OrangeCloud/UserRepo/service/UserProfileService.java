package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.dto.userprofile.AttachmentResponse;
import OrangeCloud.UserRepo.dto.userprofile.CreateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse;
import OrangeCloud.UserRepo.entity.Attachment;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.AttachmentRepository;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.WorkspaceMemberRepository;
import OrangeCloud.UserRepo.exception.UserNotFoundException;
import OrangeCloud.UserRepo.exception.CustomException;
import OrangeCloud.UserRepo.exception.ErrorCode;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileRequest;

import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserProfileService {

    private final UserProfileRepository userProfileRepository;
    private final UserRepository userRepository;
    private final WorkspaceMemberRepository workspaceMemberRepository;
    private final AttachmentRepository attachmentRepository;
    private static final UUID DEFAULT_WORKSPACE_ID = UUID.fromString("00000000-0000-0000-0000-000000000000");

    @Transactional
    public UserProfileResponse createProfile(CreateProfileRequest request, UUID userId) {
        log.info("Creating profile for user: {}", userId);
        UserProfile userProfile = UserProfile.create(
                request.workspaceId(),
                userId,
                request.nickName(),
                request.email(),
                null
        );
        UserProfile savedProfile = userProfileRepository.save(userProfile);
        return UserProfileResponse.from(savedProfile);
    }

    /**
     * 사용자 프로필을 조회하고 DTO로 반환합니다. (Redis 캐시 적용)
     * 캐시 이름: "userProfile", 키: userId
     */
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfile", key = "#userId")
    // 💡 수정: 반환 타입을 UserProfileResponse DTO로 변경
    public UserProfileResponse getProfile(UUID userId) {
        log.info("[Cacheable] Attempting to retrieve profile from DB for user: {}", userId);
        UUID defaultId = DEFAULT_WORKSPACE_ID;
        // DB 조회 (UserProfile 엔티티)
        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(DEFAULT_WORKSPACE_ID,userId)
                // 💡 수정: 정의된 UserNotFoundException을 사용
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        // 프로필 이미지 첨부파일 조회 (있으면 하나만)
        AttachmentResponse attachment = getProfileImageAttachment(profile.getProfileId());

        // 💡 수정: DTO를 반환하도록 로직을 유지
        return UserProfileResponse.from(profile, attachment);
    }
    
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfile", key = "#userId")
    // 💡 수정: 반환 타입을 UserProfileResponse DTO로 변경
    public UserProfileResponse workSpaceIdGetProfile(UUID workspaceId,UUID userId) {
        log.info("[Cacheable] Attempting to retrieve profile from DB for user: {}", userId);

        // DB 조회 (UserProfile 엔티티)
        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(workspaceId,userId)
                // 💡 수정: 정의된 UserNotFoundException을 사용
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        // 프로필 이미지 첨부파일 조회 (있으면 하나만)
        AttachmentResponse attachment = getProfileImageAttachment(profile.getProfileId());

        // 💡 수정: DTO를 반환하도록 로직을 유지
        return UserProfileResponse.from(profile, attachment);
    }
    
    /**
     * 프로필 이미지 첨부파일 조회
     * 프로필 이미지는 하나만 존재하므로 첫 번째 것을 반환
     */
    private AttachmentResponse getProfileImageAttachment(UUID profileId) {
        List<Attachment> attachments = attachmentRepository.findByEntityTypeAndEntityIdAndDeletedAtIsNull(
                Attachment.EntityType.USER_PROFILE,
                profileId
        );
        
        // 프로필 이미지가 있으면 첫 번째 것 반환, 없으면 null
        return attachments.isEmpty() ? null : AttachmentResponse.from(attachments.get(0));
    }

    /**
     * 특정 사용자의 워크스페이스 프로필을 조회합니다.
     * 요청자는 해당 워크스페이스의 멤버여야 합니다.
     * 워크스페이스 전용 프로필이 없는 경우 사용자의 기본 프로필 정보를 반환합니다.
     * 
     * @param workspaceId 워크스페이스 ID
     * @param targetUserId 조회할 사용자 ID
     * @param requestingUserId 요청하는 사용자 ID
     * @return 사용자 프로필 응답 DTO (워크스페이스 프로필 또는 기본 프로필)
     * @throws CustomException 요청자가 워크스페이스 멤버가 아닌 경우 (403 Forbidden)
     * @throws UserNotFoundException 대상 사용자가 시스템에 존재하지 않는 경우 (404 Not Found)
     */
    @Transactional(readOnly = true)
    public UserProfileResponse getWorkspaceProfileByUserId(
            UUID workspaceId, 
            UUID targetUserId, 
            UUID requestingUserId) {
        
        log.info("Fetching workspace profile: workspaceId={}, targetUserId={}, requestingUserId={}", 
                workspaceId, targetUserId, requestingUserId);
        
        // 1. 요청자가 워크스페이스 멤버인지 검증
        boolean isMember = workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspaceId, requestingUserId);
        if (!isMember) {
            log.warn("Access denied: User {} is not a member of workspace {}", requestingUserId, workspaceId);
            throw new CustomException(ErrorCode.HANDLE_ACCESS_DENIED, 
                    "You must be a member of this workspace to view member profiles");
        }
        
        // 2. 워크스페이스 전용 프로필 조회 시도
        Optional<UserProfile> workspaceProfile = userProfileRepository.findByWorkspaceIdAndUserId(workspaceId, targetUserId);
        
        if (workspaceProfile.isPresent()) {
            // 워크스페이스 프로필이 존재하는 경우 - 반환
            log.info("Workspace profile found: profileId={}, workspaceId={}, userId={}", 
                    workspaceProfile.get().getProfileId(), workspaceId, targetUserId);
            return UserProfileResponse.from(workspaceProfile.get());
        }
        
        // 3. 워크스페이스 프로필이 없는 경우 - 기본 프로필로 fallback
        log.info("Workspace profile not found, falling back to default profile: workspaceId={}, userId={}", 
                workspaceId, targetUserId);
        
        User user = userRepository.findById(targetUserId)
                .orElseThrow(() -> {
                    log.warn("User not found: userId={}", targetUserId);
                    return new UserNotFoundException("User not found in the system");
                });
        
        // 4. 기본 프로필로 UserProfileResponse 생성
        log.info("Returning default profile for user: userId={}, email={}", targetUserId, user.getEmail());
        return UserProfileResponse.builder()
                .profileId(null)  // 워크스페이스 전용 프로필 ID 없음
                .workspaceId(workspaceId)
                .userId(user.getUserId())
                .nickName(null)  // 커스텀 닉네임 없음
                .email(user.getEmail())
                .profileImageUrl(null)  // 커스텀 프로필 이미지 없음
                .build();
    }

    // 해당 사용자id에 따른 모든 프로필 가져오기
    @Transactional(readOnly = true)
    @Cacheable(value = "userProfiles", key = "#userId")
    public List<UserProfileResponse> getAllProfiles(UUID userId) {
        log.info("[Cacheable] Attempting to retrieve all profiles from DB for user: {}", userId);

        // DB 조회 (해당 사용자의 모든 UserProfile 엔티티)
        List<UserProfile> profiles = userProfileRepository.findAllByUserId(userId);

        if (profiles.isEmpty()) {
            throw new UserNotFoundException("사용자의 프로필을 찾을 수 없습니다.");
        }

        // DTO 변환 후 반환
        return profiles.stream()
                .map(UserProfileResponse::from)
                .collect(Collectors.toList());

    }



    /**
     * 사용자 프로필 닉네임, 이메일 및 이미지 URL을 통합 업데이트하고 캐시를 무효화합니다.
     * @param request
     * @return 업데이트된 UserProfile 엔티티 (Service 내부에서 사용되므로 엔티티 반환 유지)
     */
    @Transactional
    @CacheEvict(value = "userProfile", key = "#request.userId")
    public UserProfileResponse updateProfile(UpdateProfileRequest request) {
        log.info("[CacheEvict] Updating profile for user: userId={}, nickName={}, email={}, imageUrl={}", request.userId(), request.nickName(), request.email(), request.profileImageUrl());

        // 1. UserProfile 조회
        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(request.workspaceId(), request.userId())
                .orElseThrow(() -> new UserNotFoundException("프로필 업데이트 대상 사용자를 찾을 수 없습니다."));

        // 2. 닉네임 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (request.nickName() != null && !request.nickName().trim().isEmpty()) {
            profile.updateNickName(request.nickName().trim());
            log.debug("Profile nickName updated to: {}", request.nickName().trim());
        }

        // 3. 이메일 업데이트 (값이 존재하고 비어있지 않을 경우에만)
        if (request.email() != null && !request.email().trim().isEmpty()) {
            profile.updateEmail(request.email().trim());
            log.debug("Profile email updated to: {}", request.email().trim());
        }

        // 4. 이미지 URL 업데이트
        if (request.profileImageUrl() != null) {
            String urlToSave = request.profileImageUrl().trim().isEmpty() ? null : request.profileImageUrl().trim();
            profile.updateProfileImageUrl(urlToSave);
            log.debug("Profile image URL updated to: {}", urlToSave);
        }

        // 5. 변경된 프로필 저장
        UserProfile updatedProfile = userProfileRepository.save(profile);
        return UserProfileResponse.from(updatedProfile);
    }


    /**
     * 사용자 프로필을 삭제하고 캐시를 무효화합니다.
     * @param userId 사용자 ID (UUID)
     * @param workspaceId 워크스페이스 ID (UUID)
     */
    @Transactional
    @CacheEvict(value = "userProfile", key = "#userId")
    public void deleteProfile(UUID userId, UUID workspaceId) {
        log.info("[CacheEvict] Deleting profile for user: userId={}, workspaceId={}", userId, workspaceId);

        UserProfile profile = userProfileRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> new UserNotFoundException("삭제할 프로필을 찾을 수 없습니다."));

        userProfileRepository.delete(profile);
    }

}