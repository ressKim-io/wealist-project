package OrangeCloud.UserRepo.oauth;

import OrangeCloud.UserRepo.util.JwtTokenProvider;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.core.Authentication;
import org.springframework.security.web.authentication.SimpleUrlAuthenticationSuccessHandler;
import org.springframework.stereotype.Component;
import org.springframework.web.util.UriComponentsBuilder;

import java.io.IOException;

@Slf4j
@Component
@RequiredArgsConstructor
public class OAuth2SuccessHandler extends SimpleUrlAuthenticationSuccessHandler {

    private final JwtTokenProvider jwtTokenProvider;

    @Value("${oauth2.redirect-url:http://localhost:3000/oauth/callback}")
    private String redirectUrl;

    @Override
    public void onAuthenticationSuccess(
            HttpServletRequest request,
            HttpServletResponse response,
            Authentication authentication
    ) throws IOException {
        CustomOAuth2User oAuth2User = (CustomOAuth2User) authentication.getPrincipal();

        log.info("OAuth2 login successful: email={}, userId={}", oAuth2User.getEmail(), oAuth2User.getUserId());

        // JWT 토큰 생성
        String accessToken = jwtTokenProvider.generateToken(oAuth2User.getUserId());
        String refreshToken = jwtTokenProvider.generateRefreshToken(oAuth2User.getUserId());

        log.debug("Tokens generated: accessToken length={}, refreshToken length={}", 
                accessToken.length(), refreshToken.length());

        // 프론트엔드로 리다이렉트 (쿼리 파라미터로 토큰 전달)
        String targetUrl = UriComponentsBuilder.fromUriString(redirectUrl)
                .queryParam("accessToken", accessToken)
                .queryParam("refreshToken", refreshToken)
                .queryParam("userId", oAuth2User.getUserId().toString())
                .queryParam("email", oAuth2User.getEmail())
                .queryParam("name", oAuth2User.getName())
                .build()
                .toUriString();

        log.info("Redirecting to: {}", targetUrl);
        getRedirectStrategy().sendRedirect(request, response, targetUrl);
    }
}