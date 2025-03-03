// Code generated by gen.go; DO NOT EDIT.
package params

import (
	"fmt"

	"github.com/cfoust/cy/pkg/janet"
)

const (
	ParamAnimate       = "animate"
	ParamAnimations    = "animations"
	ParamDataDirectory = "data-directory"
	ParamDefaultFrame  = "default-frame"
	ParamDefaultShell  = "default-shell"
	ParamSkipInput     = "---skip-input"
)

func (p *Parameters) Animate() bool {
	value, ok := p.Get(ParamAnimate)
	if !ok {
		return defaults.Animate
	}

	realValue, ok := value.(bool)
	if !ok {
		return defaults.Animate
	}

	return realValue
}

func (p *Parameters) SetAnimate(value bool) {
	p.set(ParamAnimate, value)
}

func (p *Parameters) Animations() []string {
	value, ok := p.Get(ParamAnimations)
	if !ok {
		return defaults.Animations
	}

	realValue, ok := value.([]string)
	if !ok {
		return defaults.Animations
	}

	return realValue
}

func (p *Parameters) SetAnimations(value []string) {
	p.set(ParamAnimations, value)
}

func (p *Parameters) DataDirectory() string {
	value, ok := p.Get(ParamDataDirectory)
	if !ok {
		return defaults.DataDirectory
	}

	realValue, ok := value.(string)
	if !ok {
		return defaults.DataDirectory
	}

	return realValue
}

func (p *Parameters) SetDataDirectory(value string) {
	p.set(ParamDataDirectory, value)
}

func (p *Parameters) DefaultFrame() string {
	value, ok := p.Get(ParamDefaultFrame)
	if !ok {
		return defaults.DefaultFrame
	}

	realValue, ok := value.(string)
	if !ok {
		return defaults.DefaultFrame
	}

	return realValue
}

func (p *Parameters) SetDefaultFrame(value string) {
	p.set(ParamDefaultFrame, value)
}

func (p *Parameters) DefaultShell() string {
	value, ok := p.Get(ParamDefaultShell)
	if !ok {
		return defaults.DefaultShell
	}

	realValue, ok := value.(string)
	if !ok {
		return defaults.DefaultShell
	}

	return realValue
}

func (p *Parameters) SetDefaultShell(value string) {
	p.set(ParamDefaultShell, value)
}

func (p *Parameters) SkipInput() bool {
	value, ok := p.Get(ParamSkipInput)
	if !ok {
		return defaults.skipInput
	}

	realValue, ok := value.(bool)
	if !ok {
		return defaults.skipInput
	}

	return realValue
}

func (p *Parameters) SetSkipInput(value bool) {
	p.set(ParamSkipInput, value)
}

func (p *Parameters) isDefault(key string) bool {
	switch key {
	case ParamAnimate:
		return true
	case ParamAnimations:
		return true
	case ParamDataDirectory:
		return true
	case ParamDefaultFrame:
		return true
	case ParamDefaultShell:
		return true
	case ParamSkipInput:
		return true

	}
	return false
}

func (p *Parameters) setDefault(key string, value interface{}) error {
	janetValue, janetOk := value.(*janet.Value)
	switch key {
	case ParamAnimate:
		if !janetOk {
			realValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("invalid value for ParamAnimate, should be bool")
			}
			p.set(key, realValue)
			return nil
		}

		var translated bool
		err := janetValue.Unmarshal(&translated)
		if err != nil {
			janetValue.Free()
			return fmt.Errorf("invalid value for :animate: %s", err)
		}
		p.set(key, translated)
		return nil

	case ParamAnimations:
		if !janetOk {
			realValue, ok := value.([]string)
			if !ok {
				return fmt.Errorf("invalid value for ParamAnimations, should be []string")
			}
			p.set(key, realValue)
			return nil
		}

		var translated []string
		err := janetValue.Unmarshal(&translated)
		if err != nil {
			janetValue.Free()
			return fmt.Errorf("invalid value for :animations: %s", err)
		}
		p.set(key, translated)
		return nil

	case ParamDataDirectory:
		if !janetOk {
			realValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid value for ParamDataDirectory, should be string")
			}
			p.set(key, realValue)
			return nil
		}

		var translated string
		err := janetValue.Unmarshal(&translated)
		if err != nil {
			janetValue.Free()
			return fmt.Errorf("invalid value for :data-directory: %s", err)
		}
		p.set(key, translated)
		return nil

	case ParamDefaultFrame:
		if !janetOk {
			realValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid value for ParamDefaultFrame, should be string")
			}
			p.set(key, realValue)
			return nil
		}

		var translated string
		err := janetValue.Unmarshal(&translated)
		if err != nil {
			janetValue.Free()
			return fmt.Errorf("invalid value for :default-frame: %s", err)
		}
		p.set(key, translated)
		return nil

	case ParamDefaultShell:
		if !janetOk {
			realValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid value for ParamDefaultShell, should be string")
			}
			p.set(key, realValue)
			return nil
		}

		var translated string
		err := janetValue.Unmarshal(&translated)
		if err != nil {
			janetValue.Free()
			return fmt.Errorf("invalid value for :default-shell: %s", err)
		}
		p.set(key, translated)
		return nil

	case ParamSkipInput:
		if !janetOk {
			realValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("invalid value for ParamSkipInput, should be bool")
			}
			p.set(key, realValue)
			return nil
		}

		return fmt.Errorf(":---skip-input is a protected parameter")

	}
	return nil
}

func init() {
	_defaultParams = []DefaultParam{
		{
			Name:      "animate",
			Docstring: "Whether to enable animation.",
			Default:   defaults.Animate,
		},
		{
			Name:      "animations",
			Docstring: "A list of all of the enabled animations that will be used by\n(input/find). If this is an empty array, all built-in animations\nwill be enabled.",
			Default:   defaults.Animations,
		},
		{
			Name:      "data-directory",
			Docstring: "The directory in which .borg files will be saved. This is [inferred\non startup](replay-mode.md#recording-terminal-sessions-to-disk). If\nset to an empty string, recording to disk is disabled.",
			Default:   defaults.DataDirectory,
		},
		{
			Name:      "default-frame",
			Docstring: "The frame used for all new clients. A blank string means a random\nframe will be chosen from all frames.",
			Default:   defaults.DefaultFrame,
		},
		{
			Name:      "default-shell",
			Docstring: "The default shell with which to start panes. Defaults to the value\nof `$SHELL` on startup.",
			Default:   defaults.DefaultShell,
		},
	}
}
