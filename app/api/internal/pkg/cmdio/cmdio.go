package cmdio

import (
	"io"

	"github.com/manifoldco/promptui"
)

type IO struct {
	in  io.ReadCloser
	out io.WriteCloser
}

func New(in io.ReadCloser, out io.WriteCloser) *IO {
	return &IO{
		in:  in,
		out: out,
	}
}

// TODO customize prompt/select templates

func (i *IO) Prompt(label string, opts ...PromptOption) (string, error) {
	p := &promptui.Prompt{
		Label: label,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p.Run()
}

type PromptOption func(p *promptui.Prompt)

func WithMask(mask rune) PromptOption {
	return func(p *promptui.Prompt) {
		p.Mask = mask
	}
}

func (i *IO) Select(label string, items []string) (string, error) {
	s := &promptui.Select{
		Label: label,
		Items: items,
	}

	_, v, err := s.Run()
	return v, err
}
