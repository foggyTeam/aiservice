# AIService - Enhanced Version

This is an enhanced version of the AI Service that includes:

## New Features

### 1. Multiple Database Support
- **In-Memory Storage**: For development and testing
- **SQLite Storage**: For production use
- Configurable via environment variables

### 2. Environment Configuration
- **Development Mode**: Optimized for development with features like disabled caching to see fresh results
- **Production Mode**: Optimized for performance with caching enabled

### 3. Caching Layer
- **AI Response Caching**: Caches LLM responses to reduce API calls and costs
- **Database Query Caching**: Caches frequently accessed job data
- **Configurable TTL**: Time-to-live settings for cached data

### 4. Advanced Data Preprocessing
- **Raw Data Preservation**: Maintains original JSON models while adding analytical metadata
- **Spatial Relationship Analysis**: Identifies proximity, alignment, and clustering patterns between elements
- **Semantic Role Inference**: Determines element roles (headers, content, containers, connectors) based on properties
- **Content Type Analysis**: Classifies text content (titles, lists, paragraphs, links, etc.)
- **Element Clustering**: Groups related elements based on spatial proximity and alignment
- **Relationship Mapping**: Identifies connections and flows between elements
- **Hierarchical Structure Analysis**: Creates logical groupings and visual hierarchies
- **Multi-Modal Representation**: Combines raw data with spatial and semantic annotations for AI processing

## Configuration

### Environment Variables

#### Database Configuration
- `DB_TYPE`: Type of database to use ("memory" or "sqlite")
- `SQLITE_FILE_PATH`: Path to SQLite database file (default: "./aiservice.db")
- `DB_DEBUG`: Enable SQL logging (default: "false")

#### Environment Configuration
- `ENV`: Environment type ("dev" or "prod") - affects caching behavior
- `PORT`: Port to run the server on (default: "8080")

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Layer     │    │  Cache Layer    │    │  Storage Layer  │
│                 │◄──►│ (In-Memory)     │◄──►│                 │
│  /summarize     │    │ • Job caching   │    │ • In-Memory     │
│  /structurize   │    │ • AI response   │    │ • SQLite        │
│  /jobs/:id      │    │   caching       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Usage

### Development
```bash
ENV=dev DB_TYPE=memory go run cmd/server/main.go
```

### Production
```bash
ENV=prod DB_TYPE=sqlite SQLITE_FILE_PATH=/path/to/db.sqlite go run cmd/server/main.go
```

## Benefits

1. **Performance**: Caching reduces response times and API calls
2. **Scalability**: SQLite can handle more concurrent operations than in-memory storage
3. **Persistence**: Data survives application restarts with SQLite
4. **Flexibility**: Easy to switch between storage types based on environment
```