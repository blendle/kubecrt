package helm

import (
	"fmt"
	"os"

	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

// Init makes sure the Helm home path exists and the required subfolders.
func Init() error {
	if err := ensureDirectories(settings.Home); err != nil {
		return err
	}

	if err := ensureRepoFileFormat(settings.Home.RepositoryFile()); err != nil {
		return err
	}

	return nil
}

func ensureDirectories(home helmpath.Home) error {
	configDirectories := []string{
		home.String(),
		home.Repository(),
		home.Cache(),
		home.Plugins(),
	}

	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	return nil
}

func ensureRepoFileFormat(file string) error {
	r, err := repo.LoadRepositoriesFile(file)
	if err == repo.ErrRepoOutOfDate {
		if err := r.WriteFile(file, 0644); err != nil {
			return err
		}
	}

	return nil
}
