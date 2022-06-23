package types

type Plugin interface {
	String() string
	BuildAST(string) (interface{}, error)
	Run(interface{}) error
}
