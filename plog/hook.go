package plog

type Hook interface {
	On(entry *Entry)
}

type Hooks map[LevelType][]Hook

func (hooks Hooks) HookOn(entry *Entry) {
	for _, hook := range hooks[entry.Level] {
		hook.On(entry)
	}
}
