package main

// Artifacts contains docker images to be built and optionally publish them
type Artifacts struct {
	Registry string // default registry value
	Images   []ImageConfig
	Publish  []string // branch/tag's to publish images on
}

// GetImage return an image config of the name.  If the image is not found, nil
// is returned
func (art *Artifacts) GetImage(name string) *ImageConfig {
	for i, v := range art.Images {
		if v.Name == name {
			return &art.Images[i]
		}
	}
	return nil
}

// validateRegistry validates all the registry values by setting the default
// registry value if one is not specified along with setting the default
// docker file path
func (art *Artifacts) setDefaults() {
	for i, img := range art.Images {
		if len(img.Dockerfile) == 0 {
			art.Images[i].Dockerfile = "Dockerfile"
		}
		if len(img.Registry) == 0 && len(art.Registry) > 0 {
			art.Images[i].Registry = art.Registry
		}
	}
}
