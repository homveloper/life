# 총기 발사 시스템 구현 워크플로우

## 프로젝트 개요
서버 중심 검증 방식 총기 발사 시스템을 LIFE 프로젝트의 기존 아키텍처에 통합하여 구현합니다.
기존의 Event-Driven CQRS + Clean Architecture + Auto-Router 패턴을 활용합니다.

## 기반 아키텍처
- **통신 프로토콜**: JSON-RPC 2.0 over HTTP REST
- **라우팅**: Auto-Router 패키지를 사용한 반사 기반 핸들러 등록
- **이벤트**: Redis Streams + Watermill을 통한 이벤트 스트리밍
- **캐싱**: Redis 기반 고속 상태 저장
- **태스크 큐**: Asynq를 통한 분산 태스크 처리

---

## Phase 1: Core Infrastructure 구현
*기본 인프라스트럭처 및 도메인 모델 구축*

### 1.1 도메인 모델 및 데이터 구조 정의 `[FIRE-001]` ✅
**의존성**: 없음 (독립 작업) | **병렬 처리**: ✅ Wave 1
- [x] 총알 도메인 모델 구현 (`internal/domain/bullet/`)
  - [x] `Bullet` aggregate 구조체 정의
  - [x] `Position`, `Direction`, `Velocity` value objects
  - [x] `WeaponType` enum 정의
  - [x] 총알 상태 머신 (Active, Expired, Hit) 구현
- [x] 발사 이벤트 정의 (`internal/domain/bullet/events.go`)
  - [x] `BulletFiredEvent`
  - [x] `BulletHitEvent`  
  - [x] `BulletExpiredEvent`
- [x] Repository 인터페이스 정의 (`internal/domain/bullet/repository.go`)
  - [x] `BulletRepository` 인터페이스
  - [x] Redis 기반 구현체 (`internal/domain/bullet/redis_repository.go`)

### 1.2 Redis 데이터 스키마 설계 `[FIRE-002]` ✅
**의존성**: 없음 (독립 작업) | **병렬 처리**: ✅ Wave 1
- [x] 플레이어 수치 스키마 정의
  - [x] `player:{id}:stats` - PlayerStats 구조체 (탄약, 마지막_발사_시간, 무기_타입, 연사_세션_시작_시간)
  - [x] `fire_session:{session_id}` - FireSession 구조체 (시작시간, 연사속도, 플레이어ID)
- [x] 총알 상태 스키마 정의
  - [x] `bullet:{id}` - Redis JSON으로 Bullet 구조체 저장 (FT.CREATE 검색 인덱스 포함)
  - [x] TTL 설정 (총알 만료 시간 기반 자동 만료)
- [x] 성능 최적화를 위한 Pipeline 전략 수립 (Repository 배치 연산으로 구현)

---

## Phase 2: API Layer 구현
*JSON-RPC 2.0 기반 총기 발사 API 구현*

### 2.1 Handler 구조체 및 메서드 정의 `[FIRE-003]`
**의존성**: FIRE-001, FIRE-002, FIRE-004 | **병렬 처리**: 🔵 Wave 2
- [ ] `BulletHandler` 구조체 생성 (`internal/api/bullet/`)
  - [ ] Auto-Router 패턴 준수
  - [ ] 의존성 주입 (Repository, EventBus)
- [ ] 핸들러 메서드 구현
  - [ ] `Fire(w http.ResponseWriter, r *http.Request)` - 단발 발사
  - [ ] `StartFire(w http.ResponseWriter, r *http.Request)` - 연사 시작  
  - [ ] `StopFire(w http.ResponseWriter, r *http.Request)` - 연사 중단

### 2.2 JSON-RPC 요청/응답 구조체 정의 `[FIRE-004]`
**의존성**: 없음 (독립 작업) | **병렬 처리**: ✅ Wave 1
- [ ] 요청 구조체 (`internal/api/bullet/types.go`)
  - [ ] `FireRequest` - 단발 발사 요청
  - [ ] `StartFireRequest` - 연사 시작 요청
  - [ ] `StopFireRequest` - 연사 중단 요청
- [ ] 응답 구조체
  - [ ] `FireResponse` - 발사 승인/거부 응답
  - [ ] `StartFireResponse` - 연사 세션 정보
  - [ ] JSON-RPC 2.0 표준 준수

### 2.3 검증 로직 구현 `[FIRE-005]`
**의존성**: FIRE-002 | **병렬 처리**: 🔵 Wave 2
- [ ] 발사 권한 검증 시스템
  - [ ] 탄약 보유량 확인
  - [ ] 무기 쿨다운 계산 (현재시간 - 마지막_발사_시간 >= 쿨다운_시간)
  - [ ] 연사 상태 계산 함수 (`isFiring()` - 연사_세션_시작_시간 기반)
  - [ ] 플레이어 생존 상태 검증
- [ ] Redis Pipeline을 통한 고속 검증 (3-8ms 목표)
- [ ] 검증 실패 시 적절한 에러 응답

---

## Phase 3: Event System 통합
*Watermill + Redis Streams 기반 이벤트 처리*

### 3.1 이벤트 구조체 정의 및 발송 `[FIRE-006]` ✅
**의존성**: FIRE-001 | **병렬 처리**: ✅ Wave 1
- [x] 이벤트 구조체 정의 (`internal/domain/bullet/events.go`)
  - [x] `BulletFiredEvent` - 총알 발사 이벤트
  - [x] `BulletExpiredEvent` - 총알 만료 이벤트
  - [x] `BulletHitEvent` - 총알 충돌 이벤트
- [x] Watermill EventBus 사용 준비 (shared.BaseEvent 패턴)
- [x] 이벤트 JSON 직렬화 구현

### 3.2 SSE 브로드캐스터 통합 `[FIRE-007]`
**의존성**: FIRE-006 | **병렬 처리**: 🔵 Wave 2
- [ ] 기존 SSE 시스템과 연동
- [ ] 실시간 이벤트 브로드캐스트
  - [ ] `bullet.spawned` - 모든 클라이언트에게 총알 생성 알림
  - [ ] `bullet.expired` - 총알 만료 알림
  - [ ] `player.fire.started` / `player.fire.stopped` - 연사 상태 알림

---

## Phase 4: Command/Query Handler 구현
*CQRS 패턴 기반 비즈니스 로직 처리*

### 4.1 Command Handler 구현 `[FIRE-008]`
**의존성**: FIRE-001, FIRE-002 | **병렬 처리**: 🔵 Wave 2
- [ ] `FireBulletCommandHandler` (`internal/application/commands/bullet/`)
  - [ ] 발사 권한 검증
  - [ ] 총알 생성 및 저장
  - [ ] Watermill EventBus로 직접 이벤트 발송
- [ ] `StartFireCommandHandler`
  - [ ] 연사 세션 생성
  - [ ] 연사 상태 관리
- [ ] `StopFireCommandHandler`
  - [ ] 연사 세션 종료
  - [ ] 리소스 정리

### 4.2 Query Handler 구현 `[FIRE-009]`
**의존성**: FIRE-001 | **병렬 처리**: 🔵 Wave 2
- [ ] `BulletQueryHandler` (`internal/application/queries/bullet/`)
  - [ ] 활성 총알 조회
  - [ ] 플레이어 발사 상태 조회
  - [ ] 총알 궤적 정보 조회

---

## Phase 5: Auto-Router 통합
*기존 라우팅 시스템에 새 핸들러 등록*

### 5.1 라우터 설정 업데이트 `[FIRE-010]`
**의존성**: FIRE-003, FIRE-008, FIRE-009 | **병렬 처리**: 🟠 Wave 3
- [ ] `cmd/server/main.go` 업데이트
  - [ ] `BulletHandler` 인스턴스 생성
  - [ ] Auto-Router에 핸들러 등록
  - [ ] URL 패턴: `/api/bullet/fire`, `/api/bullet/start-fire`, `/api/bullet/stop-fire`

### 5.2 의존성 주입 설정 `[FIRE-011]`
**의존성**: FIRE-006, FIRE-007 | **병렬 처리**: 🟠 Wave 3
- [ ] IoC 컨테이너 설정 업데이트
- [ ] Repository, Watermill EventBus, Cache 의존성 주입
- [ ] 상태 계산 서비스 (`PlayerStateCalculator`) 주입
- [ ] 환경 변수 기반 설정 관리

---

## Phase 6: 연사 시스템 (Asynq 기반)
*분산 태스크 큐를 통한 연사 제어*

### 6.1 연사 태스크 정의 `[FIRE-012]`
**의존성**: 없음 (독립 작업) | **병렬 처리**: ✅ Wave 1
- [ ] Asynq 태스크 타입 정의
  - [ ] `fire:continuous` - 연사 루프 태스크
  - [ ] `fire:stop` - 연사 중단 태스크
- [ ] 태스크 페이로드 구조체 정의

### 6.2 연사 제어 워커 구현 `[FIRE-013]`
**의존성**: FIRE-012 | **병렬 처리**: 🔵 Wave 2
- [ ] `ContinuousFireWorker` 구현
- [ ] 연사 속도 제어 (600 RPM = 100ms 간격)
- [ ] 탄약 소모 및 수치 업데이트 (마지막_발사_시간 갱신)
- [ ] 연사 중단 조건 계산 (탄약 부족, 세션 만료)

---

## Phase 7: Testing & Validation
*테스트 코드 작성 및 시스템 검증*

### 7.1 단위 테스트 `[FIRE-014]`
**의존성**: FIRE-010, FIRE-011, FIRE-013 | **병렬 처리**: 🔴 Wave 4
- [ ] 도메인 모델 테스트
- [ ] Handler 테스트 (HTTP 요청/응답)
- [ ] Command/Query Handler 테스트
- [ ] Repository 테스트

### 7.2 통합 테스트 `[FIRE-015]`
**의존성**: FIRE-014 | **병렬 처리**: ❌ 순차 처리
- [ ] API 엔드포인트 전체 플로우 테스트
- [ ] 이벤트 발행/구독 테스트  
- [ ] Redis 연동 테스트
- [ ] Asynq 태스크 실행 테스트

### 7.3 성능 테스트 `[FIRE-016]`
**의존성**: FIRE-015 | **병렬 처리**: ❌ 순차 처리
- [ ] 발사 응답 시간 측정 (목표: 3-8ms)
- [ ] 동시 발사 요청 부하 테스트
- [ ] 메모리 사용량 모니터링
- [ ] Redis 성능 최적화 검증

---

## Phase 8: Documentation & Deployment
*문서화 및 배포 준비*

### 8.1 API 문서 작성 `[FIRE-017]`
**의존성**: FIRE-016 | **병렬 처리**: 🔵 Wave 5
- [ ] Swagger 스키마 업데이트
- [ ] JSON-RPC 2.0 엔드포인트 문서화
- [ ] 요청/응답 예제 작성

### 8.2 환경 설정 및 배포 `[FIRE-018]`
**의존성**: FIRE-017 | **병렬 처리**: ❌ 순차 처리
- [ ] `.env.example` 업데이트 (총기 관련 설정)
- [ ] Docker 구성 업데이트 (필요 시)
- [ ] 프로덕션 배포 체크리스트

---

## TODO: 별도 워크플로우로 분리될 구현사항

### 충돌 처리 시스템 `[TODO: 별도 워크플로우]`
- **참조**: `docs/bullet-collision-architecture.md`
- **구현 범위**: 
  - 수학적 충돌 예측 엔진
  - Asynq 기반 충돌 이벤트 스케줄링  
  - 동물 AI와의 실시간 상호작용
  - 충돌 결과 브로드캐스팅

### 클라이언트 예측 시스템 `[TODO: 별도 워크플로우]`
- **구현 범위**:
  - JavaScript 기반 예측 총알 시스템
  - 서버 동기화 및 롤백 메커니즘
  - 시각/청각 피드백 시스템

---

## 성공 지표
- [ ] 단발 발사 API 응답시간 < 8ms
- [ ] 연사 시스템 정확한 속도 제어
- [ ] 이벤트 기반 실시간 브로드캐스팅
- [ ] 100% 서버 검증 보장
- [ ] 기존 아키텍처와 완벽한 통합