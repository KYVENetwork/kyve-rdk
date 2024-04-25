package utils

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"os"
)

func GetJsonFromUrl(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	//goland:noinspection ALL
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	//goland:noinspection ALL
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func Logger() zerolog.Logger {
	writer := io.MultiWriter(os.Stdout)
	customConsoleWriter := zerolog.ConsoleWriter{Out: writer}
	customConsoleWriter.FormatCaller = func(i interface{}) string {
		return fmt.Sprintf("\x1b[36m[%s]\x1b[0m", "@kyvejs/tendermint-bsync")
	}

	logger := zerolog.New(customConsoleWriter).With().Timestamp().Logger()
	return logger
}
