package notebook

import (
	"encoding/json"
	"errors"
	"os"
)

type notebook struct {
	Cells         []cell                 `json:"cells"`
	Metadata      map[string]interface{} `json:"metadata"`
	NbFormat      int                    `json:"nbformat"`
	NbFormatMinor int                    `json:"nbformat_minor"`
}

type cell struct {
	CellType string                 `json:"cell_type"`
	Metadata map[string]interface{} `json:"metadata"`
	Source   []string               `json:"source"`
}

// AppendStep appends markdown and code cells to the notebook file.
func AppendStep(path string, title string, markdown string, code string) error {
	if path == "" {
		return errors.New("notebook path is required")
	}
	nb, err := loadOrCreate(path)
	if err != nil {
		return err
	}
	nb.Cells = append(nb.Cells, cell{
		CellType: "markdown",
		Metadata: map[string]interface{}{},
		Source:   []string{titleLine(title), "\n", markdown, "\n"},
	})
	if code != "" {
		nb.Cells = append(nb.Cells, cell{
			CellType: "code",
			Metadata: map[string]interface{}{},
			Source:   []string{code, "\n"},
		})
	}
	return writeNotebook(path, nb)
}

func loadOrCreate(path string) (*notebook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &notebook{
				Cells:         []cell{},
				Metadata:      map[string]interface{}{},
				NbFormat:      4,
				NbFormatMinor: 5,
			}, nil
		}
		return nil, err
	}
	var nb notebook
	if err := json.Unmarshal(data, &nb); err != nil {
		return nil, err
	}
	return &nb, nil
}

func writeNotebook(path string, nb *notebook) error {
	data, err := json.MarshalIndent(nb, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func titleLine(title string) string {
	if title == "" {
		return "## Step"
	}
	return "## " + title
}
