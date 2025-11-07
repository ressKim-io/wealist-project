package OrangeCloud.UserRepo.entity;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "groups")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@ToString
@EqualsAndHashCode(of = "groupId")
public class Workspace {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    @Column(name = "group_id", updatable = false, nullable = false, columnDefinition = "UUID")
    private UUID groupId;

    @Column(nullable = false)
    private String name;

    @Column(name = "company_name")
    private String companyName; // 실제로는 description 역할

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    // 소프트 삭제를 위한 필드들
    @Column(name = "is_active", nullable = false)
    @Builder.Default
    private Boolean isActive = true;

    @Column(name = "deleted_at")
    private LocalDateTime deletedAt;

    // 수동으로 groupId 설정할 수 있는 생성자 추가
    public Workspace(UUID groupId, String name, String companyName) {
        this.groupId = groupId;
        this.name = name;
        this.companyName = companyName;
        this.isActive = true;
    }

    public Workspace(String name, String companyName) {
        this.name = name;
        this.companyName = companyName;
        this.isActive = true;
    }

    public void softDelete() {
        this.isActive = false;
        this.deletedAt = LocalDateTime.now();
    }

    public void reactivate() {
        this.isActive = true;
        this.deletedAt = null;
    }
}