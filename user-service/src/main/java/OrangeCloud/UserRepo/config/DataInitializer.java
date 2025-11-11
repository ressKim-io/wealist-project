package OrangeCloud.UserRepo.config;

import OrangeCloud.UserRepo.entity.User;
import OrangeCloud.UserRepo.entity.UserProfile;
import OrangeCloud.UserRepo.repository.UserProfileRepository;
import OrangeCloud.UserRepo.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.Environment;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * 애플리케이션 시작 시 더미 데이터를 자동으로 생성합니다.
 * - 개발 환경에서만 실행됩니다 (dev, local 프로파일)
 * - 이미 데이터가 있으면 건너뜁니다
 */
@Slf4j
@Configuration
@RequiredArgsConstructor
public class DataInitializer {

    private final UserRepository userRepository;
    private final UserProfileRepository userProfileRepository;
    private final Environment environment;

    @Bean
    public CommandLineRunner initializeData() {
        return args -> {
            // 개발 환경에서만 실행
            String[] activeProfiles = environment.getActiveProfiles();
            boolean isDevelopment = Arrays.asList(activeProfiles).contains("dev")
                                 || Arrays.asList(activeProfiles).contains("local")
                                 || activeProfiles.length == 0; // 프로파일이 없으면 기본값으로 실행

            if (!isDevelopment) {
                log.info("⏭️  Production environment detected. Skipping dummy data initialization.");
                return;
            }

            // 이미 사용자가 있으면 건너뛰기
            long userCount = userRepository.count();
            if (userCount >= 10) {
                log.info("✅ Database already has {} users. Skipping initialization.", userCount);
                return;
            }

            log.info("🚀 Starting dummy data initialization...");

            // 더미 사용자 데이터
            List<DummyUserData> dummyUsers = createDummyUserData();

            List<User> createdUsers = new ArrayList<>();
            List<UserProfile> createdProfiles = new ArrayList<>();

            for (DummyUserData data : dummyUsers) {
                // 이메일 중복 체크
                if (userRepository.existsByEmailAndIsActiveTrue(data.email)) {
                    log.debug("⏭️  User already exists: {}", data.email);
                    continue;
                }

                // User 생성
                User user = User.builder()
                        .email(data.email)
                        .provider("google")
                        .googleId(data.googleId)
                        .isActive(true)
                        .build();

                User savedUser = userRepository.save(user);
                createdUsers.add(savedUser);

                // UserProfile 생성
                UserProfile profile = UserProfile.builder()
                        .userId(savedUser.getUserId())
                        .nickName(data.nickName)
                        .email(data.email)
                        .profileImageUrl(data.profileImageUrl)
                        .build();

                UserProfile savedProfile = userProfileRepository.save(profile);
                createdProfiles.add(savedProfile);

                log.info("✅ Created user: {} ({})", data.nickName, data.email);
            }

            log.info("🎉 Data initialization completed! Created {} users and {} profiles.",
                    createdUsers.size(), createdProfiles.size());
        };
    }

    /**
     * 더미 사용자 데이터 생성
     */
    private List<DummyUserData> createDummyUserData() {
        List<DummyUserData> users = new ArrayList<>();

        users.add(new DummyUserData(
                "(테스터)김철수",
                "chulsoo.kim@example.com",
                "google-id-001",
                "https://i.pravatar.cc/150?img=1"
        ));

        users.add(new DummyUserData(
                "(테스터)이영희",
                "younghee.lee@example.com",
                "google-id-002",
                "https://i.pravatar.cc/150?img=2"
        ));

        users.add(new DummyUserData(
                "(테스터)박민수",
                "minsu.park@example.com",
                "google-id-003",
                "https://i.pravatar.cc/150?img=3"
        ));

        users.add(new DummyUserData(
                "(테스터)정수진",
                "sujin.jung@example.com",
                "google-id-004",
                "https://i.pravatar.cc/150?img=4"
        ));

        users.add(new DummyUserData(
                "(테스터)최동욱",
                "dongwook.choi@example.com",
                "google-id-005",
                "https://i.pravatar.cc/150?img=5"
        ));

        users.add(new DummyUserData(
                "(테스터)한지민",
                "jimin.han@example.com",
                "google-id-006",
                "https://i.pravatar.cc/150?img=6"
        ));

        users.add(new DummyUserData(
                "(테스터)강태영",
                "taeyoung.kang@example.com",
                "google-id-007",
                "https://i.pravatar.cc/150?img=7"
        ));

        users.add(new DummyUserData(
                "(테스터)윤서연",
                "seoyeon.yoon@example.com",
                "google-id-008",
                "https://i.pravatar.cc/150?img=8"
        ));

        users.add(new DummyUserData(
                "(테스터)임준호",
                "junho.lim@example.com",
                "google-id-009",
                "https://i.pravatar.cc/150?img=9"
        ));

        users.add(new DummyUserData(
                "(테스터)송혜교",
                "hyekyo.song@example.com",
                "google-id-010",
                "https://i.pravatar.cc/150?img=10"
        ));

        return users;
    }

    /**
     * 더미 사용자 데이터를 담는 내부 클래스
     */
    private static class DummyUserData {
        String nickName;
        String email;
        String googleId;
        String profileImageUrl;

        public DummyUserData(String nickName, String email, String googleId, String profileImageUrl) {
            this.nickName = nickName;
            this.email = email;
            this.googleId = googleId;
            this.profileImageUrl = profileImageUrl;
        }
    }
}
