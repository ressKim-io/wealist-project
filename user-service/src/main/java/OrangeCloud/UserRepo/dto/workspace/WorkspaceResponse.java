package OrangeCloud.UserRepo.dto.workspace;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.UUID;

@Getter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class WorkspaceResponse {
    private UUID id;
    private String name;
    private String description;
    private UUID ownerId;
    private String ownerName;
    private String ownerEmail;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}