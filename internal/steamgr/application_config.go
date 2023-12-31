package steamgr

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/SethCurry/steamgr/internal/steamcmd"
)

type ApplicationConfig struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Validate   bool     `json:"validate"`
	InstallDir string   `json:"install_dir"`
	Binary     string   `json:"binary"`
	Args       []string `json:"args"`
	Mods       []int    `json:"mods"`
}

func ApplyApplicationConfig(ctx context.Context, conf *ApplicationConfig, factory *steamcmd.SessionFactory) error {
	sess, err := factory.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new steamcmd session: %w", err)
	}
	defer sess.Close()

	if _, err := os.Stat(conf.InstallDir); os.IsNotExist(err) {
		err = os.MkdirAll(conf.InstallDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create install dir: %w", err)
		}
	}

	err = sess.ForceInstallDir(conf.InstallDir)
	if err != nil {
		return err
	}
	time.Sleep(time.Second)

	err = sess.Login()
	if err != nil {
		return err
	}
	time.Sleep(time.Second)

	err = sess.AppUpdate(conf.ID, conf.Validate)
	if err != nil {
		return err
	}
	time.Sleep(time.Second)

	for _, modID := range conf.Mods {
		err = sess.InstallMod(conf.ID, modID)
		if err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	return nil
}

func BuildSystemdUnitFile(conf *ApplicationConfig) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	builder := strings.Builder{}

	builder.WriteString("[Unit]\n")
	builder.WriteString("Description=" + conf.Name + "\n")
	builder.WriteString("After=networking.service\n")
	builder.WriteString("\n")
	builder.WriteString("[Service]\n")
	builder.WriteString("Type=simple\n")
	binaryPath := filepath.Join(conf.InstallDir, conf.Binary)
	builder.WriteString("ExecStart=" + binaryPath + " " + strings.Join(conf.Args, " ") + "\n")
	builder.WriteString("Restart=always\n")
	builder.WriteString("User=" + u.Username + "\n")
	builder.WriteString("Group=" + u.Username + "\n")
	builder.WriteString("WorkingDirectory=" + conf.InstallDir + "\n")
	builder.WriteString("\n")
	builder.WriteString("[Install]\n")
	builder.WriteString("WantedBy=default.target\n")
	builder.WriteString("\n")

	return builder.String(), nil
}
