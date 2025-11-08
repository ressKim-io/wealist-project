package OrangeCloud.UserRepo.dto.userprofile;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Size;

@Schema(description = "사용자 프로필 업데이트 요청 DTO")
public record UpdateProfileRequest(
    @Schema(description = "업데이트할 사용자 이름", example = "새로운 이름", required = false)
    @Size(min = 2, max = 50, message = "이름은 2자 이상 50자 이하로 입력해야 합니다.")
    String name,

    @Schema(description = "업데이트할 이메일", example = "user@example.com", required = false)
    String email,

    @Schema(description = "업데이트할 프로필 이미지 URL", example = "https://new.image.url/avatar.jpg", required = false)
    String profileImageUrl
) {}
