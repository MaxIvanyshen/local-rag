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
		expected  [][]byte
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
			expected:  [][]byte{[]byte("abc")},
		},
		{
			name:      "exact multiple",
			data:      []byte("abcdef"),
			chunkSize: 3,
			expected:  [][]byte{[]byte("abc"), []byte("def")},
		},
		{
			name:      "with remainder",
			data:      []byte("abcdefg"),
			chunkSize: 3,
			expected:  [][]byte{[]byte("abc"), []byte("def"), []byte("g")},
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
		expected  [][]byte
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
			expected:  [][]byte{[]byte("abc")},
		},
		{
			name:      "one delimiter",
			data:      []byte("a,b"),
			delimiter: ',',
			expected:  [][]byte{[]byte("a"), []byte("b")},
		},
		{
			name:      "multiple delimiters",
			data:      []byte("a,b,c"),
			delimiter: ',',
			expected:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		},
		{
			name:      "leading and trailing delimiters",
			data:      []byte(",a,"),
			delimiter: ',',
			expected:  [][]byte{[]byte(""), []byte("a")},
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
		expected    [][]byte
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
			expected:    [][]byte{[]byte("hello world")},
		},
		{
			name:        "one sentence",
			data:        []byte("Hello."),
			overlapSize: 0,
			expected:    [][]byte{[]byte("Hello.")},
		},
		{
			name:        "multiple sentences",
			data:        []byte("Hi. How are you?"),
			overlapSize: 0,
			expected:    [][]byte{[]byte("Hi."), []byte(" How are you?")},
		},
		{
			name:        "with overlap",
			data:        []byte("Hi. Hello!"),
			overlapSize: 2,
			expected:    [][]byte{[]byte("Hi."), []byte("i. Hello!"), []byte("o!")},
		},
		{
			name:        "overlap larger than chunk",
			data:        []byte("Hi. Hello!"),
			overlapSize: 10,
			expected:    [][]byte{[]byte("Hi."), []byte("Hi. Hello!"), []byte("Hi. Hello!")},
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
		expected    [][]byte
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
			expected:    [][]byte{[]byte("hello world")},
		},
		{
			name:        "one paragraph",
			data:        []byte("This is a paragraph.\n\n"),
			overlapSize: 0,
			expected:    [][]byte{[]byte("This is a paragraph.\n\n")},
		},
		{
			name:        "multiple paragraphs",
			data:        []byte("First para.\n\nSecond para.\n\n"),
			overlapSize: 0,
			expected:    [][]byte{[]byte("First para.\n\n"), []byte("Second para.\n\n")},
		},
		{
			name:        "with overlap",
			data:        []byte("Para one.\n\nPara two.\n\n"),
			overlapSize: 5,
			expected:    [][]byte{[]byte("Para one.\n\n"), []byte("ne.\n\nPara two.\n\n"), []byte("wo.\n\n")},
		},
		{
			name:        "overlap larger than paragraph",
			data:        []byte("Short.\n\nLong paragraph here.\n\n"),
			overlapSize: 20,
			expected:    [][]byte{[]byte("Short.\n\n"), []byte("Short.\n\nLong paragraph here.\n\n"), []byte("ng paragraph here.\n\n")},
		},
		{
			name:        "no trailing double newline",
			data:        []byte("First.\n\nSecond."),
			overlapSize: 0,
			expected:    [][]byte{[]byte("First.\n\n"), []byte("Second.")},
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
	chunks := chunker.Chunk(data)

	require.Len(t, chunks, 4)

	// Check overlap between consecutive chunks where possible
	for i := 0; i < len(chunks)-1; i++ {
		require.GreaterOrEqual(t, len(chunks[i]), overlapSize)
		overlapLen := overlapSize
		if overlapLen > len(chunks[i+1]) {
			overlapLen = len(chunks[i+1])
		}
		expectedOverlap := chunks[i][len(chunks[i])-overlapLen:]
		actualOverlap := chunks[i+1][:overlapLen]
		require.Equal(t, expectedOverlap, actualOverlap, "Overlap mismatch between chunk %d and %d", i, i+1)
	}
}
