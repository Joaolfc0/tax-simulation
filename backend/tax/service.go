package tax

// Service is the Encore service struct for the tax bounded context.
// initService composes its dependencies; tests can build a Service literal
// with fake collaborators (e.g. Suppliers).
//
//encore:service
type Service struct {
	Suppliers SupplierResolver
}

func initService() (*Service, error) {
	return &Service{
		Suppliers: newSupplierResolver(secrets.InfoSimplesToken),
	}, nil
}
