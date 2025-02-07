package config

import (
	"github.com/containerd/platforms"
)

var defaultPlatform = ""

func DefaultPlatform() string {
	if defaultPlatform == "" {
		spec := platforms.DefaultSpec()
		if spec.Architecture == "arm" && spec.Variant == "v8" {
			spec.Variant = "v7"
		}
		return platforms.Format(platforms.Normalize(spec))
	}

	return defaultPlatform
}
