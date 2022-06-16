package tube

type Tube struct {
	content  []error
	capacity int
}

func New(capacity int) *Tube {
	return &Tube{
		content:  []error{},
		capacity: capacity,
	}
}

func (t *Tube) Push(err error) {
	t.content = append(t.content, err)
	if len(t.content) > t.capacity {
		t.content = t.content[1:]
	}
}

func (t *Tube) Content() []error {
	return t.content
}
