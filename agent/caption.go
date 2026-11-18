package agent

import (
	"errors"
	"strings"

	"github.com/AlonaBot/AlonaBot/agent/provider"
	"github.com/AlonaBot/AlonaBot/sdk"
)

// captionImages 按原顺序输出描述；未读到的图片和转述失败的图片保留占位。
func (a *component) captionImages(c *sdk.Context, model *configuredModel, text string, images []provider.Image) string {
	var content strings.Builder
	for _, image := range images {
		if len(image.Data) == 0 {
			content.WriteString(imagePlaceholder + "\n")
			continue
		}
		if c.Context().Err() != nil {
			break
		}
		request := provider.Request{MaxTokens: model.maxTokens, Messages: []provider.Message{{Role: roleUser, Content: a.config.imageCaptionPrompt, Images: []provider.Image{image}}}}
		var description strings.Builder
		_, err := model.client.Stream(c.Context(), request, func(event provider.Event) error {
			description.WriteString(event.Text)
			return c.Context().Err()
		})
		if err == nil && strings.TrimSpace(description.String()) == "" {
			err = errors.New("provider returned no image description")
		}
		if err != nil {
			c.Logger().WarnContext(c.Context(), "image caption request failed", "model", model.name, "error", err)
			content.WriteString(imagePlaceholder + "\n")
			continue
		}
		content.WriteString("[image: " + description.String() + "]\n")
	}
	content.WriteString(text)
	return content.String()
}
