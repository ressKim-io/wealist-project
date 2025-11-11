package OrangeCloud.UserRepo.config;

import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.springframework.boot.autoconfigure.jackson.Jackson2ObjectMapperBuilderCustomizer; // 👈 이 임포트 추가
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;



@Configuration
public class JacksonConfig {
    private static final Logger log = LoggerFactory.getLogger(JacksonConfig.class);
    
    public JacksonConfig() {
        log.warn("🚨 JacksonConfig가 성공적으로 로드되었습니다! 설정이 적용될 것입니다."); // 👈 이 줄 추가
    }
    // 💡 기존의 @Primary가 붙은 objectMapper() 메서드는 주석 처리하거나 제거합니다.
    
    /**
     * Spring Boot의 기본 Jackson ObjectMapper 빌더를 커스터마이징합니다.
     * 이 방법은 Spring Boot의 자동 설정과 충돌 없이 설정을 주입하는 가장 확실한 방법입니다.
     */
    @Bean
    public Jackson2ObjectMapperBuilderCustomizer jsonCustomizer() {
        
        log.info("✅ JavaTimeModule이 Jackson 빌더에 등록되었습니다."); // 👈 이 줄 추가
        return builder -> {
            // 💡 핵심 1: JavaTimeModule을 빌더에 등록하여 LocalDateTime 문제를 해결합니다.
            builder.modules(new JavaTimeModule());
            
            // 💡 핵심 2: Date를 타임스탬프(숫자) 대신 문자열로 직렬화하도록 설정합니다.
            // (이전 코드의 SerializationFeature.WRITE_DATES_AS_TIMESTAMPS, false와 동일)
            builder.featuresToDisable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        };
    }
}