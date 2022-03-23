package ffprobe

type Codec struct {
	Name    string
	Tag     string
	Profile string
}

type Codecs []*Codec

func (c Codecs) Names() []string {
	var names []string
	for _, item := range c {
		names = append(names, item.Name)
	}
	return names
}
