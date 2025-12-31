# Local RAG

A local Retrieval-Augmented Generation (RAG) system built with Go. Store documents locally, generate embeddings, and perform semantic search without relying on external APIs.

## Features

- **Document Processing**: Chunk and embed documents for efficient storage and retrieval
- **Vector Search**: Semantic search using cosine similarity on embeddings
- **Local Embeddings**: Uses Ollama for generating embeddings locally
- **REST API**: HTTP endpoints for document processing and search
- **CLI Tool**: Command-line interface for easy interaction
- **Batch Processing**: Process multiple documents concurrently with fan-out pattern
- **SQLite Vector DB**: Leverages sqlite-vec extension for vector operations

## Architecture

- **Chunker**: Splits documents into manageable chunks (paragraph-based by default)
- **Embedder**: Generates vector embeddings using Ollama models
- **Database**: SQLite with vector extension for storing chunks and embeddings
- **Service Layer**: Business logic for processing and searching
- **API**: REST endpoints for client interaction
- **CLI**: Terminal-based client for the API

## Prerequisites

- Go 1.21+
- [Ollama](https://ollama.com/) installed and running
- SQLite with vector extension support

### Installing Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull the default embedding model
ollama pull nomic-embed-text

# Start Ollama service
ollama serve
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/local-rag.git
cd local-rag
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
go build -o server .
```

4. Build the CLI tool:
```bash
go build -o rag cli/main.go
```

## Configuration

The application uses environment variables and a config file for configuration:

- `LOCAL_RAG_PORT`: Server port (default: 8080)
- `DB_PATH`: Database file path (default: ./local_rag.db)
- `OLLAMA_BASE_URL`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL`: Embedding model (default: nomic-embed-text)
- `SEARCH_TOP_K`: Number of results to return (default: 5)
- `CHUNKER_OVERLAP_BYTES`: Chunk overlap in bytes (default: 0)
- `BATCH_WORKER_COUNT`: Workers for batch processing (default: 4)

Config file: `~/.config/local_rag/config.yml`

Example config.yml:
```yaml
port: 8080
db_path: ./local_rag.db
search:
  top_k: 5
ollama:
  base_url: http://localhost:11434
  model: nomic-embed-text
chunker:
  overlap_bytes: 0
batch_processing:
  worker_count: 4
```

## Usage

### Starting the Server

```bash
./server
```

The server will start on port 8080 and create the database if it doesn't exist.

### Using the CLI

The `rag` CLI tool provides an easy way to interact with the RAG system.

#### Process a Document
```bash
./rag process path/to/document.txt
```

#### Batch Process Multiple Documents
```bash
./rag batch doc1.txt doc2.md doc3.pdf
```

#### Search Documents
```bash
./rag search "your query here"
```

#### Specify Custom Server URL
```bash
./rag -url http://localhost:9090 search "query"
```

### API Endpoints

#### Process Document
```bash
POST /api/process_document
Content-Type: application/json

{
  "document_name": "example.txt",
  "document_data": "<base64 encoded content>"
}
```

#### Batch Process Documents
```bash
POST /api/batch_process_documents
Content-Type: application/json

[
  {
    "document_name": "doc1.txt",
    "document_data": "<base64 encoded content>"
  },
  {
    "document_name": "doc2.txt",
    "document_data": "<base64 encoded content>"
  }
]
```

#### Search
```bash
POST /api/search
Content-Type: application/json

{
  "query": "search query"
}
```

Response:
```json
[
  {
    "document_name": "doc1.txt",
    "data": "relevant chunk content",
    "distance": 0.123
  }
]
```

## Development

### Running Tests
```bash
go test ./...
```

### Project Structure
```
.
├── cli/                    # CLI tool
├── chunker/                # Document chunking logic
├── config/                 # Configuration management
├── db/                     # Database operations
│   └── migrations/         # Database schema
├── embedding/              # Embedding generation
├── service/                # Business logic and API
├── test_data/              # Sample documents
├── main.go                 # Server entry point
└── README.md
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details
