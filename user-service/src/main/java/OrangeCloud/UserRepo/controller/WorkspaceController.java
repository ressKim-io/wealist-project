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

    // ============================================================================
    // 기본 CRUD 및 조회 API
    // ============================================================================

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
        log.info("Creating workspace: name={}, creator={}", request.getWorkspaceName(), userId);
        WorkspaceResponse workspace = workspaceService.createWorkspace(request, userId);
        return ResponseEntity.ok(workspace);
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

    // ============================================================================
    // Workspace 설정 API (프론트엔드 모달 대응)
    // ============================================================================

    /**
     * 워크스페이스 설정 조회
     * GET /api/workspaces/{workspaceId}/settings
     */
    @GetMapping("/{workspaceId}/settings")
    @Operation(summary = "워크스페이스 설정 조회", description = "특정 워크스페이스의 공개/승인 설정 등을 조회합니다. (멤버만 가능)")
    public ResponseEntity<WorkspaceSettingsResponse> getWorkspaceSettings(
            @PathVariable UUID workspaceId,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching workspace settings: workspaceId={}", workspaceId);
        WorkspaceSettingsResponse settings = workspaceService.getWorkspaceSettings(workspaceId, userId); // Service 호출
        return ResponseEntity.ok(settings);
    }

    /**
     * 워크스페이스 설정 수정
     * PUT /api/workspaces/{workspaceId}/settings
     */
    @PutMapping("/{workspaceId}/settings")
    @Operation(summary = "워크스페이스 설정 수정", description = "워크스페이스의 설정 정보를 수정합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<WorkspaceSettingsResponse> updateWorkspaceSettings(
            @PathVariable UUID workspaceId,
            Authentication authentication,
            @Valid @RequestBody UpdateWorkspaceSettingsRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Updating workspace settings: workspaceId={}", workspaceId);
        WorkspaceSettingsResponse settings = workspaceService.updateWorkspaceSettings(workspaceId, request, userId); // Service
                                                                                                                     // 호출
        return ResponseEntity.ok(settings);
    }

    // ============================================================================
    // 기타 API
    // ============================================================================

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
        log.debug("Searching workspaces: query={}", query);
        // TODO: searchWorkspaces 서비스 구현
        return ResponseEntity.ok(List.of());
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
     * 승인 대기 회원 목록 조회 (프론트엔드 '/pendingMembers' 경로 지원)
     * GET /api/workspaces/{workspaceId}/pendingMembers
     */
    @GetMapping("/{workspaceId}/pendingMembers")
    @Operation(summary = "승인 대기 회원 목록 조회", description = "워크스페이스의 PENDING 상태 가입 신청 목록을 조회합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<List<JoinRequestResponse>> getPendingMembers(
            @PathVariable UUID workspaceId,
            Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        log.debug("Fetching pending members: workspaceId={}", workspaceId);

        // PENDING 상태만 조회하도록 service 메서드를 호출
        List<JoinRequestResponse> pendingRequests = workspaceService.getJoinRequests(workspaceId, userId, "PENDING");

        // Note: 프론트엔드 DTO(PendingMember)와 백엔드 DTO(JoinRequestResponse)의 필드 매핑은 프론트엔드에서
        // 처리됨을 가정
        return ResponseEntity.ok(pendingRequests);
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
        log.info("Updating member role: workspaceId={}, memberId={}, newRole={}", workspaceId, memberId,
                request.getRoleName());
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

    /**
     * 회원 승인
     * POST /api/workspaces/{workspaceId}/members/{userId}/approve
     */
    @PostMapping("/{workspaceId}/members/{userId}/approve")
    @Operation(summary = "가입 신청 승인", description = "특정 사용자의 가입 신청을 승인하고 멤버로 추가합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<Void> approveMember(
            @PathVariable UUID workspaceId,
            @PathVariable UUID userId, // 여기서 userId는 신청한 사용자 ID
            Authentication authentication) {
        UUID adminId = UUID.fromString(authentication.getName());
        log.info("Approving member: workspaceId={}, userId={}, adminId={}", workspaceId, userId, adminId);
        workspaceService.approveJoinRequest(workspaceId, userId, adminId); // Service 호출
        return ResponseEntity.ok().build();
    }

    /**
     * 회원 거절
     * POST /api/workspaces/{workspaceId}/members/{userId}/reject
     */
    @PostMapping("/{workspaceId}/members/{userId}/reject")
    @Operation(summary = "가입 신청 거절", description = "특정 사용자의 가입 신청을 거절합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<Void> rejectMember(
            @PathVariable UUID workspaceId,
            @PathVariable UUID userId, // 여기서 userId는 신청한 사용자 ID
            Authentication authentication) {
        UUID adminId = UUID.fromString(authentication.getName());
        log.info("Rejecting member: workspaceId={}, userId={}, adminId={}", workspaceId, userId, adminId);
        workspaceService.rejectJoinRequest(workspaceId, userId, adminId); // Service 호출
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
     * 가입 신청 승인/거절 (기존 PUT 매핑)
     * PUT /api/workspaces/{workspaceId}/joinRequests/{requestId}
     */
    @PutMapping("/{workspaceId}/joinRequests/{requestId}")
    @Operation(summary = "가입 신청 처리", description = "가입 신청을 승인하거나 거절합니다. (OWNER/ADMIN만 가능)")
    public ResponseEntity<JoinRequestResponse> updateJoinRequest(
            @PathVariable UUID workspaceId,
            @PathVariable UUID requestId,
            Authentication authentication,
            @Valid @RequestBody UpdateJoinRequestRequest request) {
        UUID userId = UUID.fromString(authentication.getName());
        log.info("Updating join request: workspaceId={}, requestId={}, status={}", workspaceId, requestId,
                request.getStatus());
        JoinRequestResponse joinRequest = workspaceService.updateJoinRequest(workspaceId, requestId, request, userId);
        return ResponseEntity.ok(joinRequest);
    }

    /**
     * 가입 신청 목록 조회
     * GET /api/workspaces/{workspaceId}/joinRequests
     */
    @GetMapping("/{workspaceId}/joinRequests")
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