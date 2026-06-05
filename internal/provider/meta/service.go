package meta

type service struct {
	proxy metaUpstreamPort
}

func newService(proxy metaUpstreamPort) *service {
	return &service{proxy: proxy}
}
