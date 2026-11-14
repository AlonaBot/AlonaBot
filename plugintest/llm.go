package plugintest

import "github.com/AlonaBot/AlonaBot/plugintest/internal/llmagent"

func (o *options) addLLM() {
	if o.llmHandler != nil {
		p := llmagent.New()
		llmagent.Handle(p, o.llmHandler)
		o.plugins = append(o.plugins, p)
	}
}
