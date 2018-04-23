package pipe

import (
	"fmt"
	"os"
	"time"

	"github.com/laincloud/entry/server/models"
)

// SessionReplay is for session replay
type SessionReplay struct {
	timingFile     *os.File
	typescriptFile *os.File
	now            time.Time
}

// NewSessionReplay return an initialized *SessionReplay
func NewSessionReplay(s models.Session) (*SessionReplay, error) {
	typescriptFile, err := os.Create(s.TypescriptFile())
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(typescriptFile, "Script started on %s\n", time.Now())
	timingFile, err := os.Create(s.TimingFile())
	if err != nil {
		typescriptFile.Close()
		return nil, err
	}

	return &SessionReplay{
		timingFile:     timingFile,
		typescriptFile: typescriptFile,
		now:            time.Now(),
	}, nil
}

// Close close the underlying files
func (s *SessionReplay) Close() error {
	fmt.Fprintf(s.typescriptFile, "Script done on %s\n", time.Now())
	err1 := s.typescriptFile.Close()
	err2 := s.timingFile.Close()
	switch {
	case err1 != nil:
		return err1
	case err2 != nil:
		return err2
	default:
		return nil
	}
}

// record write down response and delay in respective files for future replay
func (s *SessionReplay) record(data []byte) {
	now := time.Now()
	delay := now.Sub(s.now)
	s.now = now
	s.typescriptFile.Write(data)
	fmt.Fprintf(s.timingFile, "%f %d\n", float64(delay)/1e9, len(data))
}
