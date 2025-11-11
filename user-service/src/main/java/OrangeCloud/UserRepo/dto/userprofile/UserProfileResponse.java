package OrangeCloud.UserRepo.dto.userprofile;

import OrangeCloud.UserRepo.entity.UserProfile;
import io.swagger.v3.oas.annotations.media.Schema;
import lombok.*; // * 사용 시 모든 어노테이션 임포트

import java.util.UUID;

@Getter 
@Builder
@NoArgsConstructor // Lombok이 필요로 하는 기본 생성자
@AllArgsConstructor // Lombok이 빌더 생성을 위해 사용하는 모든 필드 생성자
@Schema(description = "사용자 프로필 응답 DTO")
public class UserProfileResponse { 
    
    // 💡 필드 선언부가 명확해야 합니다.
    private UUID profileId;
    private UUID userId;
    private String nickName;
    private String email;
    private String profileImageUrl;
    // 💡 정적 팩토리 메서드는 그대로 유지
    public static UserProfileResponse from(UserProfile profile) {
        return UserProfileResponse.builder()
                .profileId(profile.getProfileId()) 
                .userId(profile.getUserId())
                .nickName(profile.getNickName())
                .email(profile.getEmail())
                .profileImageUrl(profile.getProfileImageUrl())
                .build();
    }
}