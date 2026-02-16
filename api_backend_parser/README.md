# API Backend Parser

Скрипт для получения данных доски из foggy backend и отправки их в AIService или Google Handwriting API для анализа.

## Установка и запуск

### Запуск с параметрами по умолчанию:
```bash
go run main.go
```

### Запуск с указанием конкретного board ID:
```bash
go run main.go -board <BOARD_ID>
```

### Запуск в режиме structurize:
```bash
go run main.go -board <BOARD_ID> -mode structurize
```

### Запуск с использованием Google Handwriting API:
```bash
go run main.go -board <BOARD_ID> -google
```

### Запуск с Google API и указанием языка:
```bash
go run main.go -board <BOARD_ID> -google -google-lang zh_TW
```

## Параметры командной строки

| Параметр | По умолчанию | Описание |
|----------|--------------|----------|
| `-board` | `69952d428e6fb9a24efa3bbd` | ID доски для получения из foggy backend |
| `-mode` | `summarize` | Режим работы: `summarize` или `structurize` |
| `-foggy-url` | `http://localhost:3001` | URL foggy backend |
| `-ai-url` | `http://localhost:8080` | URL AIService |
| `-user` | `parser-user` | User ID для запроса |
| `-output` | (пусто) | Файл для сохранения ответа (опционально) |
| `-google` | `false` | Использовать Google Handwriting API вместо AIService |
| `-google-lang` | `en` | Код языка для Google Handwriting API (en, zh_TW, ja, ko, и т.д.) |

## Примеры использования

### Получить суммаризацию доски:
```bash
go run main.go -board 69952d428e6fb9a24efa3bbd -mode summarize
```

### Получить структуризацию доски:
```bash
go run main.go -board 69952d428e6fb9a24efa3bbd -mode structurize
```

### Распознать почерк через Google API:
```bash
go run main.go -board 69952d428e6fb9a24efa3bbd -google
```

### Распознать почерк через Google API с указанием языка (китайский):
```bash
go run main.go -board 69952d428e6fb9a24efa3bbd -google -google-lang zh_TW
```

### Распознать почерк через Google API с указанием языка (японский):
```bash
go run main.go -board 69952d428e6fb9a24efa3bbd -google -google-lang ja
```

### Сохранить ответ в файл:
```bash
go run main.go -board 69952d428e6fb9a24efa3bbd -output response.json
```

### Работа с production окружением:
```bash
go run main.go \
  -board 69952d428e6fb9a24efa3bbd \
  -foggy-url https://foggy-backend.example.com \
  -ai-url https://ai-service.example.com \
  -user production-user
```

## Преобразование данных

Скрипт преобразует элементы из foggy backend в формат AIService:

| Foggy тип | AIService тип |
|-----------|---------------|
| `line`, `pencil`, `freehand`, `drawing` | `line` |
| `text`, `textbox`, `text-element` | `text` |
| `rectangle`, `rect`, `box` | `rectangle` |
| `ellipse`, `circle`, `oval` | `ellipse` |

## Структура запроса к AIService

### Summarize:
```json
{
  "requestId": "foggy-<board_id>-<timestamp>",
  "userId": "<user_id>",
  "requestType": "summarize",
  "board": {
    "boardId": "<board_id>",
    "imageUrl": "",
    "elements": [...]
  }
}
```

### Structurize:
```json
{
  "requestId": "foggy-<board_id>-<timestamp>",
  "userId": "<user_id>",
  "requestType": "structurize",
  "board": {
    "boardId": "<board_id>",
    "imageUrl": "",
    "elements": [...]
  },
  "file": {
    "name": "<board_name>",
    "type": "doc",
    "children": []
  }
}
```

## Структура запроса к Google Handwriting API

При использовании флага `-google` запрос отправляется на Google Handwriting Recognition API:

```json
{
  "options": "enable_pre_space",
  "requests": [
    {
      "writing_guide": {
        "writing_area_width": <width>,
        "writing_area_height": <height>
      },
      "ink": [
        [[x1, x2, x3...], [y1, y2, y3...]],  // Stroke 1
        [[x1, x2, x3...], [y1, y2, y3...]]   // Stroke 2
      ],
      "language": "en"
    }
  ]
}
```

### Параметры Google API:

| Параметр | Описание |
|----------|----------|
| `options` | Фиксированное значение `"enable_pre_space"` |
| `requests[].ink` | Массив штрихов, каждый штрих содержит [x-координаты, y-координаты] |
| `requests[].language` | Код языка распознавания |
| `requests[].writing_guide` | Опционально: размеры области письма |

### Поддерживаемые языки:

| Язык | Код | Язык | Код |
|------|-----|------|-----|
| Английский | `en` | Китайский (традиционный) | `zh_TW` |
| Китайский (упрощенный) | `zh_CN` | Японский | `ja` |
| Корейский | `ko` | Французский | `fr` |
| Немецкий | `de` | Испанский | `es` |

*Примечание: Только `zh_TW`, `en`, и `ja` полностью протестированы.*

## Выходные данные

Скрипт выводит:
1. Информацию о полученной доске
2. Количество преобразованных элементов
3. Статус ответа от AIService
4. JSON ответ от AIService

При указании `-output` ответ также сохраняется в файл.
