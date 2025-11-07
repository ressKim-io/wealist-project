package OrangeCloud.UserRepo.entity;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "workspace_members")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@ToString
@EqualsAndHashCode(of = "workspaceMemberId")
public class WorkspaceMember {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "workspace_member_id", updatable = false, nullable = false, columnDefinition = "UUID")
    private UUID workspaceMemberId;

    @Column(name = "workspace_id", nullable = false, columnDefinition = "UUID")
    private UUID workspaceId;

    @Column(name = "user_id", nullable = false, columnDefinition = "UUID")
    private UUID userId;

    @Column(name = "role_name", nullable = false)
    @Enumerated(EnumType.STRING)
    private WorkspaceRole role;

    @Column(name = "is_default", nullable = false)
    @Builder.Default
    private Boolean isDefault = false;

    @CreationTimestamp
    @Column(name = "joined_at", updatable = false)
    private LocalDateTime joinedAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @Column(name = "is_active", nullable = false)
    @Builder.Default
    private Boolean isActive = true;

    public enum WorkspaceRole {
        OWNER,
        ADMIN,
        MEMBER
    }
}