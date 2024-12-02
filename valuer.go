package cloudregistry

// ValueSetter is the interface that wraps the basic methods to set a value.
type ValueSetter interface {
	SetValue(string, any) error
}

// ValueSetterFunc is an adapter to allow the use of ordinary functions as ValueSetter.
type ValueSetterFunc func(string, any) error

func (f ValueSetterFunc) SetValue(key string, v any) error { return f(key, v) }
