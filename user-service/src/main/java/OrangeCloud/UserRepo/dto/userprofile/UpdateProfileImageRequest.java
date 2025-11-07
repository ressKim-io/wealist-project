package OrangeCloud.UserRepo.dto.userprofile;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;

@Getter
@NoArgsConstructor
@AllArgsConstructor
public class UpdateProfileImageRequest {
    @NotBlank(message = "프로필 이미지 URL은 필수입니다.")
    private String profileImageUrl;
}