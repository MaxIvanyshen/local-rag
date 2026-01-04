package chunker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixedSizeChunker_Chunk(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		chunkSize int
		expected  []ChunkResult
	}{
		{
			name:      "empty data",
			data:      []byte{},
			chunkSize: 5,
			expected:  nil,
		},
		{
			name:      "data smaller than chunk size",
			data:      []byte("abc"),
			chunkSize: 5,
			expected:  []ChunkResult{{Data: []byte("abc"), StartLine: 1, EndLine: 1}},
		},
		{
			name:      "exact multiple",
			data:      []byte("abcdef"),
			chunkSize: 3,
			expected: []ChunkResult{
				{Data: []byte("abc"), StartLine: 1, EndLine: 1},
				{Data: []byte("def"), StartLine: 1, EndLine: 1},
			},
		},
		{
			name:      "with remainder",
			data:      []byte("abcdefg"),
			chunkSize: 3,
			expected: []ChunkResult{
				{Data: []byte("abc"), StartLine: 1, EndLine: 1},
				{Data: []byte("def"), StartLine: 1, EndLine: 1},
				{Data: []byte("g"), StartLine: 1, EndLine: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := &FixedSizeChunker{ChunkSize: tt.chunkSize}
			result := chunker.Chunk(tt.data)
			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDelimiterChunker_Chunk(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		delimiter byte
		expected  []ChunkResult
	}{
		{
			name:      "empty data",
			data:      []byte{},
			delimiter: ',',
			expected:  nil,
		},
		{
			name:      "no delimiter",
			data:      []byte("abc"),
			delimiter: ',',
			expected:  []ChunkResult{{Data: []byte("abc"), StartLine: 1, EndLine: 1}},
		},
		{
			name:      "one delimiter",
			data:      []byte("a,b"),
			delimiter: ',',
			expected: []ChunkResult{
				{Data: []byte("a"), StartLine: 1, EndLine: 1},
				{Data: []byte("b"), StartLine: 1, EndLine: 1},
			},
		},
		{
			name:      "multiple delimiters",
			data:      []byte("a,b,c"),
			delimiter: ',',
			expected: []ChunkResult{
				{Data: []byte("a"), StartLine: 1, EndLine: 1},
				{Data: []byte("b"), StartLine: 1, EndLine: 1},
				{Data: []byte("c"), StartLine: 1, EndLine: 1},
			},
		},
		{
			name:      "leading and trailing delimiters",
			data:      []byte(",a,"),
			delimiter: ',',
			expected: []ChunkResult{
				{Data: []byte(""), StartLine: 1, EndLine: 1},
				{Data: []byte("a"), StartLine: 1, EndLine: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := &DelimiterChunker{Delimiter: tt.delimiter}
			result := chunker.Chunk(tt.data)
			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSentenceChunker_SetOverlap(t *testing.T) {
	chunker := &SentenceChunker{}
	chunker.SetOverlap(5)
	require.Equal(t, 5, chunker.OverlapSize)
}

func TestSentenceChunker_Chunk(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		overlapSize int
		expected    []ChunkResult
	}{
		{
			name:        "empty data",
			data:        []byte{},
			overlapSize: 0,
			expected:    nil,
		},
		{
			name:        "no sentence endings",
			data:        []byte("hello world"),
			overlapSize: 0,
			expected:    []ChunkResult{{Data: []byte("hello world"), StartLine: 1, EndLine: 1}},
		},
		{
			name:        "one sentence",
			data:        []byte("Hello."),
			overlapSize: 0,
			expected:    []ChunkResult{{Data: []byte("Hello."), StartLine: 1, EndLine: 1}},
		},
		{
			name:        "multiple sentences",
			data:        []byte("Hi. How are you?"),
			overlapSize: 0,
			expected: []ChunkResult{
				{Data: []byte("Hi."), StartLine: 1, EndLine: 1},
				{Data: []byte(" How are you?"), StartLine: 1, EndLine: 1},
			},
		},
		{
			name:        "with overlap",
			data:        []byte("Hi. Hello!"),
			overlapSize: 2,
			expected: []ChunkResult{
				{Data: []byte("Hi."), StartLine: 1, EndLine: 1},
				{Data: []byte("i. Hello!"), StartLine: 1, EndLine: 1},
				{Data: []byte("o!"), StartLine: 1, EndLine: 1},
			},
		},
		{
			name:        "overlap larger than chunk",
			data:        []byte("Hi. Hello!"),
			overlapSize: 10,
			expected: []ChunkResult{
				{Data: []byte("Hi."), StartLine: 1, EndLine: 1},
				{Data: []byte("Hi. Hello!"), StartLine: 1, EndLine: 1},
				{Data: []byte("Hi. Hello!"), StartLine: 1, EndLine: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := &SentenceChunker{}
			chunker.SetOverlap(tt.overlapSize)
			result := chunker.Chunk(tt.data)
			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test that SentenceChunker implements OverlapChunker interface
func TestSentenceChunker_ImplementsOverlapChunker(t *testing.T) {
	var _ OverlapChunker = &SentenceChunker{}
}

func TestParagraphChunker_SetOverlap(t *testing.T) {
	chunker := &ParagraphChunker{}
	chunker.SetOverlap(10)
	require.Equal(t, 10, chunker.OverlapSize)
}

func TestParagraphChunker_Chunk(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		overlapSize int
		expected    []ChunkResult
	}{
		{
			name:        "empty data",
			data:        []byte{},
			overlapSize: 0,
			expected:    nil,
		},
		{
			name:        "no paragraphs",
			data:        []byte("hello world"),
			overlapSize: 0,
			expected:    []ChunkResult{{Data: []byte("hello world"), StartLine: 1, EndLine: 1}},
		},
		{
			name:        "one paragraph",
			data:        []byte("This is a paragraph.\n\n"),
			overlapSize: 0,
			expected:    []ChunkResult{{Data: []byte("This is a paragraph.\n\n"), StartLine: 1, EndLine: 3}},
		},
		{
			name:        "multiple paragraphs",
			data:        []byte("First para.\n\nSecond para.\n\n"),
			overlapSize: 0,
			expected: []ChunkResult{
				{Data: []byte("First para.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("Second para.\n\n"), StartLine: 3, EndLine: 5},
			},
		},
		{
			name:        "with overlap",
			data:        []byte("Para one.\n\nPara two.\n\n"),
			overlapSize: 5,
			expected: []ChunkResult{
				{Data: []byte("Para one.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("ne.\n\nPara two.\n\n"), StartLine: 1, EndLine: 5},
				{Data: []byte("wo.\n\n"), StartLine: 3, EndLine: 5},
			},
		},
		{
			name:        "overlap larger than paragraph",
			data:        []byte("Short.\n\nLong paragraph here.\n\n"),
			overlapSize: 20,
			expected: []ChunkResult{
				{Data: []byte("Short.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("Short.\n\nLong paragraph here.\n\n"), StartLine: 1, EndLine: 5},
				{Data: []byte("ng paragraph here.\n\n"), StartLine: 3, EndLine: 5},
			},
		},
		{
			name:        "no trailing double newline",
			data:        []byte("First.\n\nSecond."),
			overlapSize: 0,
			expected: []ChunkResult{
				{Data: []byte("First.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("Second."), StartLine: 3, EndLine: 3},
			},
		},
		{
			name:        "multiple paragraphs",
			data:        []byte("First para.\n\nSecond para.\n\n"),
			overlapSize: 0,
			expected: []ChunkResult{
				{Data: []byte("First para.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("Second para.\n\n"), StartLine: 3, EndLine: 5},
			},
		},
		{
			name:        "with overlap",
			data:        []byte("Para one.\n\nPara two.\n\n"),
			overlapSize: 5,
			expected: []ChunkResult{
				{Data: []byte("Para one.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("ne.\n\nPara two.\n\n"), StartLine: 1, EndLine: 5},
				{Data: []byte("wo.\n\n"), StartLine: 3, EndLine: 5},
			},
		},
		{
			name:        "overlap larger than paragraph",
			data:        []byte("Short.\n\nLong paragraph here.\n\n"),
			overlapSize: 20,
			expected: []ChunkResult{
				{Data: []byte("Short.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("Short.\n\nLong paragraph here.\n\n"), StartLine: 1, EndLine: 5},
				{Data: []byte("ng paragraph here.\n\n"), StartLine: 3, EndLine: 5},
			},
		},
		{
			name:        "no trailing double newline",
			data:        []byte("First.\n\nSecond."),
			overlapSize: 0,
			expected: []ChunkResult{
				{Data: []byte("First.\n\n"), StartLine: 1, EndLine: 3},
				{Data: []byte("Second."), StartLine: 3, EndLine: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := &ParagraphChunker{}
			chunker.SetOverlap(tt.overlapSize)
			result := chunker.Chunk(tt.data)
			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test that ParagraphChunker implements OverlapChunker interface
func TestParagraphChunker_ImplementsOverlapChunker(t *testing.T) {
	var _ OverlapChunker = &ParagraphChunker{}
}

func TestSentenceChunker_WithMarkdownFile(t *testing.T) {
	data, err := os.ReadFile("../test_data/chunking.md")
	require.NoError(t, err)

	chunker := &SentenceChunker{}
	chunker.SetOverlap(0)
	chunks := chunker.Chunk(data)

	// The file contains 8 sentences ending with '.', plus remaining whitespace
	require.Len(t, chunks, 9)
}

func TestParagraphChunker_WithMarkdownFile(t *testing.T) {
	data, err := os.ReadFile("../test_data/chunking.md")
	require.NoError(t, err)

	chunker := &ParagraphChunker{}
	chunker.SetOverlap(0)
	chunks := chunker.Chunk(data)

	// The file has paragraphs separated by \n\n
	require.Len(t, chunks, 2)
}

func TestSentenceChunker_OverlapVerification(t *testing.T) {
	data := []byte("This is the first long sentence. This is the second long sentence. This is the third long sentence.")
	overlapSize := 5
	chunker := &SentenceChunker{}
	chunker.SetOverlap(overlapSize)
	chunkResults := chunker.Chunk(data)

	require.Len(t, chunkResults, 4)

	// Check overlap between consecutive chunks where possible
	for i := 0; i < len(chunkResults)-1; i++ {
		require.GreaterOrEqual(t, len(chunkResults[i].Data), overlapSize)
		overlapLen := overlapSize
		if overlapLen > len(chunkResults[i+1].Data) {
			overlapLen = len(chunkResults[i+1].Data)
		}
		expectedOverlap := chunkResults[i].Data[len(chunkResults[i].Data)-overlapLen:]
		actualOverlap := chunkResults[i+1].Data[:overlapLen]
		require.Equal(t, expectedOverlap, actualOverlap, "Overlap mismatch between chunk %d and %d", i, i+1)
	}
}
