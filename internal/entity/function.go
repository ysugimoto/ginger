package entity

// Functions is slice type for Function pointer structs.
type Functions []*Function

// Exists() returns bool which function is registered.
func (f Functions) Exists(name string) bool {
	for _, fn := range f {
		if fn.Name == name {
			return true
		}
	}
	return false
}

// Find() returns Function pointer is exists.
func (f Functions) Find(name string) *Function {
	for _, fn := range f {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

// Remove() remvoes function from application.
func (f Functions) Remove(name string) Functions {
	fns := Functions{}
	for _, fn := range f {
		if fn.Name == name {
			continue
		}
		fns = append(fns, fn)
	}
	return fns
}

// Function is the entity struct which maps from configuration.
type Function struct {
	Name       string `toml:"name"`
	Arn        string `toml:"arn"`
	MemorySize int64  `toml:"memory_size"`
	Timeout    int64  `toml:"timeout"`
}
