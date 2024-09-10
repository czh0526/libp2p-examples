package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/czh0526/libp2p-examples/pubsub/my-chat/common"
)

const (
	SettingFilename = "settings.json"
)

type Settings struct {
	Timestamp string   `json:"timestamp"`
	Bootstrap []string `json:"bootstrap"`
}

func getSettingsFile() (string, error) {
	return filepath.Join(HomeDir, "data", SettingFilename), nil
}

func LoadSettings() (*Settings, error) {
	// 获取文件
	settingsFilename, err := getSettingsFile()
	if err != nil {
		return nil, err
	}

	// 读文件
	content, err := os.ReadFile(settingsFilename)
	if err != nil {
		return nil, err
	}

	settings := Settings{
		Timestamp: common.GetTimestampString(time.Now()),
		Bootstrap: []string{},
	}
	if len(content) > 0 {
		err = json.Unmarshal(content, &settings)
		if err != nil {
			return nil, err
		}
	}

	return &settings, nil
}

func SaveSettings(settings *Settings) error {
	settingsFilename, err := getSettingsFile()
	if err != nil {
		return err
	}

	content, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return err
	}

	if err = os.WriteFile(settingsFilename, content, 0644); err != nil {
		return err
	}

	return nil
}
