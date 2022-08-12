package gont

type Option any

func Customize(opts []Option, extraOptions ...Option) []Option {
	new := []Option{}

	new = append(new, opts...)
	new = append(new, extraOptions...)

	return new
}
