package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.entity.Workspace;
import OrangeCloud.UserRepo.entity.WorkspaceMember;
import OrangeCloud.UserRepo.entity.WorkspaceJoinRequest;
import OrangeCloud.UserRepo.repository.WorkspaceMemberRepository;
import OrangeCloud.UserRepo.repository.WorkspaceJoinRequestRepository;
import OrangeCloud.UserRepo.repository.WorkspaceRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.dto.workspace.*;
import OrangeCloud.UserRepo.exception.UserNotFoundException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class WorkspaceService {

    private final WorkspaceRepository workspaceRepository;
    private final WorkspaceMemberRepository workspaceMemberRepository;
    private final WorkspaceJoinRequestRepository workspaceJoinRequestRepository;
    private final UserRepository userRepository;
    private final UserProfileRepository userProfileRepository;

    // ============================================================================
    // Workspace 생성/수정/삭제
    // ============================================================================

    /**
     * 새로운 Workspace 생성 (생성자가 자동으로 OWNER)
     */
    public WorkspaceResponse createWorkspace(CreateWorkspaceRequest request, UUID creatorId) {
        log.info("Creating workspace: name={}, creator={}", request.getName(), creatorId);

        // 생성자 정보 조회
        User creator = userRepository.findById(creatorId)
                .orElseThrow(() -> {
                    log.warn("User not found: {}", creatorId);
                    return new UserNotFoundException("사용자를 찾을 수 없습니다.");
                });

        // Workspace 생성
        Workspace workspace = Workspace.builder()
                .name(request.getName())
                .companyName(request.getName())
                .isActive(true)
                .build();

        Workspace savedWorkspace = workspaceRepository.save(workspace);
        log.info("Workspace created: workspaceId={}", savedWorkspace.getGroupId());

        // 생성자를 OWNER로 추가
        WorkspaceMember ownerMember = WorkspaceMember.builder()
                .workspaceId(savedWorkspace.getGroupId())
                .userId(creatorId)
                .role(WorkspaceMember.WorkspaceRole.OWNER)
                .isDefault(true)
                .isActive(true)
                .build();

        workspaceMemberRepository.save(ownerMember);
        log.info("Creator added as OWNER: workspaceId={}, userId={}", savedWorkspace.getGroupId(), creatorId);

        UserProfile creatorProfile = userProfileRepository.findByUserId(creatorId)
                .orElseThrow(() -> {
                    log.warn("Profile not found for user: {}", creatorId);
                    return new UserNotFoundException("프로필을 찾을 수 없습니다.");
                });

        return convertToWorkspaceResponse(savedWorkspace, creator, creatorProfile);
    }

    /**
     * Workspace 정보 수정 (OWNER만)
     */
    public WorkspaceResponse updateWorkspace(UUID workspaceId, UpdateWorkspaceRequest request, UUID requesterId) {
        log.info("Updating workspace: workspaceId={}, requester={}", workspaceId, requesterId);

        checkWorkspaceOwner(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        if (request.getName() != null && !request.getName().isEmpty()) {
            workspace.setName(request.getName());
        }
        if (request.getDescription() != null) {
            workspace.setCompanyName(request.getDescription());
        }

        Workspace updated = workspaceRepository.save(workspace);
        
        // OWNER 정보 조회
        WorkspaceMember ownerMember = workspaceMemberRepository.findOwnerByWorkspaceId(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace owner not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace owner not found");
                });
        
        User owner = userRepository.findById(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile ownerProfile = userProfileRepository.findByUserId(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        log.info("Workspace updated: workspaceId={}", workspaceId);
        return convertToWorkspaceResponse(updated, owner, ownerProfile);
    }

    /**
     * Workspace 소프트 삭제 (OWNER만)
     */
    public void deleteWorkspace(UUID workspaceId, UUID requesterId) {
        log.info("Deleting workspace: workspaceId={}, requester={}", workspaceId, requesterId);

        checkWorkspaceOwner(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        workspace.softDelete();
        workspaceRepository.save(workspace);
        log.info("Workspace deleted: workspaceId={}", workspaceId);
    }

    // ============================================================================
    // Workspace 조회
    // ============================================================================

    /**
     * Workspace 조회 (멤버만 가능)
     */
    @Transactional(readOnly = true)
    public WorkspaceResponse getWorkspace(UUID workspaceId, UUID requesterId) {
        log.debug("Fetching workspace: workspaceId={}", workspaceId);

        checkWorkspaceMember(workspaceId, requesterId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        // OWNER 정보 조회
        WorkspaceMember ownerMember = workspaceMemberRepository.findOwnerByWorkspaceId(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace owner not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace owner not found");
                });

        User owner = userRepository.findById(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile ownerProfile = userProfileRepository.findByUserId(ownerMember.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToWorkspaceResponse(workspace, owner, ownerProfile);
    }

    /**
     * 사용자가 속한 모든 Workspace 조회
     */
    @Transactional(readOnly = true)
    public List<WorkspaceResponse> getUserWorkspaces(UUID userId) {
        log.debug("Fetching workspaces for user: userId={}", userId);

        List<WorkspaceMember> members = workspaceMemberRepository.findActiveByUserId(userId);

        return members.stream()
                .map(member -> {
                    Workspace workspace = workspaceRepository.findById(member.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace not found"));
                    WorkspaceMember owner = workspaceMemberRepository.findOwnerByWorkspaceId(member.getWorkspaceId())
                            .orElseThrow(() -> new IllegalArgumentException("Workspace owner not found"));
                    User ownerUser = userRepository.findById(owner.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
                    UserProfile ownerProfile = userProfileRepository.findByUserId(owner.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
                    return convertToWorkspaceResponse(workspace, ownerUser, ownerProfile);
                })
                .collect(Collectors.toList());
    }

    /**
     * 기본 Workspace 설정
     */
    public void setDefaultWorkspace(UUID workspaceId, UUID userId) {
        log.info("Setting default workspace: workspaceId={}, userId={}", workspaceId, userId);

        checkWorkspaceMember(workspaceId, userId);

        // 기존 default 해제
        List<WorkspaceMember> userMembers = workspaceMemberRepository.findActiveByUserId(userId);
        userMembers.forEach(member -> {
            member.setIsDefault(false);
            workspaceMemberRepository.save(member);
        });

        // 새로운 default 설정
        WorkspaceMember member = workspaceMemberRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> {
                    log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
                    return new IllegalArgumentException("User is not a member of this workspace");
                });

        member.setIsDefault(true);
        workspaceMemberRepository.save(member);
        log.info("Default workspace set: workspaceId={}", workspaceId);
    }

    // ============================================================================
    // Workspace 멤버 관리
    // ============================================================================

    /**
     * Workspace 멤버 목록 조회 (멤버만 가능)
     */
    @Transactional(readOnly = true)
    public List<WorkspaceMemberResponse> getWorkspaceMembers(UUID workspaceId, UUID requesterId) {
        log.debug("Fetching members for workspace: workspaceId={}", workspaceId);

        checkWorkspaceMember(workspaceId, requesterId);

        List<WorkspaceMember> members = workspaceMemberRepository.findActiveByWorkspaceId(workspaceId);

        return members.stream()
                .map(this::convertToWorkspaceMemberResponse)
                .collect(Collectors.toList());
    }

    /**
     * 멤버 역할 변경 (OWNER만)
     */
    public WorkspaceMemberResponse updateMemberRole(UUID workspaceId, UUID memberId, UpdateMemberRoleRequest request, UUID requesterId) {
        log.info("Updating member role: workspaceId={}, memberId={}, newRole={}, requester={}", 
                workspaceId, memberId, request.getRoleName(), requesterId);

        checkWorkspaceOwner(workspaceId, requesterId);

        WorkspaceMember member = workspaceMemberRepository.findById(memberId)
                .orElseThrow(() -> {
                    log.warn("Member not found: {}", memberId);
                    return new IllegalArgumentException("Member not found");
                });

        if (!member.getWorkspaceId().equals(workspaceId)) {
            log.warn("Member does not belong to workspace: memberId={}, workspaceId={}", memberId, workspaceId);
            throw new IllegalArgumentException("Member does not belong to this workspace");
        }

        member.setRole(WorkspaceMember.WorkspaceRole.valueOf(request.getRoleName()));
        WorkspaceMember updated = workspaceMemberRepository.save(member);

        return convertToWorkspaceMemberResponse(updated);
    }

    /**
     * 멤버 제거 (OWNER/ADMIN만, OWNER는 제거 불가)
     */
    public void removeMember(UUID workspaceId, UUID memberId, UUID requesterId) {
        log.info("Removing member: workspaceId={}, memberId={}, requester={}", workspaceId, memberId, requesterId);

        checkWorkspaceAdminOrOwner(workspaceId, requesterId);

        WorkspaceMember member = workspaceMemberRepository.findById(memberId)
                .orElseThrow(() -> {
                    log.warn("Member not found: {}", memberId);
                    return new IllegalArgumentException("Member not found");
                });

        if (!member.getWorkspaceId().equals(workspaceId)) {
            log.warn("Member does not belong to workspace: memberId={}, workspaceId={}", memberId, workspaceId);
            throw new IllegalArgumentException("Member does not belong to this workspace");
        }

        if (member.getRole() == WorkspaceMember.WorkspaceRole.OWNER) {
            log.warn("Cannot remove workspace owner: memberId={}", memberId);
            throw new IllegalArgumentException("Cannot remove workspace owner");
        }

        if (member.getUserId().equals(requesterId)) {
            log.warn("User cannot remove themselves: userId={}", requesterId);
            throw new IllegalArgumentException("Cannot remove yourself");
        }

        member.setIsActive(false);
        workspaceMemberRepository.save(member);
        log.info("Member removed: workspaceId={}, memberId={}", workspaceId, memberId);
    }

    // ============================================================================
    // 가입 신청 관리
    // ============================================================================

    /**
     * Workspace 가입 신청
     */
    public JoinRequestResponse createJoinRequest(UUID workspaceId, UUID userId) {
        log.info("Creating join request: workspaceId={}, userId={}", workspaceId, userId);

        Workspace workspace = workspaceRepository.findById(workspaceId)
                .orElseThrow(() -> {
                    log.warn("Workspace not found: {}", workspaceId);
                    return new IllegalArgumentException("Workspace not found");
                });

        User user = userRepository.findById(userId)
                .orElseThrow(() -> {
                    log.warn("User not found: {}", userId);
                    return new UserNotFoundException("사용자를 찾을 수 없습니다.");
                });

        if (workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspaceId, userId)) {
            log.warn("User is already a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("User is already a member of this workspace");
        }

        if (workspaceJoinRequestRepository.existsPendingByWorkspaceIdAndUserId(workspaceId, userId)) {
            log.warn("Join request already exists: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Join request already exists");
        }

        WorkspaceJoinRequest request = WorkspaceJoinRequest.builder()
                .workspaceId(workspaceId)
                .userId(userId)
                .status(WorkspaceJoinRequest.JoinRequestStatus.PENDING)
                .build();

        WorkspaceJoinRequest saved = workspaceJoinRequestRepository.save(request);
        log.info("Join request created: requestId={}", saved.getJoinRequestId());

        UserProfile userProfile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToJoinRequestResponse(saved, user, userProfile);
    }

    /**
     * 가입 신청 승인/거절 (OWNER/ADMIN만)
     */
    public JoinRequestResponse updateJoinRequest(UUID workspaceId, UUID requestId, UpdateJoinRequestRequest updateRequest, UUID responderId) {
        log.info("Updating join request: workspaceId={}, requestId={}, status={}, responder={}", 
                workspaceId, requestId, updateRequest.getStatus(), responderId);

        checkWorkspaceAdminOrOwner(workspaceId, responderId);

        WorkspaceJoinRequest joinRequest = workspaceJoinRequestRepository.findById(requestId)
                .orElseThrow(() -> {
                    log.warn("Join request not found: {}", requestId);
                    return new IllegalArgumentException("Join request not found");
                });

        if (!joinRequest.getWorkspaceId().equals(workspaceId)) {
            log.warn("Join request does not belong to workspace: requestId={}, workspaceId={}", requestId, workspaceId);
            throw new IllegalArgumentException("Join request does not belong to this workspace");
        }

        WorkspaceJoinRequest.JoinRequestStatus newStatus = WorkspaceJoinRequest.JoinRequestStatus.valueOf(updateRequest.getStatus());

        if (newStatus == WorkspaceJoinRequest.JoinRequestStatus.APPROVED) {
            WorkspaceMember newMember = WorkspaceMember.builder()
                    .workspaceId(workspaceId)
                    .userId(joinRequest.getUserId())
                    .role(WorkspaceMember.WorkspaceRole.MEMBER)
                    .isDefault(false)
                    .isActive(true)
                    .build();
            workspaceMemberRepository.save(newMember);
            log.info("User approved and added as member: workspaceId={}, userId={}", workspaceId, joinRequest.getUserId());
        }

        joinRequest.setStatus(newStatus);
        WorkspaceJoinRequest updated = workspaceJoinRequestRepository.save(joinRequest);

        User user = userRepository.findById(joinRequest.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
        
        UserProfile userProfile = userProfileRepository.findByUserId(joinRequest.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return convertToJoinRequestResponse(updated, user, userProfile);
    }

    /**
     * Workspace의 가입 신청 목록 조회 (OWNER/ADMIN만)
     */
    @Transactional(readOnly = true)
    public List<JoinRequestResponse> getJoinRequests(UUID workspaceId, UUID requesterId, String status) {
        log.debug("Fetching join requests: workspaceId={}, status={}", workspaceId, status);

        checkWorkspaceAdminOrOwner(workspaceId, requesterId);

        List<WorkspaceJoinRequest> requests;
        if (status != null && !status.isEmpty()) {
            requests = workspaceJoinRequestRepository.findByWorkspaceIdAndStatus(
                    workspaceId, 
                    WorkspaceJoinRequest.JoinRequestStatus.valueOf(status)
            );
        } else {
            requests = workspaceJoinRequestRepository.findByWorkspaceId(workspaceId);
        }

        return requests.stream()
                .map(req -> {
                    User user = userRepository.findById(req.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));
                    UserProfile userProfile = userProfileRepository.findByUserId(req.getUserId())
                            .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));
                    return convertToJoinRequestResponse(req, user, userProfile);
                })
                .collect(Collectors.toList());
    }

    // ============================================================================
    // 권한 확인 (Private Methods)
    // ============================================================================

    /**
     * Workspace 멤버 확인
     */
    private void checkWorkspaceMember(UUID workspaceId, UUID userId) {
        if (!workspaceMemberRepository.existsByWorkspaceIdAndUserId(workspaceId, userId)) {
            log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("User is not a member of this workspace");
        }
    }

    /**
     * Workspace OWNER 확인
     */
    private void checkWorkspaceOwner(UUID workspaceId, UUID userId) {
        WorkspaceMember member = workspaceMemberRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> {
                    log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
                    return new IllegalArgumentException("User is not a member of this workspace");
                });

        if (member.getRole() != WorkspaceMember.WorkspaceRole.OWNER) {
            log.warn("User is not OWNER of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Only workspace owner can perform this action");
        }
    }

    /**
     * Workspace OWNER 또는 ADMIN 확인
     */
    private void checkWorkspaceAdminOrOwner(UUID workspaceId, UUID userId) {
        WorkspaceMember member = workspaceMemberRepository.findByWorkspaceIdAndUserId(workspaceId, userId)
                .orElseThrow(() -> {
                    log.warn("User is not a member of workspace: workspaceId={}, userId={}", workspaceId, userId);
                    return new IllegalArgumentException("User is not a member of this workspace");
                });

        if (member.getRole() != WorkspaceMember.WorkspaceRole.OWNER && 
            member.getRole() != WorkspaceMember.WorkspaceRole.ADMIN) {
            log.warn("User is not OWNER or ADMIN of workspace: workspaceId={}, userId={}", workspaceId, userId);
            throw new IllegalArgumentException("Only workspace owner or admin can perform this action");
        }
    }

    // ============================================================================
    // DTO 변환 (Private Methods)
    // ============================================================================

    private WorkspaceResponse convertToWorkspaceResponse(Workspace workspace, User owner, UserProfile ownerProfile) {
        return WorkspaceResponse.builder()
                .id(workspace.getGroupId())
                .name(workspace.getName())
                .description(workspace.getCompanyName())
                .ownerId(owner.getUserId())
                .ownerName(ownerProfile.getName())
                .ownerEmail(owner.getEmail())
                .createdAt(workspace.getCreatedAt())
                .updatedAt(workspace.getUpdatedAt())
                .build();
    }

    private WorkspaceMemberResponse convertToWorkspaceMemberResponse(WorkspaceMember member) {
        User user = userRepository.findById(member.getUserId())
                .orElseThrow(() -> new UserNotFoundException("사용자를 찾을 수 없습니다."));

        UserProfile userProfile = userProfileRepository.findByUserId(member.getUserId())
                .orElseThrow(() -> new UserNotFoundException("프로필을 찾을 수 없습니다."));

        return WorkspaceMemberResponse.builder()
                .id(member.getWorkspaceMemberId())
                .workspaceId(member.getWorkspaceId())
                .userId(member.getUserId())
                .userName(userProfile.getName())
                .userEmail(user.getEmail())
                .roleName(member.getRole().name())
                .isDefault(member.getIsDefault())
                .joinedAt(member.getJoinedAt())
                .build();
    }

    private JoinRequestResponse convertToJoinRequestResponse(WorkspaceJoinRequest request, User user, UserProfile userProfile) {
        return JoinRequestResponse.builder()
                .id(request.getJoinRequestId())
                .workspaceId(request.getWorkspaceId())
                .userId(request.getUserId())
                .userName(userProfile.getName())
                .userEmail(user.getEmail())
                .status(request.getStatus().name())
                .requestedAt(request.getRequestedAt())
                .updatedAt(request.getUpdatedAt())
                .build();
    }
}