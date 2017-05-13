package helm

import (
	"fmt"
	"os"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

const (
	stableRepository    = "stable"
	stableRepositoryURL = "https://kubernetes-charts.storage.googleapis.com"
)

// Init makes sure the Helm home path exists and the required subfolders.
func Init() error {
	if err := ensureDirectories(settings.Home); err != nil {
		return err
	}

	if err := ensureDefaultRepos(settings.Home); err != nil {
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
		home.Starters(),
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

func ensureDefaultRepos(home helmpath.Home) error {
	repoFile := home.RepositoryFile()
	if fi, err := os.Stat(repoFile); err != nil {
		f := repo.NewRepoFile()
		sr, err := initStableRepo(home.CacheIndex(stableRepository))
		if err != nil {
			return err
		}

		f.Add(sr)

		if err := f.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	} else if fi.IsDir() {
		return fmt.Errorf("%s must be a file, not a directory", repoFile)
	}
	return nil
}

func initStableRepo(cacheFile string) (*repo.Entry, error) {
	c := repo.Entry{
		Name:  stableRepository,
		URL:   stableRepositoryURL,
		Cache: cacheFile,
	}
	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return nil, err
	}

	// In this case, the cacheFile is always absolute. So passing empty string
	// is safe.
	if err := r.DownloadIndexFile(""); err != nil {
		return nil, fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", stableRepositoryURL, err.Error())
	}

	return &c, nil
}
