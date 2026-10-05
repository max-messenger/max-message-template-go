package template

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

type Getter interface {
	// Get возвращает значение по ключу и флаг существования
	Get(string) (any, bool)
}

// Values is a plain map-based Getter implementation for handlers.
type Values map[string]any

// Get возвращает значение по ключу и флаг существования.
func (v Values) Get(key string) (any, bool) {
	val, ok := v[key]

	return val, ok
}

// ReplaceValues заменяет значения в шаблоне. Поддерживает формат {key}.
// Если ключ не найден, плейсхолдер оставляется без изменений.
func ReplaceValues(template string, values Getter) string {
	var b strings.Builder
	for i := 0; i < len(template); {
		if template[i] == '{' {
			end := strings.IndexByte(template[i+1:], '}')
			if end != -1 {
				key := template[i+1 : i+1+end]
				if val, ok := values.Get(key); ok {
					_, _ = fmt.Fprint(&b, val)
					i += 2 + end

					continue
				}
			}
		}
		b.WriteByte(template[i])
		i++
	}

	return b.String()
}

// SplitButtons разбивает строку с кнопками по запятым, удаляя пустые элементы.
// Возвращает список чистых названий кнопок.
func SplitButtons(s string) []string {
	var result []string
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// ParseButton парсит строку кнопки в формате [name](call).
// Возвращает (имя_кнопки, вызов) или ("", "") при некорректном формате.
func ParseButton(s string) (string, string) {
	open := strings.IndexByte(s, '[')
	closed := strings.IndexByte(s, ']')
	parenOpen := strings.IndexByte(s, '(')
	parenClose := strings.IndexByte(s, ')')
	if open == -1 || closed == -1 || parenOpen == -1 || parenClose == -1 {
		return "", ""
	}
	name := s[open+1 : closed]
	call := s[parenOpen+1 : parenClose]

	return name, call
}

// addButton добавляет кнопку в строку клавиатуры на основе типа вызова.
// Поддерживаются: cb:, http(s)://, geo:true|false, contact:, app:<bot_id>, msg:, clip:.
// Неизвестные форматы вызова игнорируются.
func addButton(row *model.KeyboardRow, name string, call string) { //nolint:cyclop
	switch {
	case strings.HasPrefix(call, "cb:"):
		row.AddCallBack(name, call[len("cb:"):])
	case strings.HasPrefix(call, "http://") || strings.HasPrefix(call, "https://"):
		row.AddLink(name, call)
	case call == "geo:true":
		row.AddGeoLocation(name, true)
	case call == "geo:false":
		row.AddGeoLocation(name, false)
	case strings.HasPrefix(call, "contact:"):
		row.AddContact(name)
	case strings.HasPrefix(call, "app:"):
		if botID, err := strconv.ParseInt(call[len("app:"):], 10, 64); err == nil {
			row.AddOpenApp(name, botID)
		}
	case strings.HasPrefix(call, "msg:"):
		row.AddMessage(call[len("msg:"):])
	case strings.HasPrefix(call, "clip:"):
		row.AddClipboard(name, call[len("clip:"):])
	}
}

// ParseForm парсит форму с телом и клавиатурой, разделёнными "---".
// Заменяет значения в теле и кнопках, создаёт модель клавиатуры.
// Возвращает (обработанное_тело, клавиатура).
func ParseForm(template string, values Getter) (string, *model.Keyboard) {
	template = ReplaceValues(template, values)

	var keyboard string
	parts := strings.Split(template, "---")
	body := parts[0]
	if len(parts) > 1 {
		keyboard = parts[1]
	}

	kbd := &model.Keyboard{}

	lines := strings.SplitSeq(keyboard, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		row := kbd.AddRow()
		buttons := strings.SplitSeq(line, ",")

		for btn := range buttons {
			btn = strings.TrimSpace(btn)
			name, payload := ParseButton(btn)
			if name == "" || payload == "" {
				continue
			}
			addButton(row, name, payload)
		}
	}

	return body, kbd
}

func ParseFormV2(template string, values Getter) (string, []model.Attachment) {
	template = ReplaceValues(template, values)
	var urls string
	parts := strings.Split(template, "===")
	other := parts[0]
	if len(parts) > 1 {
		urls = parts[0]
		other = parts[1]
	}

	attachments := make([]model.Attachment, 0, 10)

	for _, url := range strings.Split(urls, "\n") {
		if url == "" {
			continue
		}
		attachments = append(attachments, model.Attachment{
			Type: model.AttachImage,
			Payload: model.Payload{
				URL: url,
			},
		})
	}
	body, kbd := ParseForm(other, values)
	keyboard := kbd.Build()
	if len(keyboard.Payload.Buttons) > 0 {
		attachments = append(attachments, keyboard)
	}

	return body, attachments
}

func ParseTemplate(template string, values Getter) (string, []model.Attachment) {
	if strings.Index(template, "===") == -1 {
		s, kbd := ParseForm(template, values)
		if kbd != nil {
			return s, []model.Attachment{kbd.Build()}
		}

		return s, []model.Attachment{}
	}

	parts := strings.Split(template, "===")
	lines := strings.SplitSeq(parts[0], "\n")
	template = parts[1]
	attachments := make([]model.Attachment, 0, 10)
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if a := NewAttachment(line, values); a != nil {
			attachments = append(attachments, *a)
		}
	}
	s, attach := ParseFormV2(template, values)
	if len(attach) > 0 {
		attachments = append(attachments, attach...)
	}

	return s, attachments
}

//nolint:cyclop
func NewAttachment(line string, values Getter) *model.Attachment {
	s := model.Attachment{}
	switch {
	case strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://"):
		s.Payload.URL = line

		return &s

	case strings.HasPrefix(line, "image:"):
		filename := line[len("image:"):]
		if v, ok := values.Get(filename); ok {
			s.Type = model.AttachImage
			s.FileName = filename
			s.Payload.Token = v.(string)

			return &s
		}

	case strings.HasPrefix(line, "video:"):
		if v, ok := values.Get(line[len("video:"):]); ok {
			s.Type = model.AttachVideo
			s.Payload.Token = v.(string)

			return &s
		}

	case strings.HasPrefix(line, "audio:"):
		if v, ok := values.Get(line[len("audio:"):]); ok {
			s.Type = model.AttachAudio
			s.Payload.Token = v.(string)

			return &s
		}

	case strings.HasPrefix(line, "file:"):
		if v, ok := values.Get(line[len("file:"):]); ok {
			s.Type = model.AttachFile
			s.Payload.Token = v.(string)

			return &s
		}
	case strings.HasPrefix(line, "sticker:"):
		if v, ok := values.Get(line[len("sticker:"):]); ok {
			s.Type = model.AttachSticker
			s.Payload.Token = v.(string)

			return &s
		}
	}

	return nil
}
