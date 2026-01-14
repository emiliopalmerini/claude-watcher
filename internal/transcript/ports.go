package transcript

import "io"

// Parser defines the interface for parsing Claude Code transcripts
type Parser interface {
	// Parse reads a transcript and extracts statistics and limit events
	Parse(transcriptPath string) (ParsedTranscript, error)

	// ParseReader parses from an io.Reader for flexibility and testing
	ParseReader(r io.Reader) (ParsedTranscript, error)
}
