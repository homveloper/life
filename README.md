# 🌿 LIFE: Living Intelligence for Fauna Ecosystem

**Event-Driven + CQRS + 분산 태스크 기반 야생 동물 포획 & 생태계 MMORPG**

## 🎯 프로젝트 목표

- 🦁 **야생 동물 생태계**: 아프리카 사바나 같은 대규모 야생 필드에서 동물 포획 & 관리
- 🎮 **MMORPG 핵심 기술 검증**: Event-Driven, CQRS, 분산 태스크
- 📊 **아키텍처 패턴 실험**: 현대적 게임서버 기술 스택 학습  
- 🚀 **확장 가능한 기반**: 추후 실제 게임 개발을 위한 토대 구축
- 💡 **기술 역량 증명**: 포트폴리오용 구체적 결과물

## 🏗️ 아키텍처 개요

### **3계층 + CQRS 구조**
```
┌─────────────────┐
│   API Layer     │ ← REST API + JSON-RPC 2.0
├─────────────────┤
│ Application     │ ← CQRS (Command/Query Handlers)
├─────────────────┤
│   Domain        │ ← DDD (비즈니스 로직)
├─────────────────┤
│ Infrastructure  │ ← IoC Repository + Event Bus
└─────────────────┘
      ↓
┌─────────────────┐
│   Persistence   │ ← Redis (영구 저장소 + 캐시)
└─────────────────┘
```

### **핵심 설계 원칙**
- **Stateless Server**: 서버 메모리에 상태 저장 금지
- **Event-Driven**: 이벤트 기반 서비스 간 통신  
- **Domain-Driven**: 게임 비즈니스 로직 중심 설계
- **IoC Repository**: 인프라 의존성 역전으로 결합도 최소화
- **분산 처리**: 태스크 큐 기반 비동기 처리

## 🛠️ 기술 스택

### **언어 및 프레임워크**
- **언어**: Go 1.21+
- **HTTP**: 표준 `net/http` 패키지
- **통신**: REST API + JSON-RPC 2.0
- **저장소**: Redis (영구 저장소 + 캐시)
- **문서화**: Swagger (OpenAPI 2.0)
- **로깅**: Zap (고성능 로깅 라이브러리)

### **분산 시스템**
- **이벤트 스트리밍**: Redis Streams (Watermill 패키지)
- **태스크 큐**: Asynq (Redis 기반 분산 태스크 큐)


## 📡 통신 프로토콜

### **REST + JSON-RPC 2.0 하이브리드**

**URL (REST 스타일)**
```
POST /api/v1/trainer.Create
POST /api/v1/animal.Capture
POST /api/v1/habitat.Manage
```

**요청 바디 (JSON-RPC 2.0)**
```json
{
  "jsonrpc": "2.0",
  "method": "trainer.create",
  "params": {
    "trainerId": "ranger_123",
    "nickname": "WildlifeExplorer"
  },
  "id": 1
}
```

## ⚡ Event-Driven 아키텍처

### **이벤트 기반 통신**
- **도메인 이벤트**: 게임 내 모든 행동을 이벤트로 처리
- **Watermill**: Redis Streams 기반 이벤트 메시징 라이브러리
- **Redis Streams**: 이벤트 스트리밍 및 메시지 영속화
- **느슨한 결합**: 서비스 간 이벤트 기반 비동기 통신

### **이벤트 종류**
- `TrainerCreatedEvent`: 조련사 생성
- `AnimalCapturedEvent`: 동물 포획
- `TrainerLevelUpEvent`: 조련사 레벨업
- `HabitatChangedEvent`: 서식지 변화

## 🔄 CQRS 패턴

### **Command (쓰기)**
- **Purpose**: 게임 상태 변경
- **Flow**: REST → Command Handler → Domain → Event Store
- **Examples**: CaptureAnimal, CreateTrainer, ManageHabitat

### **Query (읽기)**  
- **Purpose**: 게임 데이터 조회
- **Flow**: REST → Query Handler → Read Model → Response
- **Optimization**: 읽기 전용 최적화된 데이터베이스 뷰

## 🔄 IoC Repository 패턴

### **Inversion of Control**
- **의존성 역전**: 도메인이 인프라에 의존하지 않음
- **인터페이스 추상화**: Repository 인터페이스로 구현체 분리
- **테스트 용이성**: Mock Repository로 단위 테스트 지원
- **유연한 교체**: Redis → PostgreSQL 등 저장소 변경 용이

### **레포지토리 계층 구조**
```
Domain Layer      → Repository Interface (추상화)
Infrastructure    → Repository Implementation (구현체)  
Persistence       → Redis Client (실제 저장소)
```

### **인프라 노출 최소화**
- **도메인 순수성**: 비즈니스 로직에서 Redis 의존성 제거
- **클린 아키텍처**: 외부 의존성으로부터 도메인 보호
- **계층 분리**: 각 레이어의 명확한 역할과 책임

## 🌐 분산 태스크 시스템 (Asynq)

### **Asynq 태스크 큐**
- **Redis 기반**: Redis를 브로커로 사용하는 분산 태스크 큐
- **워커 풀**: 다중 서버에서 태스크 병렬 처리
- **재시도 로직**: 지수 백오프 기반 자동 재시도
- **웹 UI**: Asynqmon을 통한 실시간 모니터링

### **태스크 종류**
- **경험치 계산**: 복잡한 게임 로직 비동기 처리
- **알림 발송**: 플레이어 알림 및 이메일 전송
- **데이터 집계**: 게임 통계 및 랭킹 계산
- **정기 작업**: 일일 퀘스트, 서버 정비 등

## 🧠 DDD 도메인 설계

### **게임 도메인 모델**
- **Trainer**: 조련사 관련 비즈니스 로직 (레벨, 스킬, 경험치)
- **Animal**: 야생 동물 포획, 진화, 상태 관리  
- **Habitat**: 서식지 및 생태계 시스템 (사바나, 정글, 사막 등)
- **Tribe**: 부족 시스템 (향후 확장)

### **도메인 이벤트**
- 모든 비즈니스 로직은 도메인 이벤트 발생
- 이벤트를 통한 서비스 간 느슨한 결합
- 이벤트 기반 상태 동기화


## 🚀 기대 효과

### **기술적 성과**
- **Go 언어**: 동시성 프로그래밍 마스터
- **분산 시스템**: 이벤트 기반 아키텍처 경험
- **CQRS**: 읽기/쓰기 분리 패턴 적용
- **도메인 모델링**: DDD 실제 프로젝트 적용

### **비즈니스 가치**  
- **확장 가능**: 실제 게임서버로 발전 가능한 구조
- **재사용성**: 다른 게임 프로젝트에 응용 가능
- **포트폴리오**: 현대적 아키텍처 역량 증명

## 📈 다음 단계

1. **기술 스택 세팅**: Go + Watermill + Asynq + Redis
2. **기본 인프라 구축**: Redis Streams Event Bus + CQRS API
3. **도메인 모델링**: 조련사, 야생동물, 서식지
4. **기능 구현**: 순차적 기능 개발
5. **성능 테스트**: 부하 테스트 및 최적화
6. **문서화**: 아키텍처 결과 정리 및 공유

---

**현대적 게임서버 아키텍처의 핵심 기술들을 실험하고 검증하는 야생 동물 생태계 MMORPG** 🎯

*Built with Go + Watermill + Asynq + Redis*