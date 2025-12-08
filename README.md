# AI Service - Miro Board Assistant

Микросервис для интеграции AI помощника в аналог доски Miro с поддержкой анализа рукописного ввода, изображений и текста.

## Структура проекта

```
.
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── models/
│   │   └── models.go            # Data structures & schemas
│   ├── handlers/
│   │   └── analyze.go           # HTTP request handlers
│   ├── services/
│   │   └── analysis.go          # Business logic & job queue
│   └── providers/
│       ├── providers.go         # Interface definitions
│       └── implementations.go   # AI & OCR provider clients
├── go.mod
├── go.sum
└── README.md
```

## Архитектура

### Слои

1. **HTTP Handlers** (`handlers/`) - Echo фреймворк, валидация, маршрутизация
2. **Services** (`services/`) - Бизнес-логика, обработка requests, job queue
3. **Providers** (`providers/`) - Интеграция с внешними API (OpenAI, Azure, MyScript, Qwen)
4. **Models** (`models/`) - Структуры данных и JSON schemas
5. **Config** (`config/`) - Управление конфигурацией через env variables

### Pipeline обработки

```
HTTP Request
    ↓
Validation (BoardID, UserID, Input)
    ↓
Sync Processing (4s timeout)
    ├─→ TIMEOUT + CallbackURL → Enqueue Job (202 Accepted)
    └─→ Success → Return Response (200 OK)
    
Async Job Processing
    ├─→ Input Type Detection (ink/image/text)
    ├─→ Transcription (RecognizeInk / RecognizeImage / PassThrough)
    ├─→ Context Preparation (RAG-ready)
    ├─→ LLM Analysis
    └─→ Callback Delivery
```

## Входные данные (Input Schema)

### 1. Рукописный ввод (Ink)

```json
{
  "board_id": "board_123",
  "user_id": "user_456",
  "input": {
    "type": "ink",
    "strokes": [
      [
        {"x": 10.5, "y": 20.3, "t": 1700000000, "pressure": 0.8},
        {"x": 11.2, "y": 21.1, "t": 1700000010, "pressure": 0.85}
      ]
    ],
    "meta": {"pen_type": "stylus", "resolution": 300}
  },
  "context": {"page_theme": "whiteboard"}
}
```

**Параметры InkPoint:**
- `x`, `y` - координаты (float64)
- `t` - timestamp в milliseconds (опционально)
- `pressure` - давление 0-1 (опционально)
- `tilt` - угол наклона в градусах (опционально)

**Валидация:** максимум 20,000 точек на один запрос

### 2. Изображение (Image)

```json
{
  "input": {
    "type": "image",
    "image_url": "https://example.com/image.png",
    "meta": {"format": "png"}
  }
}
```

Или с Base64:
```json
{
  "input": {
    "type": "image",
    "base64": "iVBORw0KGgoAAAANSUhEUgAAAAUA...",
    "meta": {"format": "png"}
  }
}
```

### 3. Текстовый ввод (Text)

```json
{
  "input": {
    "type": "text",
    "text": "Create a diagram showing the workflow"
  }
}
```

## API Endpoints

### POST /analyze

Основной endpoint для анализа.

**Request:**
```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": "board_1",
    "user_id": "user_1",
    "input": {
      "type": "text",
      "text": "Draw a timeline from 2020 to 2025"
    },
    "callback_url": "https://miro.example.com/webhook"
  }'
```

**Response (200 OK - Sync):**
```json
{
  "intent": "create_timeline",
  "confidence": 0.92,
  "actions": [
    {
      "type": "create_shape",
      "payload": {
        "shape_type": "timeline",
        "start_year": 2020,
        "end_year": 2025,
        "position": {"x": 100, "y": 200}
      }
    }
  ],
  "explanation": "Detected request to create timeline visualization",
  "metadata": {
    "transcription_meta": {...}
  }
}
```

**Response (202 Accepted - Async):**
```json
{
  "job_id": "job_1700000000123456789",
  "status": "pending",
  "created_at": 1700000000,
  "expires_at": 1700086400
}
```

### GET /jobs/:id

Получить статус задачи.

**Response:**
```json
{
  "id": "job_123",
  "status": "completed",
  "created_at": 1700000000,
  "request": {...},
  "retries": 0
}
```

### GET /health

Health check.

## Конфигурация

Через environment variables:

```bash
# Server
PORT=8080
ENV=dev

# LLM Provider
LLM_PROVIDER=openai           # openai | qwen | anthropic
LLM_API_KEY=sk-...
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4-turbo
LLM_TIMEOUT=20s

# OCR/Ink Recognition
OCR_PROVIDER=azure            # azure | myscript | google
OCR_API_KEY=...
OCR_BASE_URL=https://...
OCR_TIMEOUT=8s

# Job Queue
JOB_QUEUE_SIZE=100
JOB_WORKERS=2
JOB_MAX_RETRIES=3
JOB_RETRY_BACKOFF=2s

# Timeouts
TIMEOUT_SYNC_PROCESS=4s
TIMEOUT_INK_RECOGNIZE=8s
TIMEOUT_LLM_REQUEST=20s
```

## Провайдеры

### LLM Providers

#### OpenAI (GPT-4, GPT-3.5)
```bash
LLM_PROVIDER=openai
LLM_API_KEY=sk-...
LLM_MODEL=gpt-4-turbo
```

#### Qwen (Alibaba)
```bash
LLM_PROVIDER=qwen
LLM_API_KEY=...
LLM_BASE_URL=https://dashscope.aliyuncs.com/api/v1
LLM_MODEL=qwen-max
```

#### Anthropic Claude
```bash
LLM_PROVIDER=anthropic
LLM_API_KEY=sk-ant-...
LLM_MODEL=claude-3-opus
```

### OCR/Ink Recognition Providers

#### Azure Ink Recognizer
```bash
OCR_PROVIDER=azure
OCR_API_KEY=<key>
OCR_BASE_URL=https://<region>.api.cognitive.microsoft.com/inkrecognizer
```

**Особенности:**
- Поддержка рукописного текста (диаграммы, уравнения, фигуры)
- Поддержка 63+ языков
- REST API

#### MyScript
```bash
OCR_PROVIDER=myscript
OCR_API_KEY=<key>
OCR_BASE_URL=https://api.myscript.com
```

**Особенности:**
- Распознавание формул и уравнений
- Распознавание фигур и диаграмм
- Поддержка 70+ языков
- REST & WebSocket API

#### Google Document AI
```bash
OCR_PROVIDER=google
OCR_API_KEY=<key>
OCR_BASE_URL=https://documentai.googleapis.com
```

**Особенности:**
- Мощный OCR
- Распознавание формуляров и таблиц
- Поддержка 250+ языков

## Best Practices for Integration

### 1. Структурированный Input

**Всегда включайте метаданные:**
```json
{
  "input": {
    "type": "ink",
    "strokes": [...],
    "meta": {
      "board_language": "en",
      "pen_color": "#000000",
      "brush_size": 2,
      "element_type": "diagram"  // важно для контекста
    }
  },
  "context": {
    "board_size": {"width": 1920, "height": 1080},
    "existing_elements": ["shape_123", "text_456"],
    "user_settings": {"theme": "dark"}
  }
}
```

### 2. Обработка рукописного ввода

**Вариант 1: Local Pre-processing (Рекомендуется)**
- На фронте распознавайте простые фигуры (линии, круги, прямоугольники)
- Отправляйте только сложный ввод (текст, диаграммы) на сервер
- Это экономит bandwidth и latency

```javascript
// Frontend
const strokes = captureInkInput();
const simpleShapes = recognizeSimpleShapes(strokes); // Местное распознавание
const complexStrokes = filterComplexStrokes(strokes);

// Отправляем только сложное
fetch('/analyze', {
  input: { type: 'ink', strokes: complexStrokes },
  context: { simple_shapes: simpleShapes }
});
```

**Вариант 2: Server-side Full Processing**
- Отправляйте все strokes на сервер
- Сервер делает распознавание
- Более точно, но медленнее

### 3. Промпт-инженеринг для LLM

```go
// В providers/implementations.go улучшите buildPrompt:

prompt := fmt.Sprintf(`Analyze this board input. You are an assistant for a collaborative whiteboard.

User's input: %s

Context:
- Board size: 1920x1080
- Existing elements: %d objects
- User language: en

Return structured JSON with:
{
  "intent": "one_of: [create_shape, create_text, create_diagram, create_chart, draw_annotation]",
  "confidence": 0.0-1.0,
  "actions": [
    {
      "type": "action_name",
      "payload": {...action-specific data...}
    }
  ],
  "explanation": "why this action was chosen"
}

Be concise and accurate.`, 
  transcription, 
  contextData,
)
```

### 4. Async Processing & Callbacks

**Webhook Callback Format:**
```json
{
  "job_id": "job_123",
  "status": "success",
  "result": {
    "intent": "create_chart",
    "confidence": 0.95,
    "actions": [...]
  },
  "timestamp": 1700000000
}
```

**Error Callback:**
```json
{
  "job_id": "job_123",
  "status": "error",
  "error": "LLM request timeout",
  "timestamp": 1700000000
}
```

## Развертывание

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

```bash
docker build -t aiservice:latest .
docker run -e PORT=8080 -e LLM_PROVIDER=openai -e LLM_API_KEY=$KEY aiservice:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aiservice
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: aiservice
        image: aiservice:latest
        ports:
        - containerPort: 8080
        env:
        - name: LLM_PROVIDER
          value: "openai"
        - name: LLM_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: openai-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## Тестирование

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Integration test
go test -tags=integration ./...
```

## Troubleshooting

### Timeout issues

Если видите много 202 responses:
1. Увеличьте `TIMEOUT_SYNC_PROCESS` (но не больше 10s)
2. Проверьте timeout на LLM_TIMEOUT и OCR_TIMEOUT
3. Масштабируйте job workers: `JOB_WORKERS=4`

### High latency

1. Убедитесь, что используете правильный LLM_MODEL (gpt-4 медленнее чем gpt-3.5)
2. Добавьте caching слой (Redis)
3. Реализуйте batch processing для похожих requests

### OCR accuracy

Для рукописного ввода:
- MyScript обычно лучше распознает формулы и диаграммы
- Azure хорош для текста
- Комбинируйте: сначала простое распознавание MyScript, потом LLM для контекста

## Дальнейшие улучшения

- [ ] Добавить Redis для caching и job storage
- [ ] RAG (Retrieval Augmented Generation) для контекста доски
- [ ] WebSocket support для real-time updates
- [ ] Rate limiting & authentication
- [ ] Metrics & monitoring (Prometheus)
- [ ] Database для persistent job storage
- [ ] Поддержка batch requests
- [ ] Model fine-tuning на примерах вашего domain

## Лицензия

MIT
