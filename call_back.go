package zorm

type Callback struct {
	creates    []*func(scope *Scope)
	updates    []*func(scope *Scope)
	deletes    []*func(scope *Scope)
	rowQueries []*func(scope *Scope)
	processors []*Callback
}

type CallbackProcesser struct {
	name      string
	before    string
	after     string
	replace   string
	remove    string
	kind      string
	processor *func(scope *Scope)
	parent    *Callback
}
