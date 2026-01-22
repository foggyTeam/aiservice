# Какие запросы я ожидаю:

TODO: Добавить явные ограничения в промтп

### Типы запросов
1) structureAnalyze
2) graphAnalyze
3) complexAnalyze

### Запросы
1) **structureAnalyze** - Запрос на организацию файлов:


```Go
type UserRequest struct {
	RequestID  string         `json:"requestId,omitempty"`
	BoardID    string         `json:"boardId" validate:"required"`
	UserID     string         `json:"userId" validate:"required"`
	Type       string         `json:"type" validate:"required"`
  UserPrompt string         `json:"userPrompt" validate:"required"`
}
```

```Go
type File struct {
    Name     string
    Content  string // Content должен быть в html формате
    Children []File
}
```

```Go
type StructureAnalyzeRequest struct {
    UserRequest
    Files      []File         `json:"files,omitempty"`
}
```

```html
- Может быть уже какая-то структура, тогда добавить в промтп уже существующую структуру
    - Prompt: userRequest.UserPrompt. 
      You need to stick to html format in Content.
For example:
    <h1 color="red">заголовок</h1>
        <ul>
            <li>список</li>
            <li>список</li>
        </ul>
    <p>текст</p>
```

```html
- Если поле пустое, тогда, нужно явно сгенерировать новую файловую структуру для пользователя
    - Prompt: userRequest.UserPrompt
      You need to stick to html format in Content.
For example:
    <h1 color="red">заголовок</h1>
        <ul>
            <li>список</li>
            <li>список</li>
        </ul>
    <p>текст</p>
```

```Go
type StructureAnalyzeResponse struct {
    LlmAnswer string `json:"llmAnswer"`
    Files []File `json:"files,omitempty"`
}
```

2) **graphAnalyze** - Запрос на создание, улучшение какого-то графа:

```Go
type GraphAnalyzeRequest struct {
    UserRequest
    Graph       string `json:"graph"`
}
```

```txt
Prompt: userRequest.UserPrompt. 
        You need to stick to html format in Content.
For example:
    // TODO Тут нужен пример графа, пока пусто, т.к. не определились с форматом.

```

```Go
type GraphResponse struct {
    LlmAnswer string `json:"llmAnswer"`
    Graph string `json:"graph"`
}
```

3) **complexAnalyze** - комплексный анализ запроса пользователя 

- Пока InkInput под вопросом, возможно ai будет нормально анализировать и по картинкам, тогда координаты нам будут не нужны.
- Из первого следует, что придется менять структуру запроса графа. Если граф будет не просто какой-то единный формат,
    то придется придумать еще один тип, который будет описывать то, что изображено на доске
```Go
type InkInput struct {
	Type    string         `json:"type"`
	Strokes [][]InkPoint   `json:"strokes" validate:"required"`
}

type InkPoint struct {
	X        float64 `json:"x" validate:"required"`
	Y        float64 `json:"y" validate:"required"`
	T        int64   `json:"t,omitempty"`        // timestamp in ms
	Pressure float64 `json:"pressure,omitempty"` // 0-1
	Tilt     float64 `json:"tilt,omitempty"`     // angle in degrees
}
```
```Go
type ComplexAnalyzeRequest struct {
    UserRequest
    Graph       string `json:"graph"`
    Files       []File `json:"files"`
    ImageURL    string `json:"imageUrl"`
    InkInput    InkInput    
}
```

```txt
Prompt: userRequest.UserPromt
        You need to analyze:
            1) [if not empty] userGraph
                    + [if not empty] analyze user graph
                + You need to stick to ... format in Files.
            2) [if not empty] image
                + analyze image
            3) [if not empty] user ink input
                + analyze user ink input
            4) [if not empty] userFiles
                    + [if not empty] analyze user graph
                + You need to stick to html format in Content.
```

```Go
type ComplexAnalyzeResponse struct {
    LlmAnswer string `json:"llmAnswer"`
    Graph     string `json:"graph",omitempty`
    Files     []File `json:"files,omitempty"`
}
```

### Примеры запросов

// Request: StructureAnalyzeRequest (example with existing files)
```json
{
  "requestId": "r-001",
  "boardId": "board-123",
  "userId": "user-abc",
  "type": "structureAnalyze",
  "userPrompt": "Create a website structure for a personal blog about travel.",
  "files": [
    {
      "name": "index.html",
      "content": "<h1>Home</h1><p>Welcome</p>",
      "children": []
    },
    {
      "name": "about.html",
      "content": "<h1>About</h1><p>Author bio</p>",
      "children": []
    }
  ]
}
```

// Response: StructureAnalyzeResponse
```json
{
  "llmAnswer": "Generated a clear site structure with pages, blog posts, and assets.",
  "files": [
    {
      "name": "index.html",
      "content": "<h1>Home</h1><p>Welcome to my travel blog</p>",
      "children": []
    },
    {
      "name": "posts",
      "content": "",
      "children": [
        {
          "name": "2025-01-01-first-trip.html",
          "content": "<h1>First Trip</h1><p>Story...</p>",
          "children": []
        }
      ]
    },
    {
      "name": "about.html",
      "content": "<h1>About</h1><p>Author bio and contact</p>",
      "children": []
    }
  ]
}
```

// Request: GraphAnalyzeRequest (example)
```json
{
  "requestId": "r-010",
  "boardId": "board-graph-1",
  "userId": "user-xyz",
  "type": "graphAnalyze",
  "userPrompt": "Improve this site sitemap and add categories.",
  "graph": "{ \"nodes\": [{\"id\":\"home\"},{\"id\":\"posts\"}], \"edges\": [{\"from\":\"home\",\"to\":\"posts\"}] }"
}
```

// Response: GraphResponse
```json
{
  "llmAnswer": "Added category nodes and reorganized edges for SEO-friendly structure.",
  "graph": "{ \"nodes\": [{\"id\":\"home\"},{\"id\":\"posts\"},{\"id\":\"category-europe\"},{\"id\":\"category-asia\"}], \"edges\": [{\"from\":\"home\",\"to\":\"posts\"},{\"from\":\"posts\",\"to\":\"category-europe\"},{\"from\":\"posts\",\"to\":\"category-asia\"}] }"
}
```
