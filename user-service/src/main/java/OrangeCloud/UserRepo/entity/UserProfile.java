package OrangeCloud.UserRepo.entity;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;
import java.time.LocalDateTime;
import java.util.UUID;


@Entity
@Table(name = "userProfile")
@Getter
@NoArgsConstructor(access = AccessLevel.PROTECTED)
@AllArgsConstructor
@Builder
public class UserProfile {

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    @Column(name = "profileId", columnDefinition = "UUID")
    private UUID profileId;

    @Column(name = "userId", columnDefinition = "UUID", nullable = false, unique = true)
    private UUID userId;

    @Column(name = "nickName", length = 50)
    private String nickName;

    @Column(name = "email", length = 100)
    private String email; // null 허용

    @Column(name = "profileImageUrl")
    private String profileImageUrl; // null 허용 (기본 이미지 사용 가능)

    @CreationTimestamp
    @Column(name = "createdAt", updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updatedAt")
    private LocalDateTime updatedAt;

    // =========================================================================
    // 💡 업데이트 로직 (Service에서 호출)
    // =========================================================================

    public void updateNickName(String nickName) {
        this.nickName = nickName;
    }

    public void updateEmail(String email) {
        this.email = email;
    }

    public void updateProfileImageUrl(String profileImageUrl) {
        this.profileImageUrl = profileImageUrl;
    }
}