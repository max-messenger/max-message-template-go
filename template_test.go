package template

import (
	_ "embed"
	"fmt"
	"reflect"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/require"
)

//go:embed example/hello.md
var tplHello string

func TestSimple(t *testing.T) {
	v := Values{}
	body, attachments := ParseForm(tplHello, v)

	fmt.Println(body, attachments)
}

func TestSimpleV2(t *testing.T) {
	v := Values{}
	body, attachments := ParseFormV2(tplHello, v)

	fmt.Println(body, attachments)
}

func TestBody(t *testing.T) {
	type args struct {
		template string
		values   Getter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"success",
			args{
				template: "**{foo}**",
				values:   values{"foo": "bar"},
			},
			"**bar**",
		},
		{
			"success1",
			args{
				template: "**{foo}**{foo1}**{foo}**",
				values:   values{"foo": "bar"},
			},
			"**bar**{foo1}**bar**",
		},
		{
			"success2",
			args{
				template: "**test**",
				values:   nil,
			},
			"**test**",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceValues(tt.args.template, tt.args.values); got != tt.want {
				t.Errorf("Body() = %v, want %v", got, tt.want)
			}
		})
	}
}

type values map[string]any

func (v values) Get(key string) (any, bool) {
	val, ok := v[key]
	return val, ok
}

func TestSplitButtons(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"success",
			args{"[name1](command1), [name2](command2),[name3](command3),   [name4](command4)"},
			[]string{"[name1](command1)", "[name2](command2)", "[name3](command3)", "[name4](command4)"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitButtons(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitButtons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseButton(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			"success",
			args{
				"[name](call)",
			},
			"name",
			"call",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseButton(tt.args.s)
			if got != tt.want {
				t.Errorf("ParseButton() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseButton() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParseForm(t *testing.T) {
	type args struct {
		template string
		values   Getter
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 model.Attachment
	}{
		{
			"success",
			args{
				`
user_id: {user_id}
---
[cb](cb:call), [cb with id](cb:call:{user_id})

[link](https://max.ru), [link1](http://max.ru)

[geo](geo:true), [geo1](geo:false)
[contact](contact:), [webapp](app:123456)
[message1](msg:message), [clipboard](clip:text)


`, values{"user_id": 1234},
			},
			"user_id: 1234",
			func() model.Attachment {
				kbd := &model.Keyboard{}

				row := kbd.AddRow()
				row.AddCallBack("cb", "call")
				row.AddCallBack("cb with id", "call:1234")

				row1 := kbd.AddRow()
				row1.AddLink("link", "https://max.ru")
				row1.AddLink("link1", "http://max.ru")

				row2 := kbd.AddRow()
				row2.AddGeoLocation("geo", true)
				row2.AddGeoLocation("geo1", false)

				row3 := kbd.AddRow()
				row3.AddContact("contact")
				row3.AddOpenApp("webapp", 123456)

				row4 := kbd.AddRow()
				row4.AddMessage("message")
				row4.AddClipboard("clipboard", "text")

				return kbd.Build()
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, kbd := ParseForm(tt.args.template, tt.args.values)
			if body != tt.want {
				t.Errorf("ParseForm() got = %v, want %v", body, tt.want)
			}
			att := kbd.Build()

			require.Equal(t, att.Type, tt.want1.Type)
			require.Equal(t, len(att.Payload.Buttons), len(tt.want1.Payload.Buttons))
			for i := range att.Payload.Buttons {
				gotRow := att.Payload.Buttons[i]
				wantRow := tt.want1.Payload.Buttons[i]
				require.Equal(t, len(gotRow), len(wantRow))
				for j := range att.Payload.Buttons[i] {
					require.Equal(t, gotRow[j], wantRow[j])
				}
			}

			require.Equal(t, att, tt.want1)
		})
	}
}
