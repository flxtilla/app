package template

type Templates interface {
	Templatr
	SwapTemplatr(Templatr)
}

type templates struct {
	Templatr
}

func New(t Templatr) Templates {
	return &templates{
		Templatr: t,
	}
}

func (t *templates) SwapTemplatr(tr Templatr) {
	t.Templatr = tr
}
