# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LIFE is an event-driven CQRS-based distributed task system for a wildlife ecosystem MMORPG. The project focuses on African savanna-style wildlife capture and management mechanics.

## Architecture

The project follows a 3-layer + CQRS structure:

```
API Layer     → REST API + JSON-RPC 2.0
Application   → CQRS (Command/Query Handlers)  
Domain        → DDD (Business Logic)
Infrastructure → IoC Repository + Event Bus
Persistence   → Redis (Storage + Cache)
```

### Core Design Principles
- **Stateless Server**: No state storage in server memory
- **Event-Driven**: Event-based inter-service communication
- **Domain-Driven**: Game business logic-centered design
- **IoC Repository**: Infrastructure dependency inversion for minimal coupling
- **Distributed Processing**: Task queue-based async processing

## Technology Stack

- **Language**: Go 1.21+
- **HTTP**: Standard `net/http` package
- **Communication**: REST API + JSON-RPC 2.0 hybrid
- **Storage**: Redis (persistent storage + cache)
- **Event Streaming**: Redis Streams with Watermill package
- **Task Queue**: Asynq (Redis-based distributed task queue)
- **Documentation**: Swagger (OpenAPI 2.0)

## Communication Protocol

Uses REST-style URLs with JSON-RPC 2.0 request bodies:

**URL Pattern:**
```
POST /api/trainer/create
POST /api/animal/capture  
POST /api/habitat/manage
```

**Request Body Example:**
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

## Domain Model

Core game domains:
- **Trainer**: Trainer-related business logic (level, skills, experience)
- **Animal**: Wild animal capture, evolution, state management
- **Habitat**: Habitat and ecosystem systems (savanna, jungle, desert)
- **Tribe**: Tribe system (future expansion)

## Event System

Key domain events:
- `TrainerCreatedEvent`: Trainer creation
- `AnimalCapturedEvent`: Animal capture
- `TrainerLevelUpEvent`: Trainer level up
- `HabitatChangedEvent`: Habitat changes

## Development Status

This appears to be a project in planning/early development phase. The repository currently contains only documentation (README.md) without actual implementation code.

## Project Structure

The project is designed but not yet implemented. Based on the README, the following structure is planned:
- Event-driven architecture using Redis Streams
- CQRS pattern with separate command and query handlers
- Domain-driven design with clear business logic separation
- IoC repository pattern for infrastructure abstraction