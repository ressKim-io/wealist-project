package OrangeCloud.UserRepo.dto.userprofile;

import io.swagger.v3.oas.annotations.media.Schema;

@Schema(description = "프로필 이미지 업데이트 요청 DTO (기존 API용)")
public record UpdateProfileImageRequest(
    String profileImageUrl
) {}
