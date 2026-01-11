package rmqcontracts

type Queue struct {
	Name       string
	Durable    bool
	Autodelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]any
	Binding    *QueueBinding
}

type QueueBinding struct {
	ExchangeNmae string
	RoutingKey   string
	Args         map[string]any
	NoWait       bool
}

type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]any
}
