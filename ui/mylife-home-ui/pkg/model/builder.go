package model

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mylife-home-ui/pkg/web/api"
	"strings"
)

type builder struct {
	Resources            map[string]*Resource
	Model                *api.Model
	ModelHash            string
	resourcesTranslation map[string]string
	stylesHash           string
}

func newBuilder() *builder {
	return &builder{
		resourcesTranslation: make(map[string]string),
		Resources:            make(map[string]*Resource),
	}
}

func (b *builder) BuildModel(definition *Definition) error {
	for _, res := range definition.Resources {
		if err := b.createResource(&res); err != nil {
			return err
		}
	}

	{
		data := createCss(definition.Styles)
		b.stylesHash = b.setResource("text/css", []byte(data))
		logger.Infof("Creating CSS resource: hash='%s', size='%d'", b.stylesHash, len(data))
	}

	windows, err := b.translateWindows(definition.Windows)
	if err != nil {
		return fmt.Errorf("failed to translate windows: %w", err)
	}
	b.Model = &api.Model{
		Windows:       windows,
		DefaultWindow: definition.DefaultWindow,
		StyleHash:     b.stylesHash,
	}

	{
		data, err := json.Marshal(b.Model)
		if err != nil {
			return fmt.Errorf("failed to marshal model: %w", err)
		}
		b.ModelHash = b.setResource("application/json", data)
		logger.Infof("Creating resource from model: hash='%s', size='%d'", b.ModelHash, len(data))
	}

	return nil
}

func (b *builder) createResource(res *DefinitionResource) error {
	data, err := base64.StdEncoding.DecodeString(res.Data)
	if err != nil {
		return fmt.Errorf("failed to decode resource '%s': %w", res.ID, err)
	}

	hash := b.setResource(res.Mime, data)
	b.resourcesTranslation[res.ID] = hash
	logger.Infof("Creating resource from id '%s': hash='%s', size='%d'", res.ID, hash, len(data))

	return nil
}

func createCss(styles []DefinitionStyle) string {
	var cssRules []string

	for _, style := range styles {
		// Avoid collision with predefined styles ids
		className := fmt.Sprintf(".user-%s", style.ID)

		// Build properties
		properties := make([]string, 0, len(style.Properties))
		for key, value := range style.Properties {
			cssKey := formatCssKey(key)
			cssValue := formatCssValue(value)
			properties = append(properties, fmt.Sprintf("  %s: %s;", cssKey, cssValue))
		}

		// Build the complete CSS rule
		rule := fmt.Sprintf("%s {\n%s\n}", className, strings.Join(properties, "\n"))
		cssRules = append(cssRules, rule)
	}

	return strings.Join(cssRules, "\n\n")
}

func formatCssValue(value any) string {
	// Convert the property value to string
	switch v := value.(type) {
	case string:
		return v
	case float64:
		// Handle numeric values (JSON numbers are float64)
		i := int(v)
		if v == float64(i) {
			return fmt.Sprintf("%d", i)
		} else {
			return fmt.Sprintf("%f", v)
		}
	case bool:
		if v {
			return "true"
		} else {
			return "false"
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatCssKey(input string) string {
	// Convert camelCase to kebab-case for CSS property names
	var result strings.Builder

	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('-')
		}
		// Convert to lowercase using bitwise OR
		result.WriteRune(r | 32)
	}

	return result.String()
}

func (b *builder) setResource(mime string, data []byte) string {
	md5Sum := md5.Sum(data)
	hash := base64.URLEncoding.EncodeToString(md5Sum[:])

	b.Resources[hash] = &Resource{
		Mime: mime,
		Data: data,
	}

	return hash
}

func (b *builder) translateWindows(windows []Window) ([]api.Window, error) {
	list := make([]api.Window, 0, len(windows))

	for _, window := range windows {
		backgroundResource, err := b.translateResource(window.BackgroundResource)
		if err != nil {
			return nil, fmt.Errorf("failed to translate background resource for window '%s': %w", window.Id, err)
		}

		controls, err := b.translateControls(window.Controls)
		if err != nil {
			return nil, fmt.Errorf("failed to translate controls for window '%s': %w", window.Id, err)
		}

		list = append(list, api.Window{
			Id:                 window.Id,
			Style:              b.translateStyle(window.Style),
			Height:             window.Height,
			Width:              window.Width,
			BackgroundResource: backgroundResource,
			Controls:           controls,
		})
	}

	return list, nil
}

func (b *builder) translateControls(controls []Control) ([]api.Control, error) {
	list := make([]api.Control, 0, len(controls))

	for _, control := range controls {
		var display *api.ControlDisplay

		if control.Display != nil {
			defaultResource, err := b.translateResource(control.Display.DefaultResource)
			if err != nil {
				return nil, fmt.Errorf("failed to translate default resource for control '%s': %w", control.Id, err)
			}

			display = &api.ControlDisplay{
				ComponentId:     control.Display.ComponentId,
				ComponentState:  control.Display.ComponentState,
				Map:             make([]api.ControlDisplayMapItem, 0, len(control.Display.Map)),
				DefaultResource: defaultResource,
			}

			for index, item := range control.Display.Map {
				resource, err := b.translateResource(item.Resource)
				if err != nil {
					return nil, fmt.Errorf("failed to translate map resource for control '%s', item #%d: %w", control.Id, index, err)
				}

				display.Map = append(display.Map, api.ControlDisplayMapItem{
					Min:      item.Min,
					Max:      item.Max,
					Value:    item.Value,
					Resource: resource,
				})
			}
		}

		list = append(list, api.Control{
			Id:              control.Id,
			Style:           b.translateStyle(control.Style),
			Height:          control.Height,
			Width:           control.Width,
			X:               control.X,
			Y:               control.Y,
			Display:         display,
			Text:            control.Text,
			PrimaryAction:   control.PrimaryAction,
			SecondaryAction: control.SecondaryAction,
		})
	}

	return list, nil
}

func (b *builder) translateStyle(style Style) api.Style {
	list := make([]string, 0, len(style))

	for _, id := range style {
		list = append(list, "user-"+id)
	}

	return list
}

func (b *builder) translateResource(id string) (api.Resource, error) {
	if id == "" {
		return api.Resource(""), nil
	}

	hash, ok := b.resourcesTranslation[id]
	if !ok {
		return api.Resource(""), fmt.Errorf("resource not found: '%s'", id)
	}

	return api.Resource(hash), nil
}
