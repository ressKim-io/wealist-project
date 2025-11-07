package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.workspace.*;
import OrangeCloud.UserRepo.service.WorkspaceService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/workspaces")
@RequiredArgsConstructor
@Slf4j
@Tag(name = "Workspace", description = "워크스페이스 관리 API")
public class WorkspaceController {

    private final WorkspaceService workspaceService;

    /**
     * 사용자가 속한 모든 워크스페이스 조회
     * GET /api/workspaces
     */
    @GetMapping
    @Operation(summary = "워크스페이스 목록 조회", description = "현재 사용자가 속한 모든 워크스페이스를 조회합니다.")
    public ResponseEntity<List<WorkspaceResponse>> getWorkspaces(Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching workspaces for user: {}", userId);
        List<WorkspaceResponse> workspaces = workspaceService.getUserWorkspaces(userId);
        return ResponseEntity.ok(workspaces);
    }

    /**
     * 워크스페이스 생성
     * POST /api/workspaces
     */
    @PostMapping
    @Operation(summary = "워크스페이스 생성", description = "새로운 워크스페이스를 생성합니다. (생성자가 OWNER)")
    public ResponseEntity<WorkspaceResponse> createWorkspace(
            Authentication authentication,
            @Valid @RequestBody CreateWorkspaceRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Creating workspace: name={}, creator={}", request.getName(), userId);
        WorkspaceResponse workspace = workspaceService.createWorkspace(request, userId);
        return ResponseEntity.ok(workspace);
    }

    /**
     * 기본 워크스페이스 설정
     * POST /api/workspaces/default
     */
    @PostMapping("/default")
    @Operation(summary = "기본 워크스페이스 설정", description = "사용자의 기본 워크스페이스를 설정합니다.")
    public ResponseEntity<Void> setDefaultWorkspace(
            Authentication authentication,
            @Valid @RequestBody SetDefaultWorkspaceRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Setting default workspace: workspaceId={}, userId={}", request.getWorkspaceId(), userId);
        workspaceService.setDefaultWorkspace(request.getWorkspaceId(), userId);
        return ResponseEntity.ok().build();
    }

    /**
     * 특정 워크스페이스 조회
     * GET /api/workspaces/{workspaceId}
     */
    @GetMapping("/{workspaceId}")
    @Operation(summary = "워크스페이스 조회", description = "특정 워크스페이스의 정보를 조회합니다. (멤버만 가능)")
    public ResponseEntity<WorkspaceResponse> getWorkspace(
            @PathVariable UUID workspaceId,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching workspace: workspaceId={}", workspaceId);
        WorkspaceResponse workspace = workspaceService.getWorkspace(workspaceId, userId);
        return ResponseEntity.ok(workspace);
    }

    /**
     * 워크스페이스 수정
     * PUT /api/workspaces/{workspaceId}
     */
    @PutMapping("/{workspaceId}")
    @Operation(summary = "워크스페이스 수정", description = "워크스페이스 정보를 수정합니다. (OWNER만 가능)")
    public ResponseEntity<WorkspaceResponse> updateWorkspace(
            @PathVariable UUID workspaceId,
            Authentication authentication,
            @Valid @RequestBody UpdateWorkspaceRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Updating workspace: workspaceId={}", workspaceId);
        WorkspaceResponse workspace = workspaceService.updateWorkspace(workspaceId, request, userId);
        return ResponseEntity.ok(workspace);
    }

    /**
     * 워크스페이스 삭제
     * DELETE /api/workspaces/{workspaceId}
     */
    @DeleteMapping("/{workspaceId}")
    @Operation(summary = "워크스페이스 삭제", description = "워크스페이스를 소프트 삭제합니다. (OWNER만 가능)")
    public ResponseEntity<Void> deleteWorkspace(
            @PathVariable UUID workspaceId,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Deleting workspace: workspaceId={}", workspaceId);
        workspaceService.deleteWorkspace(workspaceId, userId);
        return ResponseEntity.ok().build();
    }

    /**
     * 워크스페이스 검색
     * GET /api/workspaces/search
     */
    @GetMapping("/search")
    @Operation(summary = "워크스페이스 검색", description = "워크스페이스명으로 검색합니다.")
    public ResponseEntity<List<WorkspaceResponse>> searchWorkspaces(
            @RequestParam String query,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Searching workspaces: query={}", query);
        // TODO: WorkspaceService에 searchWorkspaces 메서드 추가 필요
        return ResponseEntity.ok(List.of());
    }

    // ============================================================================
    // 워크스페이스 멤버 관리
    // ============================================================================

    /**
     * 워크스페이스 멤버 목록 조회
     * GET /api/workspaces/{workspaceId}/members
     */
    @GetMapping("/{workspaceId}/members")
    @Operation(summary = "워크스페이스 멤버 조회", description = "워크스페이스의 모든 멤버를 조회합니다. (멤버만 가능)")
    public ResponseEntity<List<WorkspaceMemberResponse>> getWorkspaceMembers(
            @PathVariable UUID workspaceId,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching workspace members: workspaceId={}", workspaceId);
        List<WorkspaceMemberResponse> members = workspaceService.getWorkspaceMembers(workspaceId, userId);
        return ResponseEntity.ok(members);
    }

    /**
     * 멤버 역할 변경
     * PUT /api/workspaces/{workspaceId}/members/{memberId}/role
     */
    @PutMapping("/{workspaceId}/members/{memberId}/role")
    @Operation(summary = "멤버 역할 변경", description = "워크스페이스 멤버의 역할을 변경합니다. (OWNER만 가능)")
    public ResponseEntity<WorkspaceMemberResponse> updateMemberRole(
            @PathVariable UUID workspaceId,
            @PathVariable UUID memberId,
            Authentication authentication,
            @Valid @RequestBody UpdateMemberRoleRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Updating member role: workspaceId={}, memberId={}, newRole={}", workspaceId, memberId, request.getRoleName());
        WorkspaceMemberResponse member = workspaceService.updateMemberRole(workspaceId, memberId, request, userId);
        return ResponseEntity.ok(member);
    }

    /**
     * 멤버 제거
     * DELETE /api/workspaces/{workspaceId}/members/{memberId}
     */
    @DeleteMapping("/{workspaceId}/members/{memberId}")
    @Operation(summary = "멤버 제거", description = "워크스페이스에서 멤버를 제거합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<Void> removeMember(
            @PathVariable UUID workspaceId,
            @PathVariable UUID memberId,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Removing member: workspaceId={}, memberId={}", workspaceId, memberId);
        workspaceService.removeMember(workspaceId, memberId, userId);
        return ResponseEntity.ok().build();
    }

    // ============================================================================
    // 가입 신청 관리
    // ============================================================================

    /**
     * 워크스페이스 가입 신청
     * POST /api/workspaces/join-requests
     */
    @PostMapping("/join-requests")
    @Operation(summary = "워크스페이스 가입 신청", description = "워크스페이스 가입을 신청합니다.")
    public ResponseEntity<JoinRequestResponse> createJoinRequest(
            Authentication authentication,
            @Valid @RequestBody CreateJoinRequestRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Creating join request: workspaceId={}, userId={}", request.getWorkspaceId(), userId);
        JoinRequestResponse joinRequest = workspaceService.createJoinRequest(request.getWorkspaceId(), userId);
        return ResponseEntity.ok(joinRequest);
    }

    /**
     * 가입 신청 승인/거절
     * PUT /api/workspaces/{workspaceId}/join-requests/{requestId}
     */
    @PutMapping("/{workspaceId}/joinRequests/{requestId}")
    @Operation(summary = "가입 신청 처리", description = "가입 신청을 승인하거나 거절합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<JoinRequestResponse> updateJoinRequest(
            @PathVariable UUID workspaceId,
            @PathVariable UUID requestId,
            Authentication authentication,
            @Valid @RequestBody UpdateJoinRequestRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Updating join request: workspaceId={}, requestId={}, status={}", workspaceId, requestId, request.getStatus());
        JoinRequestResponse joinRequest = workspaceService.updateJoinRequest(workspaceId, requestId, request, userId);
        return ResponseEntity.ok(joinRequest);
    }

    /**
     * 가입 신청 목록 조회
     * GET /api/workspaces/{workspaceId}/join-requests
     */
    @GetMapping("/{workspaceId}/join-requests")
    @Operation(summary = "가입 신청 목록 조회", description = "워크스페이스의 모든 가입 신청을 조회합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<List<JoinRequestResponse>> getJoinRequests(
            @PathVariable UUID workspaceId,
            @RequestParam(required = false) String status,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching join requests: workspaceId={}, status={}", workspaceId, status);
        List<JoinRequestResponse> joinRequests = workspaceService.getJoinRequests(workspaceId, userId, status);
        return ResponseEntity.ok(joinRequests);
    }
}