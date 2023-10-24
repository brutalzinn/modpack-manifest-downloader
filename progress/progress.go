package progress

type Progress struct {
	Max       int
	Value     int
	Completed bool
	Text      string
}

func New(max int) *Progress {
	pg := Progress{
		Max:       max,
		Completed: false,
		Value:     0,
	}
	return &pg
}
func (pg *Progress) Done() {
	pg.Value++
}

func (pg *Progress) Complete() {
	pg.Completed = true
}

func (pg *Progress) SetText(text string) {
	pg.Text = text
}
