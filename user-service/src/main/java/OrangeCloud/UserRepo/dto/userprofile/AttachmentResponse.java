package OrangeCloud.UserRepo.dto.userprofile;

import OrangeCloud.UserRepo.entity.Attachment;
import lombok.Builder;
import lombok.Getter;

import java.time.LocalDateTime;
import java.util.UUID;

/**
 * 첨부파일 응답 DTO
 */
@Getter
@Builder
public class AttachmentResponse {
    private UUID attachmentId;
    private String entityType;
    private UUID entityId;
    private String status;
    private String fileName;
    private String fileUrl;
    private Long fileSize;
    private String contentType;
    private UUID uploadedBy;
    private LocalDateTime uploadedAt;
    private LocalDateTime expiresAt;

    /**
     * Entity를 DTO로 변환
     */
    public static AttachmentResponse from(Attachment attachment) {
        return AttachmentResponse.builder()
                .attachmentId(attachment.getId())
                .entityType(attachment.getEntityType().name())
                .entityId(attachment.getEntityId())
                .status(attachment.getStatus().name())
                .fileName(attachment.getFileName())
                .fileUrl(attachment.getFileUrl())
                .fileSize(attachment.getFileSize())
                .contentType(attachment.getContentType())
                .uploadedBy(attachment.getUploadedBy())
                .uploadedAt(attachment.getCreatedAt())
                .expiresAt(attachment.getExpiresAt())
                .build();
    }
}
