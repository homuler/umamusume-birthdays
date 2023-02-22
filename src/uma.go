package umamusume

import (
	"fmt"
	"io"
	"sort"

	yaml "gopkg.in/yaml.v3"
)

type Uma struct {
	Name     string `yaml:"name"`
	Birthday string `yaml:"birthday"`
	Url      string `yaml:"url"`
	Playable bool   `yaml:"playable"`
	Costumes struct {
		School   string `yaml:"school"`
		Racing   string `yaml:"racing"`
		Original string `yaml:"original"`
		SF       string `yaml:"sf"`
	} `yaml:"costumes"`
	Variations []struct {
		Url string `yaml:"url"`
	} `yaml:"variations"`
}

func ReadYAML(r io.Reader) ([]*Uma, error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse characters: %w", err)
	}

	us := make([]*Uma, 0)
	if err := yaml.Unmarshal(bs, &us); err != nil {
		return nil, err
	}
	return us, nil
}

func Update(orig []*Uma, new []*Uma) []*Uma {
	res := orig
	for _, uma := range new {
		found := false

		for _, base := range orig {
			if base.Name == uma.Name {
				if base.Birthday == "" {
					// OCRの誕生日を訂正しうるように、上書きはしない
					base.Birthday = uma.Birthday
				}
				base.Costumes = uma.Costumes
				base.Playable = uma.Playable
				found = true
				break
			}
		}

		if !found {
			res = append(res, uma)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res
}
