# ⚠️ 프론트엔드 보안 경고

## 🔴 중요: VITE_GOOGLE_CLIENT_SECRET 사용 금지

### 문제점

원본 저장소(OrangesCloud/wealist-project)의 환경변수 파일에 `VITE_GOOGLE_CLIENT_SECRET`이 정의되어 있습니다.

**이것은 심각한 보안 취약점입니다!**

### 왜 문제인가?

1. **VITE_ 접두사의 의미**
   - Vite 빌드 시스템은 `VITE_` 접두사가 붙은 환경변수를 **클라이언트 사이드 코드에 포함**시킵니다
   - 이는 빌드된 JavaScript 파일에 평문으로 노출됩니다

2. **CLIENT_SECRET의 의미**
   - OAuth Client Secret은 **서버에서만** 사용해야 하는 비밀 키입니다
   - 클라이언트(브라우저)에 노출되면 누구나 이 값을 탈취할 수 있습니다

3. **실제 위험**
   ```javascript
   // ❌ 잘못된 예시 - 브라우저에서 실행되는 코드
   const clientSecret = import.meta.env.VITE_GOOGLE_CLIENT_SECRET;
   // → 빌드된 JS 파일에 평문으로 포함됨
   // → 개발자 도구로 누구나 확인 가능
   // → 악의적 사용자가 귀하의 Google OAuth 앱을 악용 가능
   ```

### 올바른 구현 방법

#### ✅ 방법 1: 서버에서만 CLIENT_SECRET 사용 (현재 구현)

```yaml
# user-service/src/main/resources/application.yml
spring:
  security:
    oauth2:
      client:
        registration:
          google:
            client-id: ${GOOGLE_CLIENT_ID}      # ✅ 공개 가능
            client-secret: ${GOOGLE_CLIENT_SECRET}  # ✅ 서버에서만 사용
```

**환경변수:**
```bash
# ✅ 서버용 - .env 파일 (VITE_ 접두사 없음)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret  # 서버에서만 사용
```

**프론트엔드:**
```typescript
// ✅ 프론트엔드에서는 CLIENT_ID만 사용
const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID;
// CLIENT_SECRET은 절대 프론트엔드에서 사용하지 않음!
```

#### ✅ 방법 2: Authorization Code Flow 사용 (OAuth 표준)

1. **프론트엔드**: Client ID만으로 Google 로그인 시작
2. **Google**: 사용자 인증 후 Authorization Code 반환
3. **백엔드**: Authorization Code + Client Secret으로 Access Token 요청
4. **프론트엔드**: 백엔드에서 받은 토큰 사용

```typescript
// ✅ 프론트엔드 - 로그인 시작
window.location.href = `https://accounts.google.com/o/oauth2/v2/auth?` +
  `client_id=${VITE_GOOGLE_CLIENT_ID}&` +
  `redirect_uri=${REDIRECT_URI}&` +
  `response_type=code&` +
  `scope=email profile`;

// ❌ CLIENT_SECRET은 여기서 절대 사용하지 않음!
```

```java
// ✅ 백엔드 - Access Token 요청 (CLIENT_SECRET 사용)
@PostMapping("/oauth/callback")
public TokenResponse handleCallback(@RequestParam String code) {
    // Google OAuth Token Endpoint 호출
    // CLIENT_SECRET은 서버에서만 사용됨
}
```

### 현재 프로젝트 상태

✅ **현재 구현은 안전합니다**
- User Service(백엔드)에서만 `GOOGLE_CLIENT_SECRET` 사용
- 프론트엔드는 `VITE_GOOGLE_CLIENT_ID`만 사용
- `VITE_GOOGLE_CLIENT_SECRET`는 사용되지 않음

### 행동 지침

1. **절대 하지 말 것**
   - ❌ `VITE_GOOGLE_CLIENT_SECRET` 환경변수 생성 금지
   - ❌ 프론트엔드 코드에서 Client Secret 사용 금지
   - ❌ Client Secret을 빌드된 JavaScript에 포함시키기 금지

2. **반드시 할 것**
   - ✅ Client Secret은 서버 환경변수로만 관리
   - ✅ 프론트엔드는 Client ID만 사용
   - ✅ OAuth 토큰 교환은 항상 백엔드에서 처리

3. **코드 리뷰 체크리스트**
   - [ ] `VITE_` 접두사가 붙은 민감 정보가 있는가?
   - [ ] 프론트엔드에서 `CLIENT_SECRET`, `API_KEY`, `PRIVATE_KEY` 등을 사용하는가?
   - [ ] OAuth 토큰 교환이 클라이언트 사이드에서 이루어지는가?

### 참고 자료

- [Google OAuth 2.0 문서](https://developers.google.com/identity/protocols/oauth2)
- [OWASP - OAuth 2.0 Security](https://cheatsheetseries.owasp.org/cheatsheets/OAuth2_Cheat_Sheet.html)
- [Vite 환경변수 가이드](https://vitejs.dev/guide/env-and-mode.html)

---

## 📝 버전 히스토리

- **v1.0** (2025-01-11): 초기 작성
  - VITE_GOOGLE_CLIENT_SECRET 보안 경고 추가
  - 올바른 OAuth 구현 방법 문서화

---

## 💬 문의

보안 관련 질문이나 우려사항이 있다면 보안 팀에 문의하세요.
