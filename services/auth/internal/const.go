package internal

const (
	dbDriver = "postgres"

	authTokenAlgo = "HS256"

	requestParamUserID = "user_id"
)

const (
	HeaderAuth   = "Authorization"
	HeaderBearer = "Bearer "
)

type Role string

const (
	_ Role = "user"
	_ Role = "accountant"
	_ Role = "admin"
)

const (
	RabbitProtocol    = "amqp"
	RabbitDurable     = true
	RabbitAutoDelete  = false
	RabbitExclusive   = false
	RabbitNoWait      = false
	RabbitNoLocal     = false
	RabbitExchange    = "auth.out"
	RabbitMandatory   = false
	RabbitImmediate   = false
	RabbitContentType = "text/plain"
	RabbitConsumer    = ""
	RabbitAutoAck     = true
)

const userCreatedEventType = "user_created"
