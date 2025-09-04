# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LIFE is an event-driven CQRS-based distributed wildlife ecosystem MMORPG with real-time multiplayer mechanics. The project features African savanna-style wildlife interaction, real-time movement, and advanced bullet system architecture.

## Architecture

The project follows Clean Architecture with Event-Driven CQRS:

```
API Layer     → REST + JSON-RPC 2.0 hybrid with Auto-Router
Application   → CQRS Commands/Queries + Event Bus
Domain        → DDD Aggregates (Trainer, Animal, World)
Infrastructure → IoC Repository + Redis + Watermill
Persistence   → Redis (storage + cache + event streams)
```

### Core Design Principles
- **Stateless Server**: No state storage in server memory
- **Event-Driven**: Watermill + Redis Streams for event sourcing
- **Domain-Driven**: Rich domain models with business logic
- **Auto-Router**: Reflection-based handler registration
- **Real-time**: SSE for live position synchronization

## Technology Stack

- **Language**: Go 1.23+ with generics support
- **HTTP**: Standard `net/http` with custom middleware chains
- **Storage**: Redis v9.12.1 (primary database + cache + queues)
- **Events**: Watermill + Redis Streams for event sourcing
- **Tasks**: Asynq for distributed background processing
- **Auth**: JWT with OAuth integration (Google, GitHub, Discord)
- **Logging**: Zap for high-performance structured logging
- **Docs**: Swagger/OpenAPI 2.0 auto-generation

## Development Commands

```bash
# Run server (with hot reload)
go run cmd/server/main.go

# Run with specific port
SERVER_PORT=8080 go run cmd/server/main.go

# Build server binary
go build -o bin/server cmd/server/main.go

# Run tests
go test ./...

# Run specific package tests
go test ./internal/app/...

# Test with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

```
cmd/
├── server/         # HTTP server entry point
└── worker/         # Background worker (planned)

internal/
├── api/            # HTTP handlers + middleware
├── app/            # CQRS commands/queries/handlers  
├── cqrs/           # Event system + SSE utilities
├── domain/         # DDD aggregates + domain logic
└── infrastructure/ # Repository implementations

pkg/
├── autorouter/     # Reflection-based API routing
└── [shared libs]   # Reusable components

docs/               # Architecture + API documentation
```

## Communication Protocol

**Auto-Router Pattern**: REST URLs + JSON-RPC 2.0 bodies

```go
// Handler method naming convention
func (h *TrainerHandler) Create(ctx context.Context, cmd CreateTrainerCommand) (*CreateTrainerResult, error)
func (h *TrainerHandler) Move(ctx context.Context, cmd MoveTrainerCommand) (*MoveTrainerResult, error)

// Auto-generated routes:
// POST /api/v1/trainer.Create
// POST /api/v1/trainer.Move
```

**Request Example:**
```json
{
  "jsonrpc": "2.0",
  "method": "trainer.move", 
  "params": {
    "direction": {"x": 1.0, "y": 0.0},
    "timestamp": 1674123456.789
  },
  "id": 1
}
```

## Domain Model (DDD)

**Core Aggregates:**
- **Trainer**: Player character with position, movement, inventory, party
- **Animal**: Wildlife creatures with AI, states, behaviors
- **World**: Environment, zones, weather, resources
- **Account**: Authentication, profiles, OAuth linking

**Key Features Implemented:**
- Real-time movement with client-side prediction
- Inventory management with type validation  
- Experience/leveling progression system
- JWT authentication with OAuth providers
- Event-driven state synchronization via SSE

## Event System

**Domain Events** (defined in `internal/cqrs/events.go`):
- `TrainerMovedEvent`: Real-time position updates
- `TrainerCreatedEvent`: New player registration
- `SSENotificationEvent`: Generic real-time notifications

**Event Flow:**
```
Command → Handler → Domain Logic → Event → Watermill → Redis Streams → SSE Clients
```

## Middleware Chain

Standard middleware stack (in `internal/api/middleware/`):
- **CORS**: Cross-origin resource sharing
- **Auth**: JWT token validation
- **Logging**: Request/response logging with Zap
- **Recovery**: Panic recovery with error reporting
- **Static**: Static file serving for client assets

## Authentication System

- **JWT-based** authentication with refresh tokens
- **OAuth integration**: Google, GitHub, Discord providers
- **Account linking**: N:1 pattern (multiple OAuth → single account)
- **Middleware protection**: Route-level auth requirements

## Real-time Features

**SSE (Server-Sent Events):**
- Live position synchronization for multiplayer
- Event-driven updates (trainer movement, state changes)
- Client reconnection handling

**Movement System:**
- Client-side prediction with server reconciliation
- JSON merge patches for efficient updates
- Movement state tracking (start/stop/direction)

## Redis Usage Patterns

**Storage**: Primary database for all game state
**Caching**: High-frequency read optimization  
**Events**: Redis Streams for event sourcing
**Tasks**: Asynq queues for background processing
**Sessions**: JWT token blacklisting and session management

## Code Patterns

**Repository Pattern**: Interface-based with IoC dependency injection
**Command/Query**: Separate read/write operations (CQRS)
**Event Sourcing**: Append-only event streams with projections
**Error Handling**: Wrapped errors with context using `oops` library
**Logging**: Structured logging with correlation IDs

## Current Development Status

**✅ Completed:**
- Core CQRS infrastructure with auto-router
- Trainer movement system with real-time sync
- Authentication with OAuth integration
- Event sourcing with Watermill + Redis Streams
- SSE real-time communication
- Comprehensive API documentation

**🚧 In Progress:**
- Bullet firing system (advanced mathematical collision detection)
- Animal AI system with behavior patterns
- Advanced game mechanics (inventory, combat, etc.)

**📋 Planned:**
- Distributed worker scaling
- Advanced anti-cheat systems
- Mobile client support