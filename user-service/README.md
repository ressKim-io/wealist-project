# Docker 환경일 경우

# 1. 컨테이너 중지 및 제거

docker stop wealist-user-service
docker rm wealist-user-service

# 2. 이미지 재빌드 (Gradle build 후 jar 파일 생성 포함)

./gradlew clean build

# 3. Docker Compose (또는 run)으로 다시 실행

docker-compose up -d --build
