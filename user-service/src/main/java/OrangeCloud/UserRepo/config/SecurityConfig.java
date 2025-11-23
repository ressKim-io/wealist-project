package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.filter.JwtAuthenticationFilter;
import OrangeCloud.UserRepo.filter.JwtExceptionFilter;
import OrangeCloud.UserRepo.oauth.CustomOAuth2UserService;
import OrangeCloud.UserRepo.oauth.OAuth2SuccessHandler;
import OrangeCloud.UserRepo.service.AuthService;
import OrangeCloud.UserRepo.util.JwtTokenProvider;
import com.fasterxml.jackson.databind.ObjectMapper; // 💡 ObjectMapper 임포트 추가
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

import java.util.Arrays;

@Configuration
@EnableWebSecurity
public class SecurityConfig {

        @Autowired(required = false)
        private CustomOAuth2UserService customOAuth2UserService;
        
        @Autowired(required = false)
        private OAuth2SuccessHandler oAuth2SuccessHandler;
        
        private final ObjectMapper objectMapper; // 💡 JacksonConfig에서 설정된 ObjectMapper 빈 주입

        public SecurityConfig(ObjectMapper objectMapper) {
                this.objectMapper = objectMapper;
        }

        @Bean
        public SecurityFilterChain filterChain(
                        HttpSecurity http,
                        JwtTokenProvider jwtTokenProvider,
                        AuthService authService) throws Exception {
                // JWT 필터 생성
                JwtAuthenticationFilter jwtAuthenticationFilter = new JwtAuthenticationFilter(jwtTokenProvider,
                                authService);

                // 💡 수정: 주입받은 ObjectMapper를 JwtExceptionFilter에 전달
                JwtExceptionFilter jwtExceptionFilter = new JwtExceptionFilter(objectMapper);

                http
                        .csrf(csrf -> csrf.disable())
                        .cors(cors -> cors.configurationSource(corsConfigurationSource()))
                        .sessionManagement(session -> session
                                        .sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                        .authorizeHttpRequests(authz -> authz
                                        // Swagger UI 경로 허용
                                        .requestMatchers("/swagger-ui/**").permitAll()
                                        .requestMatchers("/swagger-ui.html").permitAll()
                                        .requestMatchers("/v3/api-docs/**").permitAll()
                                        .requestMatchers("/swagger-resources/**").permitAll()
                                        // 인증 API 허용 (회원가입, 로그인)
                                        .requestMatchers("/api/auth/signup").permitAll()
                                        .requestMatchers("/api/auth/login").permitAll()
                                        .requestMatchers("/api/auth/refresh").permitAll()
                                        // OAuth2 로그인 경로 허용
                                        .requestMatchers("/login/oauth2/**").permitAll()
                                        .requestMatchers("/oauth2/**").permitAll()
                                        // 테스트 엔드포인트 허용
                                        .requestMatchers("/test").permitAll()
                                        .requestMatchers("/error").permitAll()
                                        .requestMatchers("/").permitAll()
                                        .requestMatchers("/actuator/health").permitAll()
                                        // ************ 나중에 아래 전체 허용 해제 필수 **********
                                        .requestMatchers("/**").permitAll()
                                        // 나머지는 인증 필요
                                        .anyRequest().authenticated())
                        // 💡 JWT 예외 처리 필터 추가 (인증 필터보다 먼저)
                        .addFilterBefore(jwtExceptionFilter, UsernamePasswordAuthenticationFilter.class)
                        // 💡 JWT 인증 필터 추가 (ExceptionFilter 뒤, 인증 실패 시 ExceptionFilter가 잡도록)
                        .addFilterBefore(jwtAuthenticationFilter, JwtExceptionFilter.class)
                        .headers(headers -> headers
                                        .frameOptions(frame -> frame.sameOrigin()));

                // OAuth2 로그인 설정 추가 (OAuth2 설정이 있을 때만)
                if (customOAuth2UserService != null && oAuth2SuccessHandler != null) {
                        http.oauth2Login(oauth2 -> oauth2
                                        .userInfoEndpoint(userInfo -> userInfo
                                                        .userService(customOAuth2UserService))
                                        // 🔑 OAuth2 Endpoint 명시적 추가
                                        .authorizationEndpoint(
                                                        endpoint -> endpoint.baseUri("/oauth2/authorization"))
                                        .redirectionEndpoint(
                                                        endpoint -> endpoint.baseUri("/login/oauth2/code/*"))
                                        .successHandler(oAuth2SuccessHandler));
                }

                return http.build();
        }

        @Bean
        public CorsConfigurationSource corsConfigurationSource() {
                CorsConfiguration configuration = new CorsConfiguration();

                // 허용할 Origin 설정 - 개발 환경에서는 모든 Origin 허용
                configuration.setAllowedOriginPatterns(Arrays.asList("*"));

                // 또는 특정 Origin만 허용하려면 아래 주석 해제
                // configuration.setAllowedOriginPatterns(Arrays.asList(
                // "https://*.cloudfront.net", // CloudFront 도메인
                // "https://wealist.co.kr", // 프로덕션 도메인
                // "http://localhost:5173", // Vite 개발 서버
                // "http://localhost:3000", // 기존 개발 환경
                // "http://localhost:8000" // 기존 개발 환경
                // ));

                // 허용할 HTTP 메서드
                configuration.setAllowedMethods(Arrays.asList(
                                "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"));

                // 허용할 헤더
                configuration.setAllowedHeaders(Arrays.asList("*"));

                // 노출할 헤더 (클라이언트에서 접근 가능한 헤더)
                configuration.setExposedHeaders(Arrays.asList(
                                "Authorization", "Content-Type", "X-Requested-With"));

                // 인증 정보 포함 허용
                configuration.setAllowCredentials(true);

                // preflight 요청 캐시 시간 (초)
                configuration.setMaxAge(3600L);

                UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
                source.registerCorsConfiguration("/**", configuration);
                return source;
        }
}