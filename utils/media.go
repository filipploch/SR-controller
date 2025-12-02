package utils

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GetMediaDuration zwraca długość pliku multimedialnego w sekundach używając ffprobe
func GetMediaDuration(filePath string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return int(duration * 1000), nil
}

// BuildPlaylistItem tworzy pojedynczy element playlisty VLC z pliku
func BuildPlaylistItem(mediaPath, relativeFilePath string) (map[string]interface{}, error) {
	absMediaPath, err := filepath.Abs(mediaPath)
	if err != nil {
		absMediaPath = mediaPath
	}

	fullPath := filepath.Join(absMediaPath, filepath.FromSlash(relativeFilePath))

	return map[string]interface{}{
		"value": fullPath,
	}, nil
}

// BuildPlaylist tworzy playlistę VLC z wielu plików
func BuildPlaylist(mediaPath string, relativeFilePaths []string) []map[string]interface{} {
	playlist := make([]map[string]interface{}, 0, len(relativeFilePaths))

	absMediaPath, err := filepath.Abs(mediaPath)
	if err != nil {
		absMediaPath = mediaPath
	}

	for _, relPath := range relativeFilePaths {
		if relPath != "" {
			fullPath := filepath.Join(absMediaPath, filepath.FromSlash(relPath))
			playlist = append(playlist, map[string]interface{}{
				"value": fullPath,
			})
		}
	}

	return playlist
}

// BuildVLCSettings tworzy standardowe ustawienia dla VLC Video Source
func BuildVLCSettings(playlist []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"playlist": playlist,
		"loop":     false,
		"shuffle":  false,
	}
}
