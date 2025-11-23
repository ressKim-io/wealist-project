package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.userprofile.CreateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UpdateProfileRequest;
import OrangeCloud.UserRepo.dto.userprofile.UserProfileResponse;
import OrangeCloud.UserRepo.service.UserProfileService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.security.Principal;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/profiles")
@RequiredArgsConstructor
@Tag(name = "UserProfile", description = "사용자 프로필 조회 및 수정 API")
@Slf4j
public class UserProfileController {

    private final UserProfileService userProfileService;

    private UUID extractUserId(Principal principal) {
        if (principal instanceof Authentication authentication) {
            return UUID.fromString(authentication.getName());
        }
        throw new IllegalStateException("인증된 사용자 정보를 찾을 수 없습니다.");
    }

    @PostMapping
    @Operation(summary = "프로필 생성", description = "새로운 프로필을 생성합니다.")
    public ResponseEntity<UserProfileResponse> createProfile(
            Principal principal,
            @Valid @RequestBody CreateProfileRequest request) {
        UUID userId = extractUserId(principal);
        log.info("Creating profile for user: {}", userId);
        UserProfileResponse response = userProfileService.createProfile(request, userId);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/me")
    @Operation(summary = "내 프로필 조회", description = "내 프로필을 조회합니다.")
    public ResponseEntity<UserProfileResponse> getMyProfile(Principal principal) {
        UUID userId = extractUserId(principal);
        UserProfileResponse response = userProfileService.getProfile(userId);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/workspace/{workspaceId}")
    @Operation(summary = "내 조직별 프로필 조회", description = "내 조직별 프로필을 조회합니다.")
    public ResponseEntity<UserProfileResponse> getMyWorkspaceIdProfile(
            @PathVariable UUID workspaceId,
            Principal principal,
            jakarta.servlet.http.HttpServletRequest request) {
        // Enhanced logging: Log incoming request details
        log.info("=== RECEIVED REQUEST: Get Workspace Profile ===");
        log.info("Request Method: {}", request.getMethod());
        log.info("Request URI: {}", request.getRequestURI());
        log.info("Request URL: {}", request.getRequestURL());
        log.info("Path Variable - workspaceId: {}", workspaceId);
        log.info("Remote Address: {}", request.getRemoteAddr());
        log.info("Authorization Header Present: {}", request.getHeader("Authorization") != null);

        UUID userId = extractUserId(principal);
        log.info("Authenticated userId: {}", userId);

        UserProfileResponse response = userProfileService.workSpaceIdGetProfile(workspaceId, userId);

        log.info("Workspace profile retrieved successfully: workspaceId={}, userId={}, profileId={}",
                workspaceId, userId, response.getProfileId());
        log.info("=== END REQUEST: Get Workspace Profile ===");

        return ResponseEntity.ok(response);
    }

    @GetMapping("/workspace/{workspaceId}/user/{userId}")
    @Operation(
        summary = "특정 사용자의 워크스페이스 프로필 조회", 
        description = "워크스페이스 내 특정 사용자의 프로필을 조회합니다. 요청자는 해당 워크스페이스의 멤버여야 합니다."
    )
    public ResponseEntity<UserProfileResponse> getWorkspaceProfileByUserId(
            @Parameter(description = "워크스페이스 ID") @PathVariable UUID workspaceId,
            @Parameter(description = "조회할 사용자 ID") @PathVariable UUID userId,
            Principal principal) {
        UUID requestingUserId = extractUserId(principal);
        log.info("Fetching workspace profile: workspaceId={}, targetUserId={}, requestingUserId={}", 
                workspaceId, userId, requestingUserId);
        
        UserProfileResponse response = userProfileService.getWorkspaceProfileByUserId(
                workspaceId, userId, requestingUserId);
        
        log.info("Workspace profile retrieved successfully: workspaceId={}, userId={}, profileId={}", 
                workspaceId, userId, response.getProfileId());
        
        return ResponseEntity.ok(response);
    }

    @GetMapping("/all/me")
    @Operation(summary = "내 모든 프로필 조회", description = "내 모든 프로필을 조회합니다.")
    public ResponseEntity<List<UserProfileResponse>> getAllMyProfile(Principal principal) {
        UUID userId = extractUserId(principal);
        List<UserProfileResponse> response = userProfileService.getAllProfiles(userId);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/me")
    @Operation(summary = "내 프로필 정보 통합 업데이트", description = "인증된 사용자의 이름 또는 프로필 이미지 URL을 업데이트합니다.")
    public ResponseEntity<UserProfileResponse> updateMyProfile(
            Principal principal,
            @Valid @RequestBody UpdateProfileRequest request) {
        UUID userId = extractUserId(principal);
        log.info("Received integrated profile update request for user: {}", userId);

        // Ensure the userId in the request matches the authenticated user
        if (!request.userId().equals(userId)) {
            throw new IllegalArgumentException("User ID in request does not match authenticated user.");
        }

        UserProfileResponse updatedProfile = userProfileService.updateProfile(request);
        return ResponseEntity.ok(updatedProfile);
    }

    @DeleteMapping("/{workspaceId}")
    @Operation(summary = "프로필 삭제", description = "특정 워크스페이스의 프로필을 삭제합니다.")
    public ResponseEntity<Void> deleteProfile(
            Principal principal,
            @PathVariable UUID workspaceId) {
        UUID userId = extractUserId(principal);
        log.info("Deleting profile for user: {} in workspace: {}", userId, workspaceId);
        userProfileService.deleteProfile(userId, workspaceId);
        return ResponseEntity.noContent().build();
    }

}