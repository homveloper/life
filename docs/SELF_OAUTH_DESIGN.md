# 🏗️ Self-hosted OAuth Authorization Server

## 아키텍처 개요

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Game Client   │    │   Game Server   │    │   Auth Server   │
│                 │    │                 │    │                 │
│ - 게임 로직     │    │ - JWT 검증      │    │ - 사용자 관리   │
│ - 로그인 UI     │    │ - 게임 API      │    │ - OAuth 토큰    │
│                 │    │                 │    │ - 인증 페이지   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## OAuth Server 엔드포인트

### 인증 관련
- `GET /oauth/authorize` - OAuth 인증 시작점
- `POST /oauth/token` - 인증 코드를 액세스 토큰으로 교환
- `GET /oauth/userinfo` - 액세스 토큰으로 사용자 정보 조회
- `POST /oauth/introspect` - 토큰 유효성 검사

### 사용자 관리
- `GET /auth/login` - 로그인 페이지
- `POST /auth/login` - 로그인 처리
- `GET /auth/register` - 회원가입 페이지  
- `POST /auth/register` - 회원가입 처리
- `GET /auth/consent` - 권한 동의 페이지
- `POST /auth/consent` - 권한 동의 처리

## 데이터 모델

### User (사용자)
```go
type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"` // bcrypt 해시
    CreatedAt time.Time `json:"created_at"`
}
```

### OAuthClient (등록된 앱)
```go
type OAuthClient struct {
    ClientID     string   `json:"client_id"`
    ClientSecret string   `json:"client_secret"`
    RedirectURIs []string `json:"redirect_uris"`
    Scopes       []string `json:"scopes"`
    Name         string   `json:"name"`
}
```

### AuthorizationCode (인증 코드)
```go
type AuthorizationCode struct {
    Code        string    `json:"code"`
    UserID      string    `json:"user_id"`
    ClientID    string    `json:"client_id"`
    RedirectURI string    `json:"redirect_uri"`
    Scopes      []string  `json:"scopes"`
    ExpiresAt   time.Time `json:"expires_at"`
}
```

### AccessToken (액세스 토큰)
```go
type AccessToken struct {
    Token     string    `json:"token"`
    UserID    string    `json:"user_id"`
    ClientID  string    `json:"client_id"`
    Scopes    []string  `json:"scopes"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

## 플로우 상세

### 1. 인증 요청
```
GET /oauth/authorize?
  response_type=code&
  client_id=game_client&
  redirect_uri=http://game.com/callback&
  scope=profile email&
  state=random_state
```

### 2. 사용자 로그인 (우리 서버에서)
```html
<!-- 로그인 페이지 -->
<form action="/auth/login" method="POST">
  <input name="username" type="text" placeholder="아이디">
  <input name="password" type="password" placeholder="비밀번호">
  <input name="redirect" type="hidden" value="/oauth/authorize?...">
  <button type="submit">로그인</button>
</form>
```

### 3. 권한 동의
```html
<!-- 동의 페이지 -->
<form action="/auth/consent" method="POST">
  <p>LIFE 게임이 다음 권한을 요청합니다:</p>
  <ul>
    <li>프로필 정보 (profile)</li>
    <li>이메일 주소 (email)</li>
  </ul>
  <button name="allow" value="true">허용</button>
  <button name="allow" value="false">거부</button>
</form>
```

### 4. 코드 발급 및 리다이렉트
```
HTTP/1.1 302 Found
Location: http://game.com/callback?code=ABC123&state=random_state
```

### 5. 토큰 교환
```json
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=ABC123&
redirect_uri=http://game.com/callback&
client_id=game_client&
client_secret=game_secret
```

### 6. 사용자 정보 조회
```json
GET /oauth/userinfo
Authorization: Bearer ACCESS_TOKEN_XYZ

Response:
{
  "sub": "user_12345",
  "username": "player123",
  "email": "player@example.com",
  "profile": "https://example.com/profiles/player123"
}
```

## 보안 고려사항

### 1. PKCE (Proof Key for Code Exchange)
```go
// 클라이언트에서 생성
codeVerifier := generateRandomString(128)
codeChallenge := base64url(sha256(codeVerifier))

// authorize 요청에 추가
"code_challenge=" + codeChallenge + "&code_challenge_method=S256"

// token 요청에 추가  
"code_verifier=" + codeVerifier
```

### 2. State 파라미터 검증
```go
func validateState(receivedState, expectedState string) bool {
    return receivedState == expectedState
}
```

### 3. 토큰 만료 및 리프레시
```go
type RefreshToken struct {
    Token     string    `json:"token"`
    UserID    string    `json:"user_id"`
    ClientID  string    `json:"client_id"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

## 구현 단계

### Phase 1: 기본 사용자 관리
1. User 도메인 구현
2. 회원가입/로그인 페이지
3. 세션 관리

### Phase 2: OAuth 서버
1. OAuthClient 등록 시스템
2. Authorization Code 플로우
3. 토큰 발급/관리

### Phase 3: 보안 강화
1. PKCE 구현
2. Rate limiting
3. 토큰 리프레시

### Phase 4: 고급 기능
1. 다중 클라이언트 지원
2. 스코프 권한 관리
3. 감사 로그

이렇게 구현하면 완전히 독립적인 OAuth 시스템을 구축할 수 있습니다!