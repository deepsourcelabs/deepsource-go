package analyzers

type Analyzer interface {
	String() string
	Run(interface{}) error
	RegisterRule(interface{})
}
