# Chromem-go Experiments

Experiments with vector embeddings for scripture study using chromem-go.

## Purpose

Explore different chunking strategies for scripture content:
- Per sentence
- Per verse  
- Per chapter
- Per talk/document
- Mixed strategies (verse with context)

## Requirements

- LM Studio running on port 1234
- Qwen3-Embedding model loaded (4B or 8B Q4_K_M)

## LM Studio Setup

1. Start LM Studio
2. Go to Local Server tab
3. Load an embedding model (e.g., Qwen3-Embedding-4B-GGUF Q4_K_M)
4. Start server on port 1234

## Key Learnings

### Does retrieval need the embedding model?
**YES!** When you query the vector database:
1. Your query text must be converted to an embedding vector
2. That vector is compared against stored document vectors
3. The same embedding model must be used for consistency

The Go program doesn't contain the model weights - it calls the LM Studio API.

### Chunking Strategies

| Strategy | Pros | Cons |
|----------|------|------|
| Sentence | Fine-grained, precise matches | Loses context |
| Verse | Natural scripture unit | Short verses may lack context |
| Verse + Context | Balanced | More complex to implement |
| Chapter | Full context | Too broad for specific queries |
| Talk | Complete thoughts | Very broad matches |

## Running Experiments

```bash
go run . -experiment=basic     # Basic test with sample data
go run . -experiment=verse     # Test verse-level chunking
go run . -experiment=compare   # Compare different strategies
```
