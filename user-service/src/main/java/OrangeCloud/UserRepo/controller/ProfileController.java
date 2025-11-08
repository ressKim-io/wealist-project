package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileImageRequest;
import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse; // 💡 DTO 경로 수정 완료
import OrangeCloud.UserRepo.entity.UserProfile; 
import OrangeCloud.UserRepo.service.UserProfileService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.security.Principal;
import java.util.UUID;

@RestController
@RequestMapping("/api/profiles")
@RequiredArgsConstructor
@Tag(name = "UserProfile", description = "사용자 프로필 조회 및 수정 API")
@Slf4j
public class ProfileController {

    private final UserProfileService userProfileService;

    // 💡 인증 객체에서 userId (UUID String) 추출
    private UUID extractUserId(Principal principal) {
        if (principal instanceof Authentication authentication) {
            return UUID.fromString(authentication.getName());
        }
        throw new IllegalStateException("인증된 사용자 정보를 찾을 수 없습니다.");
    }

    /**
     * 내 프로필 조회
     * GET /api/profiles/me
     */
    @GetMapping("/me")
    @Operation(summary = "내 프로필 조회", description = "현재 인증된 사용자의 프로필을 조회합니다.")
    public ResponseEntity<UserProfileResponse> getMyProfile(Principal principal) {
        UUID userId = extractUserId(principal);
        // UserProfile 엔티티를 받아 DTO로 변환
        UserProfile profile = userProfileService.getProfile(userId);
        return ResponseEntity.ok(UserProfileResponse.from(profile)); 
    }

    /**
     * 프로필 사진 업데이트 (기존 엔드포인트 유지)
     * PUT /api/profiles/me/image
     */
    @PutMapping("/me/image")
    @Operation(summary = "프로필 사진 업데이트", description = "프로필 사진 URL을 업데이트합니다.")
    public ResponseEntity<UserProfileResponse> updateProfileImage( // 💡 반환 타입을 DTO로 통일
            Principal principal,
            @Valid @RequestBody UpdateProfileImageRequest request) {
        UUID userId = extractUserId(principal);
        
        // ✅ [문제 해결]: updateProfileImageUrl 대신 통합 서비스 메서드 updateProfile 호출
        // 이름과 이메일은 null로 전달하여 변경하지 않도록 합니다.
        UserProfile updatedProfile = userProfileService.updateProfile(
            userId,
            null, // 이름은 변경하지 않음
            null, // 이메일은 변경하지 않음
            request.profileImageUrl() // 이미지 URL만 업데이트
        );

        return ResponseEntity.ok(UserProfileResponse.from(updatedProfile));
    }
    
    /**
     * 인증된 사용자의 프로필 (이름 또는 이미지 URL)을 통합 업데이트합니다.
     * PUT /api/profiles/me
     */
    @Operation(summary = "내 프로필 정보 통합 업데이트", description = "인증된 사용자의 이름 또는 프로필 이미지 URL을 업데이트합니다.")
    @PutMapping("/me")
    public ResponseEntity<UserProfileResponse> updateMyProfile(
            Principal principal,
            @Valid @RequestBody UpdateProfileRequest request
    ) {
        UUID userId = extractUserId(principal);
        log.info("Received integrated profile update request for user: {}", userId);

        // ✅ 통합 서비스 메서드 호출
        UserProfile updatedProfile = userProfileService.updateProfile(
                userId,
                request.name(),
                request.email(),
                request.profileImageUrl()
        );

        return ResponseEntity.ok(UserProfileResponse.from(updatedProfile));
    }
}