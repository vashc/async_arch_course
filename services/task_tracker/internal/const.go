package internal

const (
	dbDriver = "postgres"

	requestParamUserID = "user_id"
	requestParamTaskID = "task_id"
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
	workerRole  Role = "worker"
	_           Role = "accountant"
	managerRole Role = "manager"
	adminRole   Role = "admin"
)

type TaskStatus string

const (
	createdStatus   TaskStatus = "created"
	completedStatus TaskStatus = "completed"
)

const (
	RabbitProtocol    = "amqp"
	RabbitDurable     = true
	RabbitAutoDelete  = false
	RabbitExclusive   = false
	RabbitNoWait      = false
	RabbitNoLocal     = false
	RabbitExchange    = "task_tracker.out"
	RabbitQueue       = "task_tracker.in"
	RabbitMandatory   = false
	RabbitImmediate   = false
	RabbitContentType = "text/plain"
	RabbitConsumer    = ""
	RabbitAutoAck     = true
)

type EventType string

const (
	userCreatedEventType   EventType = "user_created"
	taskCreatedEventType   EventType = "task_created"
	taskCompletedEventType EventType = "task_completed"
	taskAssignedEventType  EventType = "task_assigned"
)
