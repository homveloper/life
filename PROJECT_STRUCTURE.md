# 🏗️ LIFE 프로젝트 구조

## 디렉토리 구조

```
life/
├── cmd/                    # 애플리케이션 엔트리포인트
│   ├── server/            # HTTP 서버 애플리케이션
│   └── worker/            # Asynq 워커 애플리케이션
├── internal/              # 내부 패키지 (외부에서 import 불가)
│   ├── api/               # HTTP API 레이어
│   │   ├── handlers/      # HTTP 핸들러
│   │   └── middleware/    # HTTP 미들웨어
│   ├── application/       # CQRS 애플리케이션 레이어
│   │   ├── command/       # Command 핸들러 (쓰기 작업)
│   │   └── query/         # Query 핸들러 (읽기 작업)
│   ├── domain/            # 도메인 모델 (DDD)
│   │   ├── trainer/       # 조련사 도메인
│   │   ├── animal/        # 동물 도메인
│   │   ├── equipment/     # 장비 도메인
│   │   └── world/         # 월드/맵 도메인
│   └── infrastructure/    # 인프라스트럭처 레이어
│       ├── repository/    # Redis 리포지토리 구현체
│       └── eventstore/    # Redis Streams 이벤트 스토어
├── pkg/                   # 외부 공유 가능 패키지
│   ├── config/           # 설정 관리
│   ├── logger/           # 로깅 시스템
│   └── redis/            # Redis 클라이언트 래퍼
└── docs/                 # 프로젝트 문서
    ├── MVP_PLANNING.md
    └── PROJECT_SETUP_WORKFLOW.md
```

## 아키텍처 레이어 설명

### 1. cmd/ - 애플리케이션 엔트리포인트
- **server/**: HTTP API 서버 (메인 게임 서버)
- **worker/**: 비동기 태스크 처리 워커 (Asynq)

### 2. internal/api/ - HTTP API 레이어
- **handlers/**: REST API 엔드포인트 핸들러
- **middleware/**: CORS, 로깅, 인증 등 미들웨어

### 3. internal/application/ - CQRS 애플리케이션 레이어
- **command/**: 상태 변경 작업 (POST 요청 처리)
- **query/**: 상태 조회 작업 (상태 조회 POST 요청 처리)

### 4. internal/domain/ - 도메인 레이어 (DDD)
- **trainer/**: 조련사 엔티티, 값 객체, 도메인 서비스
- **animal/**: 동물 엔티티, 스탯, 포획 로직
- **equipment/**: 장비 아이템, 스탯 보너스
- **world/**: 맵, 위치, 스폰 시스템

### 5. internal/infrastructure/ - 인프라스트럭처 레이어
- **repository/**: 도메인 리포지토리의 Redis 구현체
- **eventstore/**: Redis Streams 기반 이벤트 스토어

### 6. pkg/ - 공유 패키지
- **config/**: 환경 설정 및 설정 파일 관리
- **logger/**: Zap 기반 구조화된 로깅
- **redis/**: Redis 클라이언트 연결 및 유틸리티

## 의존성 규칙 (Clean Architecture)

```
API Layer        → Application Layer
Application      → Domain Layer
Infrastructure   → Domain Layer (인터페이스 구현)
Package          → 어느 레이어든 사용 가능
```

- **Domain Layer**: 다른 레이어에 의존하지 않음 (순수 비즈니스 로직)
- **Application Layer**: Domain Layer에만 의존
- **Infrastructure**: Domain의 인터페이스를 구현
- **API Layer**: Application Layer를 통해 요청 처리

이 구조는 **Stateless**, **Event-Driven**, **CQRS** 원칙을 준수하며, 각 레이어의 책임이 명확히 분리되어 있습니다.