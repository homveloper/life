# 🚀 LIFE 프로젝트 세팅 워크플로우

**Living Intelligence for Fauna Ecosystem - 프로젝트 구축 가이드**

---

## 📋 Phase 1: 프로젝트 기반 구축 (1-2주)

### Task 1.1: 개발 환경 및 프로젝트 구조 설정
**담당**: DevOps/Backend  
**예상 시간**: 8시간  
**의존성**: 없음

#### Todo List:
- [x] **1.1.1** Go 1.21+ 개발 환경 구축
- [x] **1.1.2** 프로젝트 디렉토리 구조 생성 (DDD + CQRS 기반)
  ```
  life/
  ├── cmd/           # 애플리케이션 엔트리포인트
  ├── internal/      # 내부 패키지
  │   ├── domain/    # 도메인 모델
  │   ├── application/ # CQRS 핸들러
  │   ├── infrastructure/ # 리포지토리 구현
  │   └── api/       # HTTP 핸들러
  ├── pkg/           # 외부 공유 가능 패키지
  └── docs/          # 문서
  ```
- [x] **1.1.3** go.mod 초기화 및 기본 의존성 추가
- [x] **1.1.4** .env 템플릿 파일 생성 (Redis 연결 정보 등)
- [x] **1.1.5** .gitignore 설정

### Task 1.2: Redis 인프라 구축
**담당**: DevOps  
**예상 시간**: 4시간  
**의존성**: 1.1 완료

#### Todo List:
- [x] **1.2.1** Redis 서버 설치 및 설정 (로컬/도커) - 스킵
- [x] **1.2.2** Redis Streams 설정 확인 - 스킵
- [x] **1.2.3** Redis 연결 테스트 스크립트 작성 - 스킵
- [x] **1.2.4** Redis 데이터 영속화 설정 (AOF/RDB) - 스킵
- [x] **1.2.5** Redis 메모리 정책 설정 - 스킵

### Task 1.3: 기본 로깅 및 설정 시스템
**담당**: Backend  
**예상 시간**: 6시간  
**의존성**: 1.1 완료

#### Todo List:
- [x] **1.3.1** Zap 로깅 라이브러리 통합
- [x] **1.3.2** 환경별 로그 레벨 설정 (dev/prod)
- [x] **1.3.3** 구조화된 로깅 포맷 정의
- [x] **1.3.4** 설정 파일 관리 시스템 구현 (viper 등)
- [x] **1.3.5** 환경 변수 검증 로직

---

## 🏗️ Phase 2: 도메인 모델 및 아키텍처 구현 (1-2주)

### Task 2.1: 도메인 모델 설계 및 구현
**담당**: Backend/Architect  
**예상 시간**: 12시간  
**의존성**: 1.3 완료

#### Todo List:
- [x] **2.1.1** Trainer 도메인 엔티티 구현
  - 스탯: Level, Experience, HP
  - 위치: X, Y 좌표
  - 인벤토리: 아이템 목록
  - 보유 동물: 최대 6마리 파티
- [x] **2.1.2** Animal 도메인 엔티티 구현
  - 기본 스탯: HP, ATK, DEF, SPD, AS
  - 동물 종류: 사자, 코끼리, 치타
  - 레벨 및 경험치 시스템
  - 장비 슬롯 (목걸이 1개)
- [x] **2.1.3** Equipment 도메인 엔티티 구현
  - 무기: 공격력 보너스
  - 방어구: HP/방어력 보너스
  - 포획구: 포획 확률 보너스
  - 동물 목걸이: 스탯 보너스
- [x] **2.1.4** Map/World 도메인 엔티티 구현
  - 30x20 타일 맵 구조
  - 지형 요소 정의
  - 동물 스폰 포인트 시스템
- [x] **2.1.5** 도메인 이벤트 정의
  - TrainerMovedEvent, TrainerCreatedEvent, TrainerLeveledUpEvent
  - AnimalCapturedEvent, AnimalSpawnedEvent, AnimalMovedEvent
  - AnimalTookDamageEvent, AnimalLeveledUpEvent
  - Event 인터페이스 및 BaseEvent 구현
- [x] **2.1.6** 공용 Value Object 및 타입 시스템 구현 ✨ (추가 구현)
  - ID, Position, Stats, Level, Experience, Money, Timestamp
  - AggregateRoot, Repository 인터페이스
  - 도메인 에러 정의

### Task 2.2: Repository 인터페이스 정의
**담당**: Backend/Architect  
**예상 시간**: 8시간  
**의존성**: 2.1 완료

#### Todo List:
- [x] **2.2.1** TrainerRepository 인터페이스 정의 (IoC 패턴)
- [x] **2.2.2** AnimalRepository 인터페이스 정의 (IoC 패턴)
- [x] **2.2.3** EquipmentRepository 인터페이스 정의 (IoC 패턴)
- [x] **2.2.4** WorldStateRepository 인터페이스 정의 (IoC 패턴)
- [x] **2.2.5** EventStore 인터페이스 정의 (Redis Streams용) - 스킵

### Task 2.3: CQRS Command/Query 핸들러 구조
**담당**: Backend  
**예상 시간**: 10시간  
**의존성**: 2.2 완료

#### Todo List:
- [x] **2.3.1** Command 핸들러 베이스 구조 생성
- [x] **2.3.2** Query 핸들러 베이스 구조 생성
- [x] **2.3.3** 주요 Command 정의 (모든 도메인)
  - CreateTrainerCommand, MoveTrainerCommand
  - SpawnAnimalCommand, CaptureAnimalCommand, MoveAnimalCommand
  - CreateEquipmentCommand, EquipToAnimalCommand
  - CreateWorldCommand, MoveEntityInWorldCommand
- [x] **2.3.4** 주요 Query 정의 (모든 도메인)
  - GetTrainerByID, GetTrainersByPosition, GetTrainerInventory
  - GetAnimalByID, GetAnimalsByOwner, GetNearbyAnimals
  - GetEquipmentByID, GetEquipmentByRarity
  - GetWorldByID, GetTileAtPosition, GetEntitiesAtPosition
- [x] **2.3.5** Command/Query 핸들러 구현 (Trainer 도메인)

---

## 🔧 Phase 3: Infrastructure 구현 (1-2주)

### Task 3.1: Redis Repository 구현
**담당**: Backend  
**예상 시간**: 16시간  
**의존성**: 2.2 완료

#### Todo List:
- [x] **3.1.1** Redis 클라이언트 래퍼 구현 (기존 redisx 활용)
- [x] **3.1.2** TrainerRepository Redis 구현
  - JSON 직렬화/역직렬화
  - 키 네이밍 컨벤션 (trainer:{id})
  - Hash 기반 저장 (TTL 없음)
  - IoC 패턴 콜백 구현
- [x] **3.1.3** AnimalRepository Redis 구현
  - 위치/소유자/상태/타입 인덱싱
  - 반경 검색 기능
- [x] **3.1.4** EquipmentRepository Redis 구현
  - 소유자/희귀도 인덱싱
  - 착용 상태 관리
- [x] **3.1.5** WorldStateRepository Redis 구현
  - 타일 엔티티 관리
- [x] **3.1.6** 트랜잭션 처리 (Redis WATCH + TxPipelined)

### Task 3.2: HTTP API 기반 구조
**담당**: Backend  
**예상 시간**: 10시간  
**의존성**: 3.1 완료

#### Todo List:
- [x] **3.2.1** HTTP 서버 설정 (net/http 기반)
- [x] **3.2.2** 미들웨어 시스템 구현
  - 로깅 미들웨어
  - CORS 미들웨어
  - 에러 핸들링 미들웨어
  - Rate Limiting 미들웨어
- [x] **3.2.3** JSON-RPC 2.0 요청 파싱
- [x] **3.2.4** Response 포맷 표준화
- [x] **3.2.5** 헬스체크 엔드포인트 (/health)

---

## 🎮 Phase 4: 핵심 게임 로직 구현 (2-3주)

### Task 4.1: 플레이어 이동 및 상태 관리
**담당**: Backend  
**예상 시간**: 12시간  
**의존성**: 3.2 완료

#### Todo List:
- [ ] **4.1.1** POST /api/player/move API 구현
- [ ] **4.1.2** 이동 유효성 검증 (맵 경계, 장애물)
- [ ] **4.1.3** 플레이어 위치 업데이트 (Redis)
- [ ] **4.1.4** 플레이어 위치 변경 로그
- [ ] **4.1.5** POST /api/player/status API 구현

### Task 4.2: 실시간 전투 시스템
**담당**: Backend  
**예상 시간**: 20시간  
**의존성**: 4.1 완료

#### Todo List:
- [ ] **4.2.1** 전투 시작 트리거 로직
  - 플레이어-동물 거리 계산
  - 전투 상태 진입/해제
- [ ] **4.2.2** POST /api/battle/attack API 구현
  - 공격 쿨다운 검증
  - 데미지 계산 공식
  - HP 업데이트
- [ ] **4.2.3** 전투 결과 처리
  - 승리/패배 조건 확인
  - 경험치 계산 및 지급
- [ ] **4.2.4** POST /api/battle/status API 구현
- [ ] **4.2.5** 전투 관련 로그 처리
  - 전투 시작/종료 로그
  - 데미지 처리 로그

### Task 4.3: 동물 포획 시스템
**담당**: Backend  
**예상 시간**: 10시간  
**의존성**: 4.2 완료

#### Todo List:
- [ ] **4.3.1** POST /api/animal/capture API 구현
- [ ] **4.3.2** 포획 확률 계산
  - 공식: `(1 - currentHP/maxHP) * 포획구성능 * 랜덤요소`
- [ ] **4.3.3** 포획 성공/실패 처리
- [ ] **4.3.4** 포획된 동물을 플레이어 파티에 추가
- [ ] **4.3.5** AnimalCapturedEvent 발행

### Task 4.4: 장비 시스템
**담당**: Backend  
**예상 시간**: 14시간  
**의존성**: 4.1 완료

#### Todo List:
- [ ] **4.4.1** POST /api/equipment/equip API 구현
  - 장비 착용/해제 로직
  - 스탯 보너스 계산
- [ ] **4.4.2** POST /api/equipment/craft API 구현
  - 재료 소모 검증
  - 새 장비 생성
- [ ] **4.4.3** 장비 드롭 시스템
  - 동물 처치 시 재료 드롭
  - 확률 기반 희귀 재료
- [ ] **4.4.4** POST /api/inventory/items API 구현
- [ ] **4.4.5** EquipmentChangedEvent 발행

---

## 📡 Phase 5: 실시간 통신 시스템 (1-2주)

### Task 5.1: SSE (Server-Sent Events) 구현
**담당**: Backend  
**예상 시간**: 12시간  
**의존성**: 3.2 완료

#### Todo List:
- [ ] **5.1.1** GET /api/game-events SSE 엔드포인트 구현
- [ ] **5.1.2** 클라이언트 연결 관리 시스템
- [ ] **5.1.3** 이벤트 스트림을 SSE로 전송
- [ ] **5.1.4** 연결 끊김 처리 및 재연결 지원
- [ ] **5.1.5** 이벤트 필터링 (플레이어별 관련 이벤트만)

### Task 5.2: 월드 상태 동기화
**담당**: Backend  
**예상 시간**: 10시간  
**의존성**: 5.1 완료

#### Todo List:
- [ ] **5.2.1** POST /api/world/state API 구현
- [ ] **5.2.2** POST /api/world/animals API 구현
- [ ] **5.2.3** 동물 AI 이동 시뮬레이션
- [ ] **5.2.4** 동물 스폰/디스폰 로직
- [ ] **5.2.5** 맵 상태 변화 이벤트 발행

### Task 5.3: 비동기 태스크 시스템 (Asynq)
**담당**: Backend  
**예상 시간**: 8시간  
**의존성**: 3.2 완료

#### Todo List:
- [ ] **5.3.1** Asynq 워커 설정
- [ ] **5.3.2** 경험치 계산 태스크
- [ ] **5.3.3** 동물 AI 업데이트 태스크
- [ ] **5.3.4** 정기적인 맵 상태 정리 태스크
- [ ] **5.3.5** Asynqmon 웹 UI 설정

---

## 🖥️ Phase 6: 프론트엔드 기본 구현 (2-3주)

### Task 6.1: 클라이언트 기본 구조
**담당**: Frontend  
**예상 시간**: 16시간  
**의존성**: 없음 (병렬 진행 가능)

#### Todo List:
- [ ] **6.1.1** 개발 환경 선택 및 설정 (Unity 2D 또는 웹 기반)
- [ ] **6.1.2** 2D 탑뷰 카메라 설정
- [ ] **6.1.3** 기본 UI 프레임워크 설정
- [ ] **6.1.4** 스프라이트 리소스 관리 시스템
- [ ] **6.1.5** 사운드 시스템 기초 구현

### Task 6.2: 네트워크 통신 구현
**담당**: Frontend  
**예상 시간**: 12시간  
**의존성**: 6.1 완료

#### Todo List:
- [ ] **6.2.1** HTTP POST 요청 클래스 구현
- [ ] **6.2.2** SSE 클라이언트 구현
- [ ] **6.2.3** 폴백 API 호출 시스템
- [ ] **6.2.4** 응답 데이터 파싱 및 에러 처리
- [ ] **6.2.5** 네트워크 상태 모니터링

### Task 6.3: 게임 화면 구현
**담당**: Frontend  
**예상 시간**: 20시간  
**의존성**: 6.2 완료, 5.1 완료

#### Todo List:
- [ ] **6.3.1** 2D 맵 렌더링 시스템
  - 30x20 타일 맵 표시
  - 지형 요소 렌더링
- [ ] **6.3.2** 플레이어 캐릭터 구현
  - 스프라이트 애니메이션
  - 마우스 우클릭 이동
- [ ] **6.3.3** 동물 스프라이트 및 애니메이션
- [ ] **6.3.4** 전투 UI 구현
  - HP 바, 스탯 표시
  - 공격 이펙트
- [ ] **6.3.5** 인벤토리 및 장비 UI

---

## 🧪 Phase 7: 테스트 및 최적화 (1-2주)

### Task 7.1: 백엔드 테스트
**담당**: Backend  
**예상 시간**: 16시간  
**의존성**: 4.4 완료

#### Todo List:
- [ ] **7.1.1** 단위 테스트 작성 (도메인 로직)
- [ ] **7.1.2** 통합 테스트 작성 (Redis 연결)
- [ ] **7.1.3** API 엔드포인트 테스트
- [ ] **7.1.4** 부하 테스트 (동시 접속자)
- [ ] **7.1.5** 메모리 누수 검증

### Task 7.2: 시스템 통합 테스트
**담당**: Full-Stack  
**예상 시간**: 12시간  
**의존성**: 6.3 완료

#### Todo List:
- [ ] **7.2.1** 프론트엔드-백엔드 통신 테스트
- [ ] **7.2.2** SSE 연결 안정성 테스트
- [ ] **7.2.3** 게임 플레이 플로우 테스트
- [ ] **7.2.4** 다중 플레이어 동시 접속 테스트
- [ ] **7.2.5** 네트워크 단절 상황 테스트

### Task 7.3: 성능 최적화
**담당**: Backend/DevOps  
**예상 시간**: 10시간  
**의존성**: 7.2 완료

#### Todo List:
- [ ] **7.3.1** Redis 쿼리 최적화
- [ ] **7.3.2** API 응답 시간 최적화
- [ ] **7.3.3** SSE 이벤트 배치 처리
- [ ] **7.3.4** 불필요한 데이터 전송 제거
- [ ] **7.3.5** 캐싱 전략 개선

---

## 🚀 Phase 8: 배포 및 모니터링 (1주)

### Task 8.1: 배포 환경 구성
**담당**: DevOps  
**예상 시간**: 8시간  
**의존성**: 7.3 완료

#### Todo List:
- [ ] **8.1.1** Docker 컨테이너화
- [ ] **8.1.2** 환경별 설정 분리 (dev/staging/prod)
- [ ] **8.1.3** CI/CD 파이프라인 구축
- [ ] **8.1.4** 로드 밸런서 설정
- [ ] **8.1.5** SSL/TLS 인증서 적용

### Task 8.2: 모니터링 시스템
**담당**: DevOps  
**예상 시간**: 6시간  
**의존성**: 8.1 완료

#### Todo List:
- [ ] **8.2.1** 애플리케이션 로그 수집 (ELK 스택 등)
- [ ] **8.2.2** 시스템 메트릭 모니터링 (Prometheus + Grafana)
- [ ] **8.2.3** 알림 시스템 구축 (Slack, Email)
- [ ] **8.2.4** 헬스체크 및 자동 복구
- [ ] **8.2.5** 성능 대시보드 구성

---

## 📊 프로젝트 전체 요약

### 전체 예상 기간: **8-12주**
### 전체 예상 공수: **~280시간**

### Phase별 소요 시간:
- **Phase 1**: 18시간 (1-2주)
- **Phase 2**: 30시간 (1-2주)  
- **Phase 3**: 38시간 (1-2주)
- **Phase 4**: 56시간 (2-3주)
- **Phase 5**: 30시간 (1-2주)
- **Phase 6**: 48시간 (2-3주)
- **Phase 7**: 38시간 (1-2주)
- **Phase 8**: 14시간 (1주)

### 병렬 처리 가능한 태스크:
- Phase 6 (프론트엔드)는 Phase 4-5와 병렬 진행 가능
- Phase 1-2는 순차 처리 필수
- Phase 7-8은 모든 개발 완료 후 진행

### 주요 리스크 및 대응방안:
- **Redis 성능 이슈**: 초기 부하 테스트로 검증
- **SSE 연결 안정성**: 폴백 API 시스템으로 대응
- **실시간 동기화**: 이벤트 순서 보장 메커니즘 필수
- **프론트엔드 복잡도**: MVP 범위 엄격히 관리

---

**🎯 성공 기준: MVP 범위 내에서 안정적인 실시간 야생동물 포획 게임 구현**