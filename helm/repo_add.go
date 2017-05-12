package helm

import (
	"fmt"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/repo"
)

// AddRepository adds a new repository to the Helm index.
func AddRepository(name, url string) error {
	home := settings.Home

	f, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return err
	}

	cif := home.CacheIndex(name)
	c := repo.Entry{
		Name:     name,
		Cache:    cif,
		URL:      url,
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if err := r.DownloadIndexFile(home.Cache()); err != nil {
		return fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", url, err.Error())
	}

	f.Update(&c)

	return f.WriteFile(home.RepositoryFile(), 0644)
}
