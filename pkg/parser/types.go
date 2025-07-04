package parser

// DeltagramPart represents a single part of a deltagram
type DeltagramPart struct {
	ContentLocation string
	ContentType     string
	DeltaOperation  string
	Content         string
}

// Deltagram represents a complete deltagram with all its parts
type Deltagram struct {
	UUID  string
	Parts []DeltagramPart
}

// Parser defines the interface for parsing deltagrams
type Parser interface {
	Parse(content string) (*Deltagram, error)
}