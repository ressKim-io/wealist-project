package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.MessageApiResponse;
import OrangeCloud.UserRepo.dto.user.UserIdRequest;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.service.UserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/users")
@RequiredArgsConstructor
@Slf4j
@Tag(name = "User", description = "사용자 관리 API")
public class UserController {

    private final UserService userService;

    /**
     * 내 정보 조회
     * GET /api/users/me
     */
    @GetMapping("/me")
    @Operation(summary = "내 정보 조회", description = "현재 인증된 사용자의 정보를 조회합니다.")
    public ResponseEntity<User> getMyInfo(Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching user info for ID: {}", userId);
        User user = userService.getUserById(userId);
        return ResponseEntity.ok(user);
    }

    /**
     * 특정 사용자 정보 조회
     * GET /api/users/{userId}
     */
    @GetMapping("/{userId}")
    @Operation(summary = "사용자 정보 조회", description = "특정 사용자의 정보를 조회합니다.")
    public ResponseEntity<User> getUserInfo(@PathVariable UUID userId) {
        log.debug("Fetching user info for ID: {}", userId);
        User user = userService.getUserById(userId);
        return ResponseEntity.ok(user);
    }

    /**
     * 사용자 소프트 삭제
     * DELETE /api/users/me
     */
    @DeleteMapping("/me")
    @Operation(summary = "계정 삭제", description = "현재 사용자 계정을 삭제합니다.")
    public ResponseEntity<MessageApiResponse> deleteMyAccount(Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Deleting user account: {}", userId);
        userService.softDeleteUser(userId);
        return ResponseEntity.ok(MessageApiResponse.success("계정이 삭제되었습니다."));
    }

    /**
     * 사용자 소프트 삭제 (Admin)
     * DELETE /api/users/{userId}
     */
    @DeleteMapping("/{userId}")
    @Operation(summary = "사용자 삭제 (관리자용)", description = "특정 사용자를 삭제합니다.")
    public ResponseEntity<MessageApiResponse> deleteUser(@PathVariable UUID userId) {
        log.info("Deleting user: {}", userId);
        userService.softDeleteUser(userId);
        return ResponseEntity.ok(MessageApiResponse.success("사용자가 삭제되었습니다."));
    }

    /**
     * 사용자 복구
     * PUT /api/users/{userId}/restore
     */
    @PutMapping("/{userId}/restore")
    @Operation(summary = "사용자 복구", description = "삭제된 사용자를 복구합니다.")
    public ResponseEntity<MessageApiResponse> restoreUser(@PathVariable UUID userId) {
        log.info("Restoring user: {}", userId);
        userService.restoreUser(userId);
        return ResponseEntity.ok(MessageApiResponse.success("사용자가 복구되었습니다."));
    }

    /**
     * 여러 사용자 정보 일괄 조회
     * POST /api/users/batch
     */
    @PostMapping("/batch")
    @Operation(summary = "사용자 일괄 조회", description = "여러 사용자의 정보를 한 번에 조회합니다.")
    public ResponseEntity<MessageApiResponse> getUsersBatch(@RequestBody UserIdRequest request) {
        log.debug("Batch fetching users: count={}", request.getUserIds().size());
        List<User> users = userService.getUsersByIds(request.getUserIds());
        return ResponseEntity.ok(MessageApiResponse.success("사용자 조회 성공", users));
    }
}