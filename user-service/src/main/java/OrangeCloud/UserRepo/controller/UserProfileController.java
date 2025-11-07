package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.MessageApiResponse;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileImageRequest;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.service.UserProfileService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@RestController
@RequestMapping("/api/profiles")
@RequiredArgsConstructor
@Tag(name = "UserProfile", description = "사용자 프로필 관련 API")
public class UserProfileController {

    private final UserProfileService userProfileService;

    /**
     * 내 프로필 조회
     * GET /api/profiles/me
     */
    @GetMapping("/me")
    @Operation(summary = "내 프로필 조회", description = "현재 인증된 사용자의 프로필을 조회합니다.")
    public ResponseEntity<UserProfile> getMyProfile(Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        UserProfile profile = userProfileService.getProfile(userId);
        return ResponseEntity.ok(profile);
    }

    /**
     * 프로필 사진 업데이트
     * PUT /api/profiles/me/image
     */
    @PutMapping("/me/image")
    @Operation(summary = "프로필 사진 업데이트", description = "프로필 사진 URL을 업데이트합니다.")
    public ResponseEntity<UserProfile> updateProfileImage(
            Authentication authentication,
            @Valid @RequestBody UpdateProfileImageRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        UserProfile profile = userProfileService.updateProfileImageUrl(userId, request.getProfileImageUrl());
        return ResponseEntity.ok(profile);
    }
}