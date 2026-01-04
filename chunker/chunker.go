package chunker

type ChunkResult struct {
	Data      []byte
	StartLine int
	EndLine   int
}

type Chunker interface {
	Chunk(data []byte) []ChunkResult
}

// countLines counts the number of lines in the given data
func countLines(data []byte) int {
	count := 1
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}

// calculateLines calculates start and end line numbers for a chunk within the full data
func calculateLines(fullData []byte, chunkStart, chunkEnd int) (int, int) {
	startLine := 1
	for i := 0; i < chunkStart; i++ {
		if fullData[i] == '\n' {
			startLine++
		}
	}
	endLine := startLine + countLines(fullData[chunkStart:chunkEnd]) - 1
	return startLine, endLine
}

// FixedSizeChunker splits data into chunks of a fixed size.
type FixedSizeChunker struct {
	ChunkSize int
}

func (f *FixedSizeChunker) Chunk(data []byte) []ChunkResult {
	var chunks []ChunkResult
	for i := 0; i < len(data); i += f.ChunkSize {
		end := i + f.ChunkSize
		if end > len(data) {
			end = len(data)
		}
		startLine, endLine := calculateLines(data, i, end)
		chunks = append(chunks, ChunkResult{
			Data:      data[i:end],
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return chunks
}

// DelimiterChunker splits data based on a specified delimiter byte.
type DelimiterChunker struct {
	Delimiter byte
}

func (d *DelimiterChunker) Chunk(data []byte) []ChunkResult {
	var chunks []ChunkResult
	start := 0
	for i := range len(data) {
		if data[i] == d.Delimiter {
			startLine, endLine := calculateLines(data, start, i)
			chunks = append(chunks, ChunkResult{
				Data:      data[start:i],
				StartLine: startLine,
				EndLine:   endLine,
			})
			start = i + 1
		}
	}
	if start < len(data) {
		startLine, endLine := calculateLines(data, start, len(data))
		chunks = append(chunks, ChunkResult{
			Data:      data[start:],
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return chunks
}

type OverlapChunker interface {
	Chunk(data []byte) []ChunkResult
	SetOverlap(size int)
}

type SentenceChunker struct {
	OverlapSize int
}

func (s *SentenceChunker) SetOverlap(size int) {
	s.OverlapSize = size
}

func (s *SentenceChunker) Chunk(data []byte) []ChunkResult {
	var chunks []ChunkResult
	start := 0
	for i := range len(data) {
		if data[i] == '.' || data[i] == '!' || data[i] == '?' {
			end := i + 1
			if end > len(data) {
				end = len(data)
			}
			startLine, endLine := calculateLines(data, start, end)
			chunks = append(chunks, ChunkResult{
				Data:      data[start:end],
				StartLine: startLine,
				EndLine:   endLine,
			})
			start = end - s.OverlapSize
			if start < 0 {
				start = 0
			}
		}
	}
	if start < len(data) {
		startLine, endLine := calculateLines(data, start, len(data))
		chunks = append(chunks, ChunkResult{
			Data:      data[start:],
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return chunks
}

// ParagraphChunker splits data into paragraphs based on double newline characters.
type ParagraphChunker struct {
	OverlapSize int
}

func NewParagraphChunker(overlap int) *ParagraphChunker {
	return &ParagraphChunker{
		OverlapSize: overlap,
	}
}

func (p *ParagraphChunker) SetOverlap(size int) {
	p.OverlapSize = size
}

func (p *ParagraphChunker) Chunk(data []byte) []ChunkResult {
	var chunks []ChunkResult
	start := 0
	for i := range len(data) - 1 {
		if data[i] == '\n' && data[i+1] == '\n' {
			end := i + 2
			if end > len(data) {
				end = len(data)
			}
			startLine, endLine := calculateLines(data, start, end)
			chunks = append(chunks, ChunkResult{
				Data:      data[start:end],
				StartLine: startLine,
				EndLine:   endLine,
			})
			start = end - p.OverlapSize
			if start < 0 {
				start = 0
			}
		}
	}
	if start < len(data) {
		startLine, endLine := calculateLines(data, start, len(data))
		chunks = append(chunks, ChunkResult{
			Data:      data[start:],
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return chunks
}
