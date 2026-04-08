package domain

import (
	"fmt"
	"time"
)

type BuildInfo struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildTime time.Time `json:"build_time"`
	GoVersion string    `json:"go_version"`
	Env       string    `json:"env"`
}

func NewBuildInfo(version, commit, buildTime, goVersion, env string) (BuildInfo, error) {
	if version == "" {
		return BuildInfo{}, fmt.Errorf("version is required")
	}

	// Handle "unknown" or empty buildTime for development builds
	var bt time.Time
	if buildTime == "unknown" || buildTime == "" {
		bt = time.Now()
	} else {
		var err error
		bt, err = time.Parse(time.RFC3339, buildTime)
		if err != nil {
			return BuildInfo{}, fmt.Errorf("parse build time: %w", err)
		}
	}

	return BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: bt,
		GoVersion: goVersion,
		Env:       env,
	}, nil
}
