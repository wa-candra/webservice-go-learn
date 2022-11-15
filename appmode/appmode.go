package appmode

type AppMode uint

const (
	Unkown AppMode = iota
	Development
	Production
	Testing
)
