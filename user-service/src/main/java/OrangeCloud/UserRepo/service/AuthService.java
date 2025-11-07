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
import java.time.LocalDateTime;
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

    // ============================================================================
    // 테스트용 로그인
    // ============================================================================

    /**
     * 테스트용 Google OAuth 사용자 생성 및 로그인
     */
    public AuthResponse TestLogin() {
        log.debug("Creating test Google OAuth user");

        String testEmail = "test_" + System.currentTimeMillis() + "@gmail.com";
        String testGoogleId = "test_google_id_" + System.currentTimeMillis();
        String testName = "Test User " + System.currentTimeMillis();

        // User 생성
        User testUser = User.builder()
                .email(testEmail)
                .googleId(testGoogleId)
                .provider("google")
                .isActive(true)
                .build();

        User savedUser = userRepository.save(testUser);
        log.debug("Created test Google user with ID: {}", savedUser.getUserId());

        // UserProfile 생성
        UserProfile profile = UserProfile.builder()
                .userId(savedUser.getUserId())
                .name(testName)
                .build();

        userProfileRepository.save(profile);
        log.debug("Created profile for test user: {}", savedUser.getUserId());

        // JWT 토큰 생성
        String accessToken = tokenProvider.generateToken(savedUser.getUserId());
        String refreshToken = tokenProvider.generateRefreshToken(savedUser.getUserId());

        return new AuthResponse(
                accessToken,
                refreshToken,
                savedUser.getUserId(),
                testName,
                savedUser.getEmail()
        );
    }

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
                    OrangeCloud.UserRepo.exception.ErrorCode.TOKEN_BLACKLISTED
            );
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
                profile.getName(),
                user.getEmail()
        );
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