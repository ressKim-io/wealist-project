package OrangeCloud.UserRepo.service;

import OrangeCloud.UserRepo.dto.auth.AuthResponse;
import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.UserRepository;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.util.JwtTokenProvider;
import OrangeCloud.UserRepo.exception.UserNotFoundException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.util.Date;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Transactional
@Slf4j
public class AuthService {

    private final UserRepository userRepository;
    private final UserProfileRepository userProfileRepository;
    private final JwtTokenProvider tokenProvider;
    private final RedisTemplate<String, Object> redisTemplate;
    // private final WorkspaceService workspaceService;

    // ============================================================================
    // 로그아웃
    // ============================================================================

    /**
     * 로그아웃 - 토큰을 Redis 블랙리스트에 추가
     */
    public void logout(String token) {
        log.debug("Attempting to log out token");

        tokenProvider.validateToken(token);

        Date expirationDate = tokenProvider.getExpirationDateFromToken(token);
        long ttl = expirationDate.getTime() - System.currentTimeMillis();

        if (ttl > 0) {
            redisTemplate.opsForValue().set(token, "blacklisted", Duration.ofMillis(ttl));
            log.debug("Token blacklisted successfully in Redis with TTL: {}ms", ttl);
        } else {
            log.warn("Token is already expired. Not adding to blacklist");
        }
    }

    // ============================================================================
    // 토큰 갱신
    // ============================================================================

    /**
     * Refresh Token을 사용하여 새로운 Access Token 발급
     */
    public AuthResponse refreshToken(String refreshToken) {
        log.debug("Attempting to refresh token");

        tokenProvider.validateToken(refreshToken);

        if (isTokenBlacklisted(refreshToken)) {
            log.warn("Refresh token is blacklisted");
            throw new OrangeCloud.UserRepo.exception.CustomJwtException(
                    OrangeCloud.UserRepo.exception.ErrorCode.TOKEN_BLACKLISTED);
        }

        UUID userId = tokenProvider.getUserIdFromToken(refreshToken);
        log.debug("Extracted user ID from refresh token: {}", userId);

        User user = userRepository.findById(userId)
                .orElseThrow(() -> {
                    log.warn("User not found for ID: {}", userId);
                    return new UserNotFoundException("사용자를 찾을 수 없습니다.");
                });

        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> {
                    log.warn("Profile not found for user: {}", userId);
                    return new UserNotFoundException("프로필을 찾을 수 없습니다.");
                });

        // 기존 refresh token 블랙리스트 추가
        Date expirationDate = tokenProvider.getExpirationDateFromToken(refreshToken);
        long ttl = expirationDate.getTime() - System.currentTimeMillis();
        if (ttl > 0) {
            redisTemplate.opsForValue().set(refreshToken, "blacklisted", Duration.ofMillis(ttl));
            log.debug("Old refresh token blacklisted with TTL: {}ms", ttl);
        }

        // 새로운 토큰 생성
        String newAccessToken = tokenProvider.generateToken(user.getUserId());
        String newRefreshToken = tokenProvider.generateRefreshToken(user.getUserId());

        return new AuthResponse(
                newAccessToken,
                newRefreshToken,
                user.getUserId(),
                profile.getNickName(),
                user.getEmail());
    }

    // ============================================================================
    // 토큰 유효성 검증 (외부 서비스용)
    // ============================================================================

    /**
     * Access Token의 유효성을 검증하고 사용자 ID를 반환합니다.
     */
    public UUID validateTokenAndGetUserId(String token) {
        log.debug("Validating token for external service use.");

        // 1. 토큰 유효성 검사 (서명, 만료 시간 확인)
        // 토큰이 유효하지 않으면 이 시점에서 CustomJwtException이 throw됩니다.
        tokenProvider.validateToken(token);

        // 2. 토큰이 블랙리스트에 있는지 확인 (로그아웃된 토큰인지 확인)
        if (isTokenBlacklisted(token)) {
            log.warn("Attempted to use a blacklisted token.");
            throw new OrangeCloud.UserRepo.exception.CustomJwtException(
                    OrangeCloud.UserRepo.exception.ErrorCode.TOKEN_BLACKLISTED);
        }

        // 3. 토큰에서 User ID 추출
        UUID userId = tokenProvider.getUserIdFromToken(token);
        log.info("Token validated successfully, user ID: {}", userId);

        return userId;
    }

    // ============================================================================
    // 사용자 정보 조회
    // ============================================================================

    /**
     * 현재 인증된 사용자 정보 조회
     */
    public User getCurrentUser(UUID userId) {
        log.debug("Fetching user for ID: {}", userId);

        User user = userRepository.findById(userId)
                .orElseThrow(() -> {
                    log.warn("User not found for ID: {}", userId);
                    return new UserNotFoundException("사용자를 찾을 수 없습니다.");
                });

        log.debug("Successfully retrieved user for ID: {}", userId);
        return user;
    }

    // ============================================================================
    // 토큰 블랙리스트 확인
    // ============================================================================

    /**
     * 토큰이 Redis 블랙리스트에 있는지 확인
     */
    public boolean isTokenBlacklisted(String token) {
        log.debug("Checking if token is blacklisted");
        Boolean isBlacklisted = redisTemplate.hasKey(token);
        return Boolean.TRUE.equals(isBlacklisted);
    }
}