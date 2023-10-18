package steamcmd

import (
	"context"
	"fmt"
)

func NewSession(ctx context.Context) (*Session, error) {
	ioSession, err := NewSessionIO(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session IO: %w", err)
	}

	return &Session{
		IO: ioSession,
	}, nil
}

type Session struct {
	IO *SessionIO
}

func (s *Session) LoginAnonymous() error {
	_, err := s.IO.Exec("login anonymous")
	if err != nil {
		return fmt.Errorf("failed to execute login command: %w", err)
	}

	return nil
}

func (s *Session) ForceInstallDir(installDir string) error {
	_, err := s.IO.Exec("force_install_dir " + installDir)
	if err != nil {
		return fmt.Errorf("failed to force install dir %s: %w", installDir, err)
	}

	return nil
}

func (s *Session) AppUpdate(appID int, validate bool) error {
	cmd := fmt.Sprintf("app_update %d", appID)
	if validate {
		cmd += " validate"
	}

	if _, err := s.IO.Exec(cmd); err != nil {
		return fmt.Errorf("failed to execute app update command: %w", err)
	}

	return nil
}

func (s *Session) Close() error {
	return s.IO.Close()
}
