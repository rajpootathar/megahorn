package platform

type Registry struct {
	platforms map[string]Platform
	order     []string
}

func NewRegistry() *Registry {
	return &Registry{
		platforms: make(map[string]Platform),
	}
}

func (r *Registry) Register(p Platform) {
	r.platforms[p.Name()] = p
	r.order = append(r.order, p.Name())
}

func (r *Registry) Get(name string) (Platform, bool) {
	p, ok := r.platforms[name]
	return p, ok
}

func (r *Registry) All() []Platform {
	result := make([]Platform, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.platforms[name])
	}
	return result
}

func (r *Registry) Authenticated() []Platform {
	var result []Platform
	for _, name := range r.order {
		if r.platforms[name].Status() == AuthStatusAuthenticated {
			result = append(result, r.platforms[name])
		}
	}
	return result
}
