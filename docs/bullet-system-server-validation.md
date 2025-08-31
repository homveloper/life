# 서버 중심 검증 방식 총기 발사 시스템 (2차 플랜)

## 개요

2D 게임 특성을 고려하여, 서버에서 모든 총기 발사와 충돌을 검증한 후 브로드캐스팅하는 보수적 방식입니다. 초고속 반응성보다는 **공정성과 안정성**을 우선합니다.

### 서버 검증 방식의 핵심 철학
- **"서버가 진실이다"**
- **"검증 먼저, 재미는 그 다음"**  
- **100% 서버 검증 + 클라이언트 예측**

## 시스템 구성요소

- **클라이언트**: 예측 기반 브라우저 게임 클라이언트 (JavaScript)
- **서버**: 권위적 게임 로직 처리 및 검증 서버 (Go)
- **검증 시스템**: 실시간 발사 권한 검증
- **충돌 처리**: 수학적 예측 기반 충돌 엔진 (bullet-collision-architecture.md)
- **이벤트 스케줄링**: Asynq 기반 정확한 시점 실행
- **SSE**: 서버 중재 브로드캐스트 통신

## 낙관적 예측 + Redis 최적화 총기 발사 시퀀스

```mermaid
sequenceDiagram
    participant C1 as 발사자 클라이언트
    participant C2 as 관찰자 클라이언트
    participant Bot as 게임내 봇 (AI)
    participant S as 서버 (HTTP REST)
    participant R as Redis 캐시
    participant SSE as SSE 브로드캐스터
    
    Note over C1: 🔫 사용자가 발사 키 누름 (MouseDown)
    
    C1->>C1: 0ms - 낙관적 예측 (98% 신뢰도)<br/>✅ 로컬 상태 체크 (탄약, 쿨다운)<br/>✅ 즉시 총구 화염 + 사운드<br/>✅ 예측 총알 생성 및 발사<br/>✅ 로컬 탄약 차감
    
    par 예측과 동시에 서버 요청
        C1->>S: HTTP POST /api/trainer/fire<br/>fire.request{direction, timestamp}
        
        Note over S: ⚡ Redis 초고속 검증 (3-8ms 목표)
        
        S->>R: 1ms - Pipeline 조회<br/>HMGET player:123 ammo, last_fire, cooldown
        
        R-->>S: 캐시된 플레이어 상태 반환<br/>{ammo: 29, last_fire: 1234567890}
        
        S->>S: 2ms - 간단 검증<br/>✅ 탄약 > 0?<br/>✅ 쿨다운 경과?<br/>✅ 기본 권한 체크
        
        alt 검증 성공
            Note over S: 98% 케이스 - 예측 성공
            S->>R: 3ms - Pipeline 상태 업데이트<br/>HINCRBY player:123 ammo -1<br/>HSET player:123 last_fire NOW<br/>HSET bullet:abc123 {...}<br/>EXPIRE bullet:abc123 10
            
            Note over S: 🎯 즉시 충돌 예측 및 스케줄링
            S->>S: 수학적 충돌 엔진 실행<br/>📐 모든 동물과 충돌 시점 계산<br/>⚡ Asynq 태스크 스케줄링
            
            S-->>C1: 5-8ms - HTTP 200 응답<br/>fire.approved{bullet_id, server_time}
            
            C1->>C1: 예측 확정<br/>✅ 예측 → 확정 전환<br/>✅ 서버 시간 동기화<br/>✅ 총알 ID 업데이트
            
            par SSE 브로드캐스트 (비동기)
                SSE-->>C2: bullet.spawned<br/>{bullet_id, owner, trajectory}
            end
            
            C2->>C2: 총알 렌더링<br/>✅ 궤적 생성<br/>✅ 관찰자 시점 총알 생성
            
        else 검증 실패
            Note over S: 2% 케이스 - 예측 실패
            S-->>C1: 8ms - HTTP 400 응답<br/>fire.rejected{reason: "no_ammo"}
            
            C1->>C1: 부드러운 롤백<br/>❌ 예측 총알 페이드아웃<br/>❌ 사운드 중단<br/>💡 실패 사유 표시 (탄약 없음)<br/>📚 실패 패턴 학습
        end
    end
    
    Note over R,SSE: 🎯 수학적 예측 기반 충돌 처리 (이벤트 기반)
    
    Note over S: 발사 승인 즉시 충돌 예측 수행
    S->>S: 수학적 충돌 계산<br/>📐 해석적 궤적-원 교점 계산<br/>⏰ 정확한 충돌 시점 예측<br/>🎯 가장 빠른 충돌 선택
    
    alt 충돌 예측 성공
        S->>S: Asynq 태스크 스케줄링<br/>collision:execute 태스크 생성<br/>정확한 충돌 시점에 실행 예약
        
        Note over S: 예측된 정확한 시점에 충돌 실행
        S->>Bot: 권위적 피해 적용<br/>실시간 위치 재검증<br/>봇 HP 차감 (-25)<br/>충돌 결과 저장
        
        Bot->>Bot: AI 즉시 반응<br/>✅ 피격 애니메이션<br/>✅ 회피/반격 행동<br/>✅ 상태 변화 (분노/도망)
        
        par SSE 즉시 브로드캐스트
            SSE-->>C1: animal.hit<br/>{target: bot_456, damage: 25, score_gain: +100}
            SSE-->>C2: animal.hit<br/>{shooter: player_123, victim: bot_456, damage: 25}
        end
        
        C1->>C1: 즉시 명중 피드백<br/>✅ 타격 마커<br/>✅ +100 점수<br/>✅ 킬 사운드
        
        C2->>C2: 실시간 시각 효과<br/>✅ 봇 피격 애니메이션<br/>✅ 혈액 파티클<br/>✅ 데미지 넘버<br/>✅ 킬피드 업데이트
        
    else 충돌 예측 없음
        S->>S: 총알 자연 만료 스케줄링<br/>bullet:expire 태스크 생성<br/>사정거리 도달 시점에 실행 예약
        
        Note over S: 사정거리 도달 시점에 자동 실행
        S->>R: DEL bullet:abc123
        SSE-->>C1: bullet.expired{bullet_id}
        SSE-->>C2: bullet.expired{bullet_id}
    end
    
    Note over C1,C2: 📊 혁신적 성능: 체감 0ms + 수학적 정확도 100% + CPU 97% 절약
```

## 낙관적 예측 연사 시스템 (이벤트 기반)

```mermaid
sequenceDiagram
    participant C1 as 발사자 클라이언트
    participant S as 서버 HTTP REST
    participant R as Redis 캐시
    participant SSE as SSE 브로드캐스터
    participant Others as 관찰자 클라이언트들
    
    Note over C1: 사용자가 연사 키 누름 MouseDown
    
    C1->>C1: 0ms - 연사 시작 예측<br/>로컬 연사 상태 ON<br/>즉시 총구 화염 시작<br/>연사 사운드 루프 시작<br/>첫 총알 즉시 발사
    
    C1->>S: HTTP POST /api/trainer/fire-start<br/>fire.start weapon assault_rifle
    
    S->>R: Redis 연사 권한 검증<br/>HMGET player:123 ammo weapon fire_mode
    
    alt 연사 승인 95퍼센트 케이스
        S->>R: 연사 상태 저장<br/>HSET player:123 firing true<br/>HSET fire_session:123 rate 600
        
        S-->>C1: 3-5ms - 연사 승인<br/>fire.start.approved session_id fire_rate
        
        C1->>C1: 연사 확정<br/>예측에서 확정 전환<br/>서버 연사 속도 동기화
        
        SSE-->>Others: player.fire.started<br/>shooter weapon fire_rate
        
        Others->>Others: 관찰자 시점 연사 시작<br/>발사자 총구 화염 이펙트<br/>연사 사운드 재생
        
        Note over S: 서버 제어 연사 루프 100ms = 600RPM
        
        loop 연사 중 서버 타이머 기반
            S->>R: 탄약 및 연사 상태 체크<br/>HMGET player:123 ammo firing
            
            alt 탄약 충분 그리고 연사 중
                S->>R: 총알 생성 + 탄약 차감<br/>HINCRBY player:123 ammo -1<br/>HSET bullet:xyz data
                
                Note over S: 각 총알에 대해 충돌 예측 실행
                S->>S: 수학적 충돌 엔진<br/>📐 연사 총알별 충돌 시점 계산<br/>⚡ Asynq 태스크 스케줄링
                
                SSE-->>C1: bullet.fired<br/>bullet_id confirmed
                SSE-->>Others: bullet.spawned<br/>bullet_id trajectory
                
                C1->>C1: 연사 총알 확정<br/>다음 총알 예측 생성
                Others->>Others: 관찰자 시점 총알 생성<br/>시각 효과 렌더링
            else 탄약 부족
                S->>R: 연사 강제 중단<br/>HSET player:123 firing false
                
                SSE-->>C1: fire.stopped<br/>reason no_ammo
                SSE-->>Others: player.fire.stopped<br/>shooter reason
                
                C1->>C1: 연사 중단<br/>연사 이펙트 페이드아웃<br/>탄약 없음 알림
                Others->>Others: 관찰자 시점 연사 중단
            end
        end
    else 연사 거부 5퍼센트 케이스
        S-->>C1: 연사 거부<br/>fire.start.rejected reason
        
        C1->>C1: 연사 예측 롤백<br/>연사 이펙트 중단<br/>거부 사유 표시
    end
    
    Note over C1: 사용자가 연사 키 뗌 MouseUp
    
    C1->>C1: 즉시 연사 중단<br/>연사 이펙트 중단<br/>사운드 페이드아웃
    
    C1->>S: HTTP POST /api/trainer/fire-stop<br/>fire.stop sesㄱsion_id
    
    S->>R: 연사 상태 정리<br/>HSET player:123 firing false<br/>DEL fire_session:123
    
    SSE-->>Others: player.fire.stopped<br/>shooter voluntary true
    
    Others->>Others: 관찰자 시점 연사 중단<br/>총구 화염 중단<br/>연사 사운드 중단
    
    Note over C1: 연사 성능 - 체감 즉시 반응 서버 동기화 3-8ms
```

## 낙관적 예측 + Redis 방식의 장단점

### ✅ 장점
1. **체감 즉시 반응**: 98% 예측 정확도로 0ms 체감 지연
2. **서버 검증 유지**: 권위적 검증으로 공정성 보장
3. **Redis 초고속**: 3-8ms 서버 응답으로 빠른 확정
4. **부드러운 실패 처리**: 2% 실패 시에도 자연스러운 롤백
5. **HTTP REST 활용**: 기존 인프라 그대로 사용
6. **학습형 시스템**: 실패 패턴 학습으로 예측 정확도 향상

### ⚠️ 고려사항  
1. **클라이언트 구현 복잡도**: 예측-확정 로직 및 롤백 시스템 구현 필요
2. **예측 실패 UX**: 2% 실패 케이스에 대한 자연스러운 사용자 피드백 설계
3. **메모리 관리**: 예측 상태 및 학습 데이터의 효율적 관리
4. **네트워크 품질 의존성**: 불안정한 네트워크에서 예측 정확도 저하 가능

## 데이터 구조

### 서버 총알 상태
```go
type ServerBullet struct {
    ID          string    `json:"id"`
    OwnerID     string    `json:"owner_id"`
    StartPos    Position  `json:"start_pos"`
    Direction   Direction `json:"direction"`
    Speed       float64   `json:"speed"`
    Damage      int       `json:"damage"`
    CreatedAt   time.Time `json:"created_at"`
    MaxDistance float64   `json:"max_distance"`
    
    // 서버 전용 상태
    CurrentPos  Position  `json:"current_pos"`
    IsActive    bool      `json:"is_active"`
    LastUpdate  time.Time `json:"last_update"`
}

// 총알 위치 업데이트 (서버에서만)
func (b *ServerBullet) UpdatePosition() {
    elapsed := time.Since(b.CreatedAt).Seconds()
    b.CurrentPos = Position{
        X: b.StartPos.X + (b.Direction.X * b.Speed * elapsed),
        Y: b.StartPos.Y + (b.Direction.Y * b.Speed * elapsed),
    }
    b.LastUpdate = time.Now()
}

// 충돌 검사 (서버 권위적 - 봇 대상)
func (b *ServerBullet) CheckCollision(bots []Bot) *HitResult {
    for _, bot := range bots {
        // 발사자가 플레이어이므로 모든 봇은 충돌 대상
        
        distance := math.Sqrt(
            math.Pow(b.CurrentPos.X - bot.Position.X, 2) +
            math.Pow(b.CurrentPos.Y - bot.Position.Y, 2)
        )
        
        if distance < bot.HitboxRadius {
            return &HitResult{
                VictimID: bot.ID,
                VictimType: "bot",
                Damage:   b.Damage,
                HitPos:   b.CurrentPos,
                IsKill:   (bot.HP - b.Damage) <= 0,
                ScoreGain: calculateBotScore(bot.Type, bot.Level),
            }
        }
    }
    return nil
}
```

### 클라이언트 예측 시스템
```javascript
// 클라이언트 예측 총알 (임시)
class PredictiveBullet {
    constructor(data) {
        this.id = `pred_${Date.now()}_${Math.random()}`;
        this.serverId = null; // 서버 확인 후 설정
        this.ownerID = data.ownerID;
        this.startPos = data.startPos;
        this.direction = data.direction;
        this.speed = data.speed || 25.0;
        this.damage = data.damage || 25;
        this.firedAt = performance.now();
        this.maxDistance = 100.0;
        this.isPrediction = true; // 예측 상태
        this.isConfirmed = false; // 서버 확인 여부
    }
    
    // 서버 확인 시 예측을 확정으로 전환
    confirmWithServer(serverData) {
        this.serverId = serverData.bullet_id;
        this.isConfirmed = true;
        this.isPrediction = false;
        
        // 서버 시간으로 동기화
        const serverTime = serverData.server_timestamp;
        const clientTime = performance.now();
        this.firedAt = clientTime - (Date.now() - serverTime);
    }
    
    // 예측 실패 시 총알 제거
    rejectPrediction() {
        this.isPrediction = false;
        this.isConfirmed = false;
        // UI에서 제거되어야 함
    }
}
```

### 서버 메시지 형식
```json
// 클라이언트 → 서버: 발사 요청
{
    "jsonrpc": "2.0",
    "method": "fire.request", 
    "params": {
        "weapon": "pistol",
        "direction": {"x": 1.0, "y": 0.0},
        "client_timestamp": 1756563570123.456
    },
    "id": 1
}

// 서버 → 클라이언트: 발사 승인
{
    "jsonrpc": "2.0",
    "method": "fire.approved",
    "params": {
        "bullet_id": "srv_bullet_001",
        "server_timestamp": 1756563570125.789,
        "trajectory": {
            "start_pos": {"x": 15.5, "y": 10.2},
            "direction": {"x": 1.0, "y": 0.0},
            "speed": 25.0,
            "damage": 25
        },
        "ammo_remaining": 23
    }
}

// 서버 → 모든 클라이언트: 봇 충돌 확인
{
    "jsonrpc": "2.0",
    "method": "bot.hit.confirmed",
    "params": {
        "bullet_id": "srv_bullet_001",
        "shooter_id": "player_123",
        "victim_id": "bot_456",
        "victim_type": "wolf",
        "victim_level": 3,
        "damage": 25,
        "hit_pos": {"x": 20.3, "y": 12.1},
        "victim_hp": 75,
        "is_kill": false,
        "score_gain": 100,
        "server_timestamp": 1756563570200.456
    }
}

// 클라이언트 → 서버: 연사 시작
{
    "jsonrpc": "2.0",
    "method": "fire.start",
    "params": {
        "weapon": "assault_rifle",
        "direction": {"x": 1.0, "y": 0.0}
    },
    "id": 2
}

// 서버 → 클라이언트: 연사 시작 승인
{
    "jsonrpc": "2.0",
    "method": "fire.start.approved",
    "params": {
        "fire_session_id": "session_001",
        "fire_rate": 600,
        "burst_mode": false
    }
}
```

## 성능 예측 (낙관적 예측 + Redis 방식)

| 지표 | 예상 성능 | 설명 |
|------|-----------|------|
| **발사 체감 반응시간** | **0ms** | 98% 예측 정확도로 즉시 반응 |
| **서버 응답시간** | **3-8ms** | Redis Pipeline + HTTP Keep-Alive |
| **예측 정확도** | **98%+** | 학습형 로컬 상태 기반 예측 |
| **충돌 정확도** | **100%** | 수학적 해석해로 완벽한 정확도 |
| **치팅 가능성** | **거의 0%** | 서버 최종 검증 + 실시간 재검증 |
| **서버 CPU 사용률** | **매우 낮음** | 수학적 예측으로 97% CPU 절약 |
| **네트워크 사용량** | **중간** | HTTP + SSE 효율적 활용 |
| **동시 접속자 수** | **100-200명** | Redis 성능 + 예측 시스템으로 확장성 향상 |

## 구현 단계

### Phase 1: 기본 서버 검증 시스템
- [ ] 서버 총기 발사 검증 로직
- [ ] 클라이언트 예측 시스템
- [ ] 기본 충돌 감지
- [ ] 승인/거부 메시지 처리

### Phase 2: 연사 시스템
- [ ] 이벤트 기반 연사 시작/중단
- [ ] 서버 측 연사 속도 제어
- [ ] 탄약 관리 시스템
- [ ] 무기별 연사 특성

### Phase 3: 수학적 충돌 시스템 통합
- [ ] 수학적 충돌 엔진 구현 (bullet-collision-architecture.md 참조)
- [ ] Asynq 기반 이벤트 스케줄링 시스템
- [ ] 동물 이동 시 실시간 재계산 로직
- [ ] 서버 검증과 충돌 예측의 완벽한 동기화

### Phase 4: 안정성
- [ ] 네트워크 끊김 처리
- [ ] 예측 실패 보정
- [ ] 서버 장애 복구
- [ ] 클라이언트 재동기화

**목표**: 수학적 예측 기반 혁신적 성능 + 완벽한 공정성 🎯🚀  
**상세**: bullet-collision-architecture.md 문서 참조

## 클라이언트 예측 시스템 상태 흐름

### 예측 총알 상태 머신

```mermaid
stateDiagram-v2
    [*] --> Idle : 게임 시작
    
    Idle --> Predicting : 사용자 발사 입력
    
    state Predicting {
        [*] --> LocalValidation
        LocalValidation --> RenderingPrediction : 로컬 검증 통과
        LocalValidation --> [*] : 검증 실패 (탄약/쿨다운)
        
        RenderingPrediction --> WaitingServerResponse : 서버 요청 전송
        
        state WaitingServerResponse {
            [*] --> PendingResponse
            PendingResponse --> ServerApproved : HTTP 200
            PendingResponse --> ServerRejected : HTTP 400/403
            PendingResponse --> ServerTimeout : 타임아웃 (>1초)
        }
    }
    
    Predicting --> Confirmed : server_approved
    Predicting --> RolledBack : server_rejected/timeout
    
    state Confirmed {
        [*] --> SyncWithServer
        SyncWithServer --> ActiveBullet : 서버 ID 할당
        
        ActiveBullet --> HitTarget : 충돌 감지
        ActiveBullet --> ExpiredRange : 사정거리 초과
        ActiveBullet --> ExpiredTime : 시간 초과
    }
    
    state RolledBack {
        [*] --> FadeOutEffect
        FadeOutEffect --> ShowErrorMessage
        ShowErrorMessage --> LearnPattern : 실패 패턴 학습
        LearnPattern --> [*]
    }
    
    Confirmed --> [*] : 총알 소멸
    RolledBack --> [*] : 롤백 완료
    
    note right of Predicting
        isPrediction: true
        isConfirmed: false
        렌더링 진행 중
        서버 응답 대기
    end note
    
    note right of Confirmed
        isPrediction: false
        isConfirmed: true
        serverId 할당됨
        충돌 감지 활성화
    end note
    
    note right of RolledBack
        부드러운 시각 효과
        사용자 피드백 제공
        예측 정확도 개선
    end note
```

### 예측 시스템 처리 플로우

```mermaid
flowchart TD
    A[🎮 사용자 발사 입력] --> B{로컬 상태 검증}
    
    B -->|✅ 통과| C[예측 총알 생성]
    B -->|❌ 실패| D[🚫 발사 거부]
    
    C --> E[즉시 시각/청각 피드백]
    C --> F[서버 검증 요청]
    
    E --> G[예측 렌더링 시작]
    F --> H{서버 응답}
    
    H -->|승인 98%| I[✅ 예측 확정]
    H -->|거부 2%| J[❌ 예측 롤백]
    H -->|타임아웃| K[⏰ 타임아웃 처리]
    
    I --> L[서버 ID 할당]
    I --> M[시간 동기화]
    
    J --> N[페이드아웃 효과]
    J --> O[에러 메시지 표시]
    J --> P[실패 패턴 학습]
    
    K --> Q[네트워크 재시도]
    K --> R[로컬 상태로 복귀]
    
    L --> S[충돌 감지 활성화]
    M --> S
    
    S --> T{충돌 발생?}
    T -->|적중| U[🎯 타격 확인]
    T -->|사정거리 초과| V[📏 총알 소멸]
    T -->|시간 초과| W[⏰ 자동 제거]
    
    N --> X[예측 시스템 학습]
    O --> X
    P --> X
    
    style C fill:#e3f2fd
    style I fill:#c8e6c9  
    style J fill:#ffcdd2
    style K fill:#fff3e0
    
    classDef prediction fill:#e1f5fe,stroke:#0277bd
    classDef confirmed fill:#e8f5e8,stroke:#388e3c
    classDef failed fill:#fce4ec,stroke:#d32f2f
    classDef timeout fill:#fff8e1,stroke:#f57c00
    
    class C,G prediction
    class I,L,M,S confirmed
    class J,N,O,P failed
    class K,Q,R timeout
```
