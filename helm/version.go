package helm

import (
	"fmt"

	"github.com/Masterminds/semver"

	"k8s.io/helm/cmd/helm/search"
	"k8s.io/helm/pkg/repo"
)

// GetAcceptableVersion accepts a SemVer constraint, and finds the best matching
// chart version.
func GetAcceptableVersion(name, constraint string) (string, error) {
	index, err := buildIndex()
	if err != nil {
		return "", err
	}

	var res []*search.Result
	res, err = index.Search(name, 10, false)
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", fmt.Errorf(`unable to find chart named "%s"`, name)
	}

	search.SortScore(res)

	if constraint != "" {
		res, err = applyConstraint(constraint, res)
		if err != nil {
			return "", err
		}
	}

	return res[0].Chart.Version, nil
}

func buildIndex() (*search.Index, error) {
	rf, err := repo.LoadRepositoriesFile(settings.Home.RepositoryFile())
	if err != nil {
		return nil, err
	}

	i := search.NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := settings.Home.CacheIndex(n)
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			continue
		}

		i.AddRepo(n, ind, true)
	}
	return i, nil
}

func applyConstraint(version string, res []*search.Result) ([]*search.Result, error) {
	constraint, err := semver.NewConstraint(version)
	if err != nil {
		return res, fmt.Errorf("invalid chart version/constraint format: %s", err)
	}

	data := res[:0]
	for _, r := range res {
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil || constraint.Check(v) {
			data = append(data, r)
		}
	}

	if len(data) == 0 {
		return data, fmt.Errorf("unable to fulfil chart version constraint %s", version)
	}

	return data, nil
}
