package OrangeCloud.UserRepo.controller;

import OrangeCloud.UserRepo.dto.*;
import OrangeCloud.UserRepo.dto.auth.AuthResponse;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.service.AuthService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import io.swagger.v3.oas.annotations.Operation;
import OrangeCloud.UserRepo.exception.InvalidTokenException;
import OrangeCloud.UserRepo.exception.ErrorCode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@RestController
@RequestMapping("/api/auth")
@CrossOrigin(origins = "*", maxAge = 3600)
@Tag(name = "Authentication", description = "인증 관련 API")
public class AuthController {

    private static final Logger logger = LoggerFactory.getLogger(AuthController.class);

    private final AuthService authService;

    @Autowired
    public AuthController(AuthService authService) {
        this.authService = authService;
    }

    /**
     * 로그아웃
     * POST /api/auth/logout
     */
    @PostMapping("/logout")
    @Operation(summary = "로그아웃", description = "현재 세션을 종료하고 토큰을 무효화합니다.")
    public ResponseEntity<MessageApiResponse> logout(HttpServletRequest request) {
        logger.debug("Received logout request.");
        String token = extractTokenFromRequest(request);
        authService.logout(token);
        logger.info("로그아웃 성공");
        return ResponseEntity.ok(new MessageApiResponse(true, "로그아웃 성공"));
    }

    /**
     * 토큰 갱신
     * POST /api/auth/refresh
     */
    @PostMapping("/refresh")
    @Operation(summary = "토큰 갱신", description = "Refresh Token을 사용하여 Access Token을 갱신합니다.")
    public ResponseEntity<AuthResponse> refresh(@Valid @RequestBody RefreshTokenRequest refreshRequest) {
        logger.debug("Received refresh token request.");
        AuthResponse authResponse = authService.refreshToken(refreshRequest.getRefreshToken());
        logger.info("토큰 갱신 성공");
        return ResponseEntity.ok(authResponse);
    }

    /**
     * 현재 인증된 사용자 정보 조회
     * GET /api/auth/me
     */
    @GetMapping("/me")
    @Operation(summary = "내 정보 조회", description = "현재 인증된 사용자의 정보를 조회합니다.")
    public ResponseEntity<User> getCurrentUser(Authentication authentication) {
        UUID userId = UUID.fromString(authentication.getName());
        logger.debug("Received request to get current user info for ID: {}", userId);
        User user = authService.getCurrentUser(userId);
        logger.info("사용자 정보 조회 성공: {}", userId);
        return ResponseEntity.ok(user);
    }

    /**
     * 테스트용 Google OAuth 사용자 생성
     * GET /api/auth/test
     */
    @GetMapping("/test")
    @Operation(summary = "테스트 로그인", description = "테스트용 Google OAuth 사용자를 생성하고 토큰을 발급합니다.")
    public ResponseEntity<AuthResponse> testUser() {
        logger.debug("Test login request");
        return ResponseEntity.ok(authService.TestLogin());
    }

    /**
     * Request에서 Bearer 토큰 추출
     */
    private String extractTokenFromRequest(HttpServletRequest request) {
        logger.debug("Attempting to extract token from request.");
        String bearerToken = request.getHeader("Authorization");
        if (bearerToken != null && bearerToken.startsWith("Bearer ")) {
            String token = bearerToken.substring(7);
            logger.debug("Token extracted successfully.");
            return token;
        }
        logger.warn("Authorization 헤더에서 토큰을 찾을 수 없습니다.");
        throw new InvalidTokenException("Authorization 헤더에서 토큰을 찾을 수 없습니다.");
    }
}