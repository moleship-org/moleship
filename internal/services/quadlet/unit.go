package quadlet

type Unit interface {
	Name() string

	Kind() Kind

	Render() (string, error)

	Validate() error
}

func Filename(u Unit) string {
	return u.Name() + "." + u.Kind().Extension()
}

func ServiceName(u Unit) string {
	return u.Kind().ServiceName(u.Name())
}
