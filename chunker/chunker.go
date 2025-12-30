package chunker

type Chunker interface {
	Chunk(data []byte) [][]byte
}

// FixedSizeChunker splits data into chunks of a fixed size.
type FixedSizeChunker struct {
	ChunkSize int
}

func (f *FixedSizeChunker) Chunk(data []byte) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += f.ChunkSize {
		end := i + f.ChunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

// DelimiterChunker splits data based on a specified delimiter byte.
type DelimiterChunker struct {
	Delimiter byte
}

func (d *DelimiterChunker) Chunk(data []byte) [][]byte {
	var chunks [][]byte
	start := 0
	for i := range len(data) {
		if data[i] == d.Delimiter {
			chunks = append(chunks, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		chunks = append(chunks, data[start:])
	}
	return chunks
}

type OverlapChunker interface {
	Chunker
	SetOverlap(size int)
}

type SentenceChunker struct {
	OverlapSize int
}

func (s *SentenceChunker) SetOverlap(size int) {
	s.OverlapSize = size
}

func (s *SentenceChunker) Chunk(data []byte) [][]byte {
	var chunks [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '.' || data[i] == '!' || data[i] == '?' {
			end := i + 1
			if end > len(data) {
				end = len(data)
			}
			chunk := data[start:end]
			chunks = append(chunks, chunk)
			start = end - s.OverlapSize
			if start < 0 {
				start = 0
			}
		}
	}
	if start < len(data) {
		chunks = append(chunks, data[start:])
	}
	return chunks
}
