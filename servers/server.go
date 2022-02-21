package servers

type Server interface {
	Run(address string) error
}
