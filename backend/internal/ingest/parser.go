package ingest

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"corelog/backend/internal/api"
	"corelog/backend/internal/dlp"
)

// Regex to identify if a line starts with a timestamp (space or T separator)
var tsPrefixRegex = regexp.MustCompile(`^\d{4}[-/]\d{2}[-/]\d{2}[T\s]\d{2}:\d{2}:\d{2}`)

// Generic log format: "2026-04-01 16:30:16 [INFO] my-app - Something happened"
// Also handles ISO separator: "2026-04-01T16:30:16Z [INFO] my-app - Something happened"
var logRegex = regexp.MustCompile(`^(\d{4}[-/]\d{2}[-/]\d{2})[T\s](\d{2}:\d{2}:\d{2})[^\s\[]*\s+\[(.*?)\]\s+(.*?)\s+-\s+(.*)$`)

// Docker/ISO format: "[2026-04-01T03:01:00.245Z] [INFO] message"
var dockerLogRegex = regexp.MustCompile(`^\[(\d{4}[-/]\d{2}[-/]\d{2})T(\d{2}:\d{2}:\d{2})[^\]]*\]\s+\[(.*?)\]\s+(.*)$`)

type pendingInfo struct {
	entry     *api.LogEntry
	updatedAt time.Time
}

type Parser struct {
	mu           sync.Mutex
	pending      map[string]*pendingInfo
	outChannel   chan api.LogEntry
	flushTimeout time.Duration
}

func NewParser(outChan chan api.LogEntry) *Parser {
	p := &Parser{
		pending:      make(map[string]*pendingInfo),
		outChannel:   outChan,
		flushTimeout: 500 * time.Millisecond,
	}
	go p.backgroundFlush()
	return p
}

func (p *Parser) backgroundFlush() {
	ticker := time.NewTicker(200 * time.Millisecond)
	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		for src, info := range p.pending {
			if now.Sub(info.updatedAt) >= p.flushTimeout {
				p.outChannel <- *info.entry
				delete(p.pending, src)
			}
		}
		p.mu.Unlock()
	}
}

// ParseLine ingests a raw string line from the file watcher
func (p *Parser) ParseLine(rawLine string) {
	p.ParseLineFromSource(rawLine, "unknown")
}

// ParseLineFromSource ingests a line associando-a a uma fonte (ex: nome do container Docker)
func (p *Parser) ParseLineFromSource(rawLine string, defaultSource string) {
	if strings.TrimSpace(rawLine) == "" {
		return
	}

	sanitized := dlp.Sanitize(rawLine)

	p.mu.Lock()
	defer p.mu.Unlock()

	info, exists := p.pending[defaultSource]

	if tsPrefixRegex.MatchString(sanitized) {
		// Formato padrão: "2026-04-01 16:30:16 [INFO] app - mensagem"
		if exists {
			p.outChannel <- *info.entry
		}
		matches := logRegex.FindStringSubmatch(sanitized)
		if len(matches) == 6 {
			p.pending[defaultSource] = &pendingInfo{
				entry: &api.LogEntry{
					Timestamp: matches[1] + " " + matches[2], // Normaliza para "YYYY-MM-DD HH:MM:SS"
					Level:     matches[3],
					SourceApp: matches[4],
					Message:   matches[5],
				},
				updatedAt: time.Now(),
			}
		} else {
			p.pending[defaultSource] = &pendingInfo{
				entry: &api.LogEntry{
					Timestamp: time.Now().UTC().Format("2006-01-02 15:04:05"),
					Level:     "INFO",
					SourceApp: defaultSource,
					Message:   sanitized,
				},
				updatedAt: time.Now(),
			}
		}
	} else if matches := dockerLogRegex.FindStringSubmatch(sanitized); len(matches) == 5 {
		// Formato Docker/ISO: "[2026-04-01T03:01:00Z] [INFO] mensagem"
		if exists {
			p.outChannel <- *info.entry
		}
		p.pending[defaultSource] = &pendingInfo{
			entry: &api.LogEntry{
				Timestamp: matches[1] + " " + matches[2],
				Level:     matches[3],
				SourceApp: defaultSource,
				Message:   matches[4],
			},
			updatedAt: time.Now(),
		}
	} else {
		// Stack trace ou linha multiline
		if exists {
			info.entry.Message += "\n" + sanitized
			info.updatedAt = time.Now()
		} else {
			p.pending[defaultSource] = &pendingInfo{
				entry: &api.LogEntry{
					Timestamp: time.Now().UTC().Format("2006-01-02 15:04:05"),
					Level:     "INFO",
					SourceApp: defaultSource,
					Message:   sanitized,
				},
				updatedAt: time.Now(),
			}
		}
	}
}

func (p *Parser) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for src, info := range p.pending {
		p.outChannel <- *info.entry
		delete(p.pending, src)
	}
}

