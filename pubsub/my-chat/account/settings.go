package account

import (
	"encoding/json"
	"errors"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/common"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/config"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	SettingFilename = "settings.json"
)

type Settings struct {
	Timestamp string             `json:"timestamp"`
	Data      map[string]Account `json:"data"`
}

type Account struct {
	Id    string `json:"id"`
	Phone string `json:"phone"`
}

func getSettingsFile() (io.ReadWriteCloser, error) {
	settingFilepath := filepath.Join(config.HomeDir, "data/account", SettingFilename)
	return os.OpenFile(settingFilepath, os.O_RDWR|os.O_CREATE, 0666)
}

func loadSettings() (*Settings, error) {
	// 获取文件
	settingsFile, err := getSettingsFile()
	if err != nil {
		return nil, err
	}
	defer settingsFile.Close()

	// 读文件
	content, err := io.ReadAll(settingsFile)
	if err != nil {
		return nil, err
	}

	settings := Settings{
		Timestamp: common.GetTimestampString(time.Now()),
		Data:      make(map[string]Account),
	}
	if len(content) > 0 {
		err = json.Unmarshal(content, &settings)
		if err != nil {
			return nil, err
		}
	}

	return &settings, nil
}

func saveSettings(settings *Settings) error {
	settingsFile, err := getSettingsFile()
	if err != nil {
		return err
	}
	defer settingsFile.Close()

	data, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return err
	}

	written, err := settingsFile.Write(data)
	if err != nil {
		return err
	}
	if written != len(data) {
		return errors.New("failed to write all settings")
	}

	return nil
}
