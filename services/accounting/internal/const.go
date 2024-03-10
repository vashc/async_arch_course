package internal

const (
	dbDriver = "postgres"

	requestParamUserID = "user_id"
)

const (
	HeaderAuth   = "Authorization"
	HeaderBearer = "Bearer "
)

const (
	CtxAuthToken = "token"
)

type Role string

const (
	_              Role = "worker"
	accountantRole Role = "accountant"
	_              Role = "manager"
	adminRole      Role = "admin"
)

type TaskStatus string

const (
	RabbitProtocol    = "amqp"
	RabbitDurable     = true
	RabbitAutoDelete  = false
	RabbitExclusive   = false
	RabbitNoWait      = false
	RabbitNoLocal     = false
	RabbitExchange    = "accounting.out"
	RabbitQueue       = "accounting.in"
	RabbitMandatory   = false
	RabbitImmediate   = false
	RabbitContentType = "text/plain"
	RabbitConsumer    = ""
	RabbitAutoAck     = true
)

type EventType string

const (
	userCreatedEventType   EventType = "user_created"
	taskCompletedEventType EventType = "task_completed"
	taskAssignedEventType  EventType = "task_assigned"
)
