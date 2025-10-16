package golater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GeneratePreviewForTemplate(temp Template) string {
	var b strings.Builder
	endl := strings.ReplaceAll(temp.EndlineFormat, `\n`, "\n")

	b.WriteString(fmt.Sprintf("-> %s%s", temp.Repeated.Folder, endl))
	for i := 0; i < len(temp.Repeated.Files); i++ {
		prefix := "├"
		if i == len(temp.Repeated.Files)-1 {
			prefix = "└"
		}
		b.WriteString(fmt.Sprintf("%s── %s.%s%s", prefix, temp.Repeated.Files[i].Filename, temp.Repeated.Files[i].Ext, endl))
	}

	for i := 0; i < len(temp.Root); i++ {
		prefix := "├"
		if i == len(temp.Root)-1 {
			prefix = "└"
		}
		b.WriteString(fmt.Sprintf("%s── %s.%s%s", prefix, temp.Root[i].Filename, temp.Root[i].Ext, endl))
	}

	return previewBoxStyle.Render(b.String())
}

// ReadTemplates reads configuration file (golater.json) from constant path - home + "/.config/golater/golater.json"
// if this file doesn't exist, ReadTemplates creates new empty config and returns empty GolaterConfig
// all logs are pushed into msg_chan (1st parameter)
func ReadTemplates(msg_chan chan string) (GolaterConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return GolaterConfig{}, err
	}
	path := home + "/.config/golater/golater.json"

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(home+"/.config/golater", 0755); err != nil {
			return GolaterConfig{}, err
		}
		f, err := os.Create(path)
		if err != nil {
			return GolaterConfig{}, err
		}
		defer f.Close()
		msg_chan <- "created empty config:" + path
		return GolaterConfig{}, nil
	}
	defer file.Close()

	bytes, err := os.ReadFile(path)
	if err != nil {
		return GolaterConfig{}, err
	}
	var cfg GolaterConfig
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return GolaterConfig{}, err
	}

	if len(cfg.Template) > 0 {

		msg_chan <- fmt.Sprintf("Successfully loaded %d templates", len(cfg.Template))
	}
	return cfg, nil
}

func SpawnTemplate(t Template, iter int, ch chan string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	ch <- fmt.Sprintf("spawning template at %s with %d repeated parts", currentDir, iter)
	const perm os.FileMode = 0755

	for i := 0; i < iter; i++ {
		folder := strings.ReplaceAll(t.Repeated.Folder, "$n", fmt.Sprintf("%d", i+1))
		if err := os.MkdirAll(folder, perm); err != nil {
			return err
		}

		for _, file := range t.Repeated.Files {
			filename := strings.ReplaceAll(file.Filename, "$n", fmt.Sprintf("%d", i+1))
			filename = strings.TrimPrefix(filename, "/")

			path := filepath.Join(folder, fmt.Sprintf("%s.%s", filename, strings.TrimPrefix(file.Ext, ".")))
			if err := os.MkdirAll(filepath.Dir(path), perm); err != nil {
				return err
			}

			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if len(file.Data) > 0 {
				content := strings.Join(file.Data, t.EndlineFormat)
				if _, err := f.WriteString(content); err != nil {
					return nil
				}
			}

			ch <- "created file: " + path
		}
	}

	for _, file := range t.Root {
		filename := strings.TrimPrefix(file.Filename, "/")
		path := fmt.Sprintf("%s.%s", filename, strings.TrimPrefix(file.Ext, "."))

		if err := os.MkdirAll(filepath.Dir(path), perm); err != nil {
			return err
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if len(file.Data) > 0 {
			content := strings.Join(file.Data, t.EndlineFormat)
			if _, err := f.WriteString(content); err != nil {
				return err
			}
		}

		ch <- "created file: " + path
	}

	return nil
}
