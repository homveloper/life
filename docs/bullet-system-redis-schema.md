# 총기 발사 시스템 Redis 데이터 스키마

## 개요
LIFE 프로젝트의 총기 발사 시스템에서 사용되는 Redis 데이터 구조 및 스키마 정의입니다.

## 1. 총알 데이터 (Bullets)

### 기본 키 구조
```
bullet:{bullet_id}
```

### 데이터 필드
```
HSET bullet:{bullet_id}
  data:       "{json_serialized_bullet_data}"     # 전체 총알 객체 JSON
  player_id:  "{player_id}"                       # 발사한 플레이어 ID
  state:      "active|hit|expired"                # 총알 상태
  fired_at:   {unix_timestamp}                    # 발사 시각
  expires_at: {unix_timestamp}                    # 만료 시각
  pos_x:      {float}                             # 현재 X 좌표
  pos_y:      {float}                             # 현재 Y 좌표
```

### TTL 설정
- 자동 만료: `expires_at` 기준으로 계산된 TTL
- 최소 TTL: 10초 (만료된 총알도 짧은 시간 보관)

## 2. 플레이어 발사 통계 (Player Stats)

### 키 구조
```
player:{player_id}:stats
```

### 데이터 필드
```
HSET player:{player_id}:stats
  player_id:            "{player_id}"             # 플레이어 ID
  ammo_count:           {integer}                 # 현재 탄약 수량
  weapon_type:          "{weapon_type}"           # 무기 타입
  last_fire_time:       {unix_timestamp}         # 마지막 발사 시간
  fire_session_started: {unix_timestamp}         # 연사 세션 시작 시간 (옵셔널)
  updated_at:           {unix_timestamp}         # 마지막 업데이트 시간
```

### 검증 로직에 사용되는 계산
- **쿨다운 검증**: `current_time - last_fire_time >= weapon_cooldown_ms`
- **연사 상태**: `fire_session_started != null`
- **탄약 검증**: `ammo_count > 0`

## 3. 연사 세션 (Fire Sessions)

### 키 구조
```
fire_session:{session_id}
```

### 데이터 구조 (JSON)
```json
{
  "session_id": "uuid",
  "player_id": "player_id",
  "weapon_type": "assault_rifle",
  "fire_rate": 600,
  "started_at": "timestamp",
  "last_fired_at": "timestamp",
  "bullets_shot": 15,
  "is_active": true
}
```

### 플레이어 활성 세션 인덱스
```
player:{player_id}:active_session → {session_id}
```

### TTL 설정
- 세션 TTL: 1시간 (자동 정리)

## 4. 인덱스 구조

### 상태별 인덱스
```
idx:bullet:state:active      # 활성 총알 목록
idx:bullet:state:hit         # 명중한 총알 목록  
idx:bullet:state:expired     # 만료된 총알 목록
```

### 플레이어별 인덱스
```
idx:bullet:player:{player_id}  # 특정 플레이어의 총알 목록
```

### 위치별 인덱스 (공간 검색용)
```
idx:bullet:position:{x}:{y}    # 정수 좌표 기준 총알 인덱스
```

## 5. 성능 최적화 전략

### Redis Pipeline 사용
모든 쓰기 작업에서 Pipeline 사용으로 네트워크 라운드트립 최소화:

```go
pipe := redis.Pipeline()
pipe.HMSet(ctx, bulletKey, fields)
pipe.Expire(ctx, bulletKey, ttl)
pipe.SAdd(ctx, stateIndex, bulletID)
pipe.SAdd(ctx, playerIndex, bulletID)
_, err := pipe.Exec(ctx)
```

### TTL 기반 자동 정리
- 총알: `expires_at` 기준 자동 만료
- 세션: 1시간 TTL
- 인덱스: 총알과 동일한 TTL 설정

### 메모리 최적화
- JSON 직렬화로 복잡한 객체 저장
- 인덱스는 Set 자료구조로 중복 제거
- 만료된 데이터 자동 삭제

## 6. 예시 데이터

### 활성 총알 예시
```
HSET bullet:uuid-1234
  data: '{"id":"uuid-1234","player_id":"player-456","weapon_type":"assault_rifle",...}'
  player_id: "player-456"
  state: "active"
  fired_at: 1674123456
  expires_at: 1674123466
  pos_x: 245.7
  pos_y: 182.3

EXPIRE bullet:uuid-1234 10
```

### 플레이어 통계 예시
```
HSET player:456:stats
  player_id: "player-456"
  ammo_count: 28
  weapon_type: "assault_rifle"
  last_fire_time: 1674123456
  fire_session_started: 1674123450
  updated_at: 1674123456
```

### 인덱스 예시
```
SADD idx:bullet:state:active "uuid-1234" "uuid-5678" "uuid-9012"
SADD idx:bullet:player:player-456 "uuid-1234" "uuid-5678"
SADD idx:bullet:position:245:182 "uuid-1234"
```

## 7. 성능 목표

### 응답 시간 목표
- 단발 발사 API: **< 8ms**
- 검증 로직 (Pipeline): **3-8ms**
- 총알 상태 조회: **< 5ms**

### 동시성 처리
- Redis Watch/Multi/Exec를 통한 낙관적 락
- Pipeline을 통한 배치 처리
- TTL 기반 자동 리소스 정리

## 8. 운영 고려사항

### 모니터링 지표
- 활성 총알 수량: `SCARD idx:bullet:state:active`
- 플레이어별 총알 수: `SCARD idx:bullet:player:{id}`
- 메모리 사용량: 총알 데이터 + 인덱스 크기

### 정리 작업
- 만료된 인덱스 정리 (TTL로 대부분 자동화)
- 비활성 플레이어 통계 정리 (옵셔널)
- 성능 로그 및 메트릭 수집